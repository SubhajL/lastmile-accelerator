import { describe, it, expect, vi } from 'vitest';
import type { PublishEvent } from './types.js';
import { createPublishConsumer } from './publish-consumer.js';

function evt(overrides: Partial<PublishEvent> = {}): PublishEvent {
  return {
    publishId: 'pb1',
    projectId: 'p1',
    tenantId: 't1',
    status: 'started',
    snapshotId: 's1',
    ...overrides
  };
}

describe('consumers/publish-consumer', () => {
  it('maps status to template and priority', async () => {
    const queue = { enqueue: vi.fn().mockResolvedValue('jid') };
    const c = createPublishConsumer({ queue: queue as any, defaultMaxAttempts: 3, resolveRecipients: vi.fn().mockResolvedValue(['u1']) });

    await c.handle(evt({ status: 'started' }));
    expect(queue.enqueue).toHaveBeenCalledWith(expect.objectContaining({ templateName: 'publish-started', priority: 'normal' }));

    await c.handle(evt({ status: 'healthy' }));
    expect(queue.enqueue).toHaveBeenCalledWith(expect.objectContaining({ templateName: 'publish-healthy', priority: 'normal' }));

    await c.handle(evt({ status: 'rolledback' }));
    expect(queue.enqueue).toHaveBeenCalledWith(expect.objectContaining({ templateName: 'publish-rolledback', priority: 'high' }));

    await c.handle(evt({ status: 'failed', error: 'x' }));
    expect(queue.enqueue).toHaveBeenCalledWith(expect.objectContaining({ templateName: 'publish-failed', priority: 'high' }));
  });
});
