import type { FixListCreatedEvent, FixesAppliedEvent, NotificationPriority } from './types.js';

export interface FixesConsumerOptions {
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
  resolveRecipients: (evt: FixListCreatedEvent | FixesAppliedEvent) => Promise<string[]>;
}

export function createFixesConsumer(opts: FixesConsumerOptions) {
  return {
    async handleCreated(evt: FixListCreatedEvent) {
      const users = await opts.resolveRecipients(evt);
      let count = 0;
      for (const userId of users) {
        await opts.queue.enqueue({
          tenantId: evt.tenantId,
          userId,
          channel: 'email',
          templateName: 'fixes-created',
          priority: 'normal',
          payload: { snapshotId: evt.snapshotId, projectId: evt.projectId, count: evt.count },
          maxAttempts: opts.defaultMaxAttempts
        });
        count += 1;
      }
      return { ok: true as const, enqueued: count };
    },

    async handleApplied(evt: FixesAppliedEvent) {
      const users = await opts.resolveRecipients(evt);
      let count = 0;
      for (const userId of users) {
        await opts.queue.enqueue({
          tenantId: evt.tenantId,
          userId,
          channel: 'email',
          templateName: 'fixes-applied',
          priority: 'high',
          payload: { snapshotId: evt.snapshotId, projectId: evt.projectId, fixIds: evt.fixIds, manifestUrl: evt.manifestUrl },
          maxAttempts: opts.defaultMaxAttempts
        });
        count += 1;
      }
      return { ok: true as const, enqueued: count };
    }
  };
}
