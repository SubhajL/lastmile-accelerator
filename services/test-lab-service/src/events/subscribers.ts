import type { NatsWrapper } from '../clients/nats.js';
import { SUBJECTS } from './contracts.js';
import type { RunnerOrchestrator } from '../services/runner.js';
import type { BrowserGridRunner } from '../services/browser-grid.js';

export async function subscribeRunRequests(nats: NatsWrapper, runner: RunnerOrchestrator) {
  await nats.subscribe(SUBJECTS.runRequest, async (evt: any) => {
    await runner.handleRunRequested(evt);
  });
}

export async function subscribeBrowserRequests(nats: NatsWrapper, grid: BrowserGridRunner) {
  await nats.subscribe(SUBJECTS.browserRequest, async (evt: any) => {
    await grid.handleBrowserRequested(evt);
  });
}

export async function wireEventSubscribers(args: { nats: NatsWrapper; runner: RunnerOrchestrator; grid: BrowserGridRunner }) {
  await subscribeRunRequests(args.nats, args.runner);
  await subscribeBrowserRequests(args.nats, args.grid);
}
