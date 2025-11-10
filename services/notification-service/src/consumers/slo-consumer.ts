import type { SLOBudgetExhaustedEvent, NotificationPriority } from './types.js';

export interface SloConsumerOptions {
  queue: {
    enqueue: (job: {
      tenantId: string;
      userId: string;
      channel: string;
      templateName: string;
      priority: NotificationPriority;
      payload: Record<string, unknown>;
      maxAttempts: number;
    }) => Promise<string>;
  };
  defaultMaxAttempts: number;
  resolveRecipients: (evt: SLOBudgetExhaustedEvent) => Promise<string[]>;
}

export function createSloConsumer(opts: SloConsumerOptions) {
  return {
    async handle(evt: SLOBudgetExhaustedEvent) {
      const users = await opts.resolveRecipients(evt);
      let count = 0;
      for (const userId of users) {
        await opts.queue.enqueue({
          tenantId: evt.tenantId,
          userId,
          channel: 'email',
          templateName: 'slo-budget-exhausted',
          priority: 'high',
          payload: { sloId: evt.sloId, projectId: evt.projectId, metric: evt.metric, budgetRemaining: evt.budgetRemaining },
          maxAttempts: opts.defaultMaxAttempts
        });
        count += 1;
      }
      return { ok: true as const, enqueued: count };
    }
  };
}
