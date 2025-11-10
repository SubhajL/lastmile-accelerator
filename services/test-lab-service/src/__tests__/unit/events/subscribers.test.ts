import { describe, it, expect, vi } from 'vitest';
import { wireEventSubscribers } from '../../../events/subscribers.js';
import { SUBJECTS } from '../../../events/contracts.js';

describe('event subscribers', () => {
  it('run.request and browser.request delegate to handlers', async () => {
    const handlers: Record<string, (d: any) => Promise<void>> = {};
    const nats = {
      subscribe: vi.fn().mockImplementation(async (subject: string, handler: (d: any) => Promise<void>) => {
        handlers[subject] = handler;
        return { subject } as any;
      }),
    } as any;

    const runner = { handleRunRequested: vi.fn().mockResolvedValue(undefined) } as any;
    const grid = { handleBrowserRequested: vi.fn().mockResolvedValue(undefined) } as any;

    await wireEventSubscribers({ nats, runner, grid });

    await handlers[SUBJECTS.runRequest]!({ runId: 'r', projectId: 'p' });
    await handlers[SUBJECTS.browserRequest]!({ browserRunId: 'br', testRunId: 'r', browser: 'chrome' });

    expect(runner.handleRunRequested).toHaveBeenCalled();
    expect(grid.handleBrowserRequested).toHaveBeenCalled();
  });
});
