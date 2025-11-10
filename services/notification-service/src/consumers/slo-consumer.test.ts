import { describe, it, expect, vi } from 'vitest';
import type { SLOBudgetExhaustedEvent } from './types.js';
import { createSloConsumer } from './slo-consumer.js';

function evt(overrides: Partial<SLOBudgetExhaustedEvent> = {}): SLOBudgetExhaustedEvent {
  return {
    sloId: 'slo1',
    projectId: 'p1',
    tenantId: 't1',
    metric: 'latency',
    budgetRemaining: 0,
    ...overrides
  };
}

describe('consumers/slo-consumer', () => {
  it('enqueues high priority budget exhausted notification', async () => {
    const queue = { enqueue: vi.fn().mockResolvedValue('jid') };
    const c = createSloConsumer({ queue: queue as any, defaultMaxAttempts: 3, resolveRecipients: vi.fn().mockResolvedValue(['u1']) });

    const res = await c.handle(evt());

    expect(res).toEqual({ ok: true, enqueued: 1 });
    expect(queue.enqueue).toHaveBeenCalledWith(expect.objectContaining({ templateName: 'slo-budget-exhausted', priority: 'high' }));
  });
});
