import type { S3Client } from '@aws-sdk/client-s3';
import type { NatsWrapper } from '../clients/nats.js';
import { putArtifact } from '../clients/s3.js';
import { SUBJECTS } from '../events/contracts.js';

export interface WebDriverLike {
  get(url: string): Promise<void>;
  takeScreenshot(): Promise<string>; // base64
  quit(): Promise<void>;
}

export interface BrowserTestRunsRepoLike {
  updateStatus(id: string, status: string, fields?: { startedAt?: Date; finishedAt?: Date; screenshots?: any; logs?: any }): Promise<any>;
}

export async function withRetry<T>(fn: () => Promise<T>, opts: { retries: number; backoffMs: number }): Promise<T> {
  let attempt = 0;
  let lastErr: any;
  while (attempt <= opts.retries) {
    try {
      return await fn();
    } catch (e) {
      lastErr = e;
      if (attempt === opts.retries) break;
      await new Promise((r) => setTimeout(r, opts.backoffMs));
      attempt++;
    }
  }
  throw lastErr;
}

export class BrowserGridRunner {
  constructor(
    private deps: {
      createDriver: (browser: string, gridUrl: string) => Promise<WebDriverLike>;
      gridUrl: string;
      s3: S3Client;
      bucketScreenshots: string;
      nats: NatsWrapper;
      browserRunsRepo: BrowserTestRunsRepoLike;
      testFlow: (driver: WebDriverLike) => Promise<void>;
      retry: { retries: number; backoffMs: number };
    }
  ) {}

  async handleBrowserRequested(req: { browserRunId: string; testRunId: string; browser: string }) {
    const { browserRunId, testRunId, browser } = req;
    await this.deps.browserRunsRepo.updateStatus(browserRunId, 'running', { startedAt: new Date() });
    await this.deps.nats.publish(SUBJECTS.browserStarted, {
      browserRunId,
      testRunId,
      status: 'running',
      startedAt: new Date().toISOString(),
    });

    let driver: WebDriverLike | null = null;
    try {
      driver = await this.deps.createDriver(browser, this.deps.gridUrl);
      await withRetry(() => this.deps.testFlow(driver!), this.deps.retry);

      const b64 = await driver.takeScreenshot();
      const buf = Buffer.from(b64, 'base64');
      const key = `screenshots/${testRunId}/${browserRunId}.png`;
      const uploaded = await putArtifact(this.deps.s3 as any, { bucket: this.deps.bucketScreenshots, key, body: buf, contentType: 'image/png' });

      await this.deps.browserRunsRepo.updateStatus(browserRunId, 'passed', { finishedAt: new Date(), screenshots: [uploaded] });
      await this.deps.nats.publish(SUBJECTS.browserFinished, {
        browserRunId,
        testRunId,
        status: 'passed',
        finishedAt: new Date().toISOString(),
        screenshots: [uploaded],
      });
    } catch (e) {
      await this.deps.browserRunsRepo.updateStatus(browserRunId, 'failed', { finishedAt: new Date(), logs: { error: (e as Error).message } });
      await this.deps.nats.publish(SUBJECTS.browserFinished, {
        browserRunId,
        testRunId,
        status: 'failed',
        finishedAt: new Date().toISOString(),
        logs: { error: (e as Error).message },
      });
    } finally {
      try { await driver?.quit(); } catch {}
    }
  }
}
