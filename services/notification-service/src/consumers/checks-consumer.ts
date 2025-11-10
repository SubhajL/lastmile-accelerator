import type { CheckFailedEvent, NotificationPriority } from './types.js';

export interface ChecksConsumerOptions {
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
  resolveRecipients: (evt: CheckFailedEvent) => Promise<string[]>;
}

function priorityFor(severity: CheckFailedEvent['severity']): NotificationPriority {
  switch (severity) {
    case 'low':
    case 'medium':
      return 'normal';
    case 'high':
    case 'critical':
      return 'high';
  }
}

export function createChecksConsumer(opts: ChecksConsumerOptions) {
  return {
    async handle(evt: CheckFailedEvent) {
      const users = await opts.resolveRecipients(evt);
      const prio = priorityFor(evt.severity);
      let count = 0;
      for (const userId of users) {
        await opts.queue.enqueue({
          tenantId: evt.tenantId,
          userId,
          channel: 'email',
          templateName: 'check-failed',
          priority: prio,
          payload: { checkId: evt.checkId, projectId: evt.projectId, type: evt.checkType, message: evt.message, severity: evt.severity },
          maxAttempts: opts.defaultMaxAttempts
        });
        count += 1;
      }
      return { ok: true as const, enqueued: count };
    }
  };
}
