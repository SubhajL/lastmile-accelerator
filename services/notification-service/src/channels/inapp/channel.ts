import type { NotificationJob } from '../../consumers/types.js';

export function createInAppChannel(opts: { publisher: { publish: (userId: string, payload: Record<string, unknown>) => Promise<string> }; resolveToUserId: (job: NotificationJob) => Promise<string> }) {
  return {
    async send(job: NotificationJob) {
      try {
        const userId = await opts.resolveToUserId(job);
        const title = job.templateName.replace(/[-_]/g, ' ');
        const body = JSON.stringify(job.payload);
        await opts.publisher.publish(userId, { title, body, data: job.payload });
        return { ok: true as const };
      } catch (e) {
        const msg = e instanceof Error ? e.message : String(e);
        return { ok: false as const, error: msg };
      }
    }
  };
}
