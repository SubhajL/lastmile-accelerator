import { describe, it, expect, vi } from 'vitest';
import { BrowserGridRunner, withRetry } from '../../../services/browser-grid.js';

vi.mock('../../../clients/s3.js', () => ({
  putArtifact: vi.fn().mockResolvedValue({ bucket: 'b', key: 'screenshots/tt/br.png' }),
}));

describe('Browser grid service', () => {
  it('withRetry retries then succeeds', async () => {
    const fn = vi.fn().mockRejectedValueOnce(new Error('x')).mockResolvedValueOnce('ok');
    const res = await withRetry(fn, { retries: 2, backoffMs: 1 });
    expect(res).toBe('ok');
    expect(fn).toHaveBeenCalledTimes(2);
  });

  it('success flow uploads screenshot, marks passed, publishes finished', async () => {
    const createDriver = vi.fn().mockResolvedValue({
      get: vi.fn(),
      takeScreenshot: vi.fn().mockResolvedValue(Buffer.from('img').toString('base64')),
      quit: vi.fn().mockResolvedValue(undefined),
    });
    const browserRunsRepo = { updateStatus: vi.fn().mockResolvedValue({}) } as any;
    const nats = { publish: vi.fn() } as any;

    const runner = new BrowserGridRunner({
      createDriver,
      gridUrl: 'http://grid',
      s3: {} as any,
      bucketScreenshots: 'b',
      nats,
      browserRunsRepo,
      testFlow: vi.fn().mockResolvedValue(undefined),
      retry: { retries: 0, backoffMs: 1 },
    });

    await runner.handleBrowserRequested({ browserRunId: 'br', testRunId: 'tt', browser: 'chrome' });

    expect(browserRunsRepo.updateStatus).toHaveBeenCalledWith('br', 'running', expect.any(Object));
    expect(browserRunsRepo.updateStatus).toHaveBeenCalledWith('br', 'passed', expect.any(Object));
    expect(nats.publish).toHaveBeenCalled();
  });

  it('failure after retries marks failed', async () => {
    const createDriver = vi.fn().mockResolvedValue({
      get: vi.fn(),
      takeScreenshot: vi.fn().mockResolvedValue(Buffer.from('img').toString('base64')),
      quit: vi.fn().mockResolvedValue(undefined),
    });
    const browserRunsRepo = { updateStatus: vi.fn().mockResolvedValue({}) } as any;
    const nats = { publish: vi.fn() } as any;

    const runner = new BrowserGridRunner({
      createDriver,
      gridUrl: 'http://grid',
      s3: {} as any,
      bucketScreenshots: 'b',
      nats,
      browserRunsRepo,
      testFlow: vi.fn().mockRejectedValue(new Error('fail!')),
      retry: { retries: 1, backoffMs: 1 },
    });

    await runner.handleBrowserRequested({ browserRunId: 'br', testRunId: 'tt', browser: 'chrome' });

    expect(browserRunsRepo.updateStatus).toHaveBeenCalledWith('br', 'failed', expect.any(Object));
  });
});
