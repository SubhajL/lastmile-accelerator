import type { PublishEvent, NotificationPriority } from './types.js';

export interface PublishConsumerOptions {
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
  resolveRecipients: (evt: PublishEvent) => Promise<string[]>;
}

function deriveTemplateAndPriority(status: PublishEvent['status']): { template: string; priority: NotificationPriority } {
  switch (status) {
    case 'started':
      return { template: 'publish-started', priority: 'normal' };
    case 'healthy':
      return { template: 'publish-healthy', priority: 'normal' };
    case 'rolledback':
      return { template: 'publish-rolledback', priority: 'high' };
    case 'failed':
      return { template: 'publish-failed', priority: 'high' };
  }
}

export function createPublishConsumer(opts: PublishConsumerOptions) {
  return {
    async handle(evt: PublishEvent) {
      const users = await opts.resolveRecipients(evt);
      const { template, priority } = deriveTemplateAndPriority(evt.status);
      let count = 0;
      for (const userId of users) {
        await opts.queue.enqueue({
          tenantId: evt.tenantId,
          userId,
          channel: 'email',
          templateName: template,
          priority,
          payload: { publishId: evt.publishId, projectId: evt.projectId, snapshotId: evt.snapshotId, status: evt.status, error: evt.error },
          maxAttempts: opts.defaultMaxAttempts
        });
        count += 1;
      }
      return { ok: true as const, enqueued: count };
    }
  };
}
