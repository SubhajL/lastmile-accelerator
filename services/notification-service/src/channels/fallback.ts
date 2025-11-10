import type { NotificationJob } from '../consumers/types.js';
import type { ChannelAdapter } from '../notifications/dispatcher.js';

export function createFallbackChannel(opts: { primary: ChannelAdapter; fallback: ChannelAdapter }): ChannelAdapter {
  return {
    async send(job: NotificationJob) {
      const first = await opts.primary.send(job);
      if (first.ok) return first;
      const second = await opts.fallback.send(job);
      if (second.ok) return second;
      return { ok: false, error: second.error } as const;
    }
  };
}
