import type { NotificationJob } from '../../consumers/types.js';
import type { ChannelAdapter } from '../../notifications/dispatcher.js';

export interface Limiter { tryRemoveTokens: (n: number) => boolean | Promise<boolean>; }

export function createRateLimitedChannel(opts: { inner: ChannelAdapter; limiter: Limiter; onLimited?: (job: NotificationJob) => Promise<{ ok: true } | { ok: false; error: string }> }) : ChannelAdapter {
  return {
    async send(job) {
      const allowed = await opts.limiter.tryRemoveTokens(1);
      if (!allowed) {
        if (opts.onLimited) return opts.onLimited(job);
        return { ok: false as const, error: 'rate_limited' };
      }
      return opts.inner.send(job);
    }
  };
}
