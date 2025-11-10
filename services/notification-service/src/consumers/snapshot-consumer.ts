import type { SnapshotReadyEvent } from './types.js';
import type { NotificationPriority } from './types.js';

export interface SnapshotConsumerOptions {
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
  resolveRecipients: (event: SnapshotReadyEvent) => Promise<string[]>;
}

export function createSnapshotConsumer(opts: SnapshotConsumerOptions) {
  return {
    async handle(evt: SnapshotReadyEvent): Promise<{ ok: true; enqueued: number } | { ok: false; error: string }> {
      try {
        const users = await opts.resolveRecipients(evt);
        let count = 0;
        for (const userId of users) {
          await opts.queue.enqueue({
            tenantId: evt.tenantId,
            userId,
            channel: 'email',
            templateName: 'snapshot-ready',
            priority: 'normal',
            payload: { snapshotId: evt.snapshotId, projectId: evt.projectId },
            maxAttempts: opts.defaultMaxAttempts
          });
          count += 1;
        }
        return { ok: true, enqueued: count };
      } catch (err) {
        const msg = err instanceof Error ? err.message : String(err);
        return { ok: false, error: msg };
      }
    }
  };
}
