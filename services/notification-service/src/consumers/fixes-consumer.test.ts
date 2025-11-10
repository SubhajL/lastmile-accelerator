import { describe, it, expect, vi } from 'vitest';
import type { FixListCreatedEvent, FixesAppliedEvent } from './types.js';
import { createFixesConsumer } from './fixes-consumer.js';

function created(overrides: Partial<FixListCreatedEvent> = {}): FixListCreatedEvent {
  return {
    snapshotId: 's1',
    projectId: 'p1',
    tenantId: 't1',
    count: 3,
    tierBreakdown: { tier1: 1, tier2: 1, tier3: 1 },
    ...overrides
  };
}

function applied(overrides: Partial<FixesAppliedEvent> = {}): FixesAppliedEvent {
  return {
    snapshotId: 's1',
    projectId: 'p1',
    tenantId: 't1',
    fixIds: ['f1','f2'],
    manifestUrl: 'http://example.com/manifest',
    ...overrides
  };
}

describe('consumers/fixes-consumer', () => {
  it('enqueues notification for fixes.created', async () => {
    const queue = { enqueue: vi.fn().mockResolvedValue('jid-1') };
    const c = createFixesConsumer({ queue: queue as any, defaultMaxAttempts: 3, resolveRecipients: vi.fn().mockResolvedValue(['u1']) });

    const res = await c.handleCreated(created());

    expect(res).toEqual({ ok: true, enqueued: 1 });
    expect(queue.enqueue).toHaveBeenCalledWith({
      tenantId: 't1',
      userId: 'u1',
      channel: 'email',
      templateName: 'fixes-created',
      priority: 'normal',
      payload: { snapshotId: 's1', projectId: 'p1', count: 3 },
      maxAttempts: 3
    });
  });

  it('enqueues high priority for fixes.applied', async () => {
    const queue = { enqueue: vi.fn().mockResolvedValue('jid-2') };
    const c = createFixesConsumer({ queue: queue as any, defaultMaxAttempts: 3, resolveRecipients: vi.fn().mockResolvedValue(['u2']) });

    const res = await c.handleApplied(applied());

    expect(res).toEqual({ ok: true, enqueued: 1 });
    expect(queue.enqueue).toHaveBeenCalledWith(expect.objectContaining({
      templateName: 'fixes-applied',
      priority: 'high'
    }));
  });
});
