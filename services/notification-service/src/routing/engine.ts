import type { NotificationJob } from '../consumers/types.js';
import type { UserPreferences } from '../prefs/store.js';

export interface RoutingEngineOptions {
  prefs: { get: (userId: string) => Promise<UserPreferences> };
  now: () => number;
}

export type RoutingDecision =
  | { status: 'allow' }
  | { status: 'block'; reason: string }
  | { status: 'defer'; reason: string };

export function createRoutingEngine(opts: RoutingEngineOptions) {
  function parseTime(str: string): number {
    const [h, m] = str.split(':').map((s) => parseInt(s, 10));
    return h * 60 + (m || 0);
    }

  function minutesSinceMidnight(epoch: number): number {
    const d = new Date(epoch);
    return d.getHours() * 60 + d.getMinutes();
  }

  function withinQuietHours(qh: { start: string; end: string }, nowEpoch: number): boolean {
    const start = parseTime(qh.start);
    const end = parseTime(qh.end);
    const nowMin = minutesSinceMidnight(nowEpoch);
    if (start === end) return false; // disabled or zero-length
    if (start < end) {
      // Same-day window
      return nowMin >= start && nowMin < end;
    }
    // Overnight window (e.g., 22:00-07:00)
    return nowMin >= start || nowMin < end;
  }

  return {
    async evaluate(job: NotificationJob): Promise<RoutingDecision> {
      const prefs = await opts.prefs.get(job.userId);

      // Channel toggle
      if (prefs.channels && prefs.channels[job.channel] === false) {
        return { status: 'block', reason: 'channel_disabled' };
      }

      // Frequency
      switch (prefs.frequency) {
        case 'mute':
          return { status: 'block', reason: 'muted' };
        case 'digest_hourly':
          return { status: 'defer', reason: 'digest_hourly' };
        case 'digest_daily':
          return { status: 'defer', reason: 'digest_daily' };
      }

      // Quiet hours
      if (prefs.quietHours && withinQuietHours(prefs.quietHours, opts.now())) {
        return { status: 'defer', reason: 'quiet_hours' };
      }

      return { status: 'allow' };
    }
  };
}
