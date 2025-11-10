import { describe, it, expect, vi } from 'vitest';
import type { SnapshotReadyEvent } from './types.js';
import { createSnapshotConsumer } from './snapshot-consumer.js';

function event(overrides: Partial<SnapshotReadyEvent> = {}): SnapshotReadyEvent {
  return {
    snapshotId: 'snap-1',
    projectId: 'proj-1',
    tenantId: 'tenant-1',
    checksCompleted: 10,
    issuesFound: 2,
    ...overrides
  };
}

describe('consumers/snapshot-consumer', () => {
  it('handle enqueues snapshot-ready notification', async () => {
    const queue = { enqueue: vi.fn().mockResolvedValue('jid-1') };
    const consumer = createSnapshotConsumer({
      queue: queue as any,
      defaultMaxAttempts: 4,
      resolveRecipients: vi.fn().mockResolvedValue(['user-1'])
    });

    const res = await consumer.handle(event());

    expect(res).toEqual({ ok: true, enqueued: 1 });
    expect(queue.enqueue).toHaveBeenCalledWith({
      tenantId: 'tenant-1',
      userId: 'user-1',
      channel: 'email',
      templateName: 'snapshot-ready',
      priority: 'normal',
      payload: { snapshotId: 'snap-1', projectId: 'proj-1' },
      maxAttempts: 4
    });
  });
});
