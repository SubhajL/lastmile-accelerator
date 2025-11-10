import { describe, it, expect, vi } from 'vitest';
import { publishRunStarted, publishRunFinished, publishBrowserStarted, publishBrowserFinished } from '../../../events/publishers.js';
import { SUBJECTS } from '../../../events/contracts.js';

describe('event publishers', () => {
  const nats = { publish: vi.fn() } as any;

  it('run.started validates and publishes', async () => {
    const payload = { runId: '11111111-1111-1111-1111-111111111111', projectId: '11111111-1111-1111-1111-111111111111', status: 'running', startedAt: '2024-01-01T00:00:00.000Z' };
    await publishRunStarted(nats, payload);
    expect(nats.publish).toHaveBeenCalledWith(SUBJECTS.runStarted, payload);
  });

  it('run.finished publishes with artifacts', async () => {
    const payload = { runId: '11111111-1111-1111-1111-111111111111', projectId: '11111111-1111-1111-1111-111111111111', status: 'passed', finishedAt: '2024-01-01T00:05:00.000Z', artifacts: [{ bucket: 'b', key: 'k' }] };
    await publishRunFinished(nats, payload);
    expect(nats.publish).toHaveBeenCalledWith(SUBJECTS.runFinished, payload);
  });

  it('browser.started validates and publishes', async () => {
    const payload = { browserRunId: '33333333-3333-3333-3333-333333333333', testRunId: '22222222-2222-2222-2222-222222222222', status: 'running', startedAt: '2024-01-01T00:01:00.000Z' };
    await publishBrowserStarted(nats, payload);
    expect(nats.publish).toHaveBeenCalledWith(SUBJECTS.browserStarted, payload);
  });

  it('browser.finished publishes with screenshots/logs', async () => {
    const payload = { browserRunId: '33333333-3333-3333-3333-333333333333', testRunId: '22222222-2222-2222-2222-222222222222', status: 'failed', finishedAt: '2024-01-01T00:03:00.000Z', screenshots: [{ bucket: 'b', key: 's' }], logs: { lines: 10 } };
    await publishBrowserFinished(nats, payload);
    expect(nats.publish).toHaveBeenCalledWith(SUBJECTS.browserFinished, payload);
  });

  it('invalid payload throws', async () => {
    await expect(publishRunStarted(nats, { foo: 'bar' } as any)).rejects.toBeTruthy();
  });
});
