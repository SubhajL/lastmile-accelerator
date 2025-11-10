import type { NotificationJob } from '../../consumers/types.js';
import type { ChannelAdapter } from '../../notifications/dispatcher.js';

export function createSmsFallbackToEmail(opts: { sms: ChannelAdapter; email: ChannelAdapter }): ChannelAdapter {
  return {
    async send(job: NotificationJob) {
      const r = await opts.sms.send(job);
      if (r.ok) return r;
      return opts.email.send(job);
    }
  };
}
