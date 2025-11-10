import type { CriticalErrorEvent, NotificationPriority } from './types.js';

export interface ErrorsConsumerOptions {
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
  resolveRecipients: (evt: CriticalErrorEvent) => Promise<string[]>;
}

export function createErrorsConsumer(opts: ErrorsConsumerOptions) {
  return {
    async handle(evt: CriticalErrorEvent) {
      const users = await opts.resolveRecipients(evt);
      let count = 0;
      for (const userId of users) {
        await opts.queue.enqueue({
          tenantId: evt.tenantId,
          userId,
          channel: 'email',
          templateName: 'critical-error',
          priority: 'critical',
          payload: { errorId: evt.errorId, projectId: evt.projectId, service: evt.service, message: evt.message, stackTrace: evt.stackTrace },
          maxAttempts: opts.defaultMaxAttempts
        });
        count += 1;
      }
      return { ok: true as const, enqueued: count };
    }
  };
}
