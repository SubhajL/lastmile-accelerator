export type DigestFrequency = 'immediate' | 'digest_hourly' | 'digest_daily' | 'mute';

export interface QuietHours {
  start: string; // 'HH:MM'
  end: string;   // 'HH:MM'
  timezone?: string; // optional IANA tz (not interpreted here)
}

export interface UserPreferences {
  channels?: Record<string, boolean>; // channel -> enabled; default true
  quietHours?: QuietHours;
  frequency?: DigestFrequency;
}

export function createPreferenceStore(opts: { redis: { get: (k:string)=>Promise<string|null>; set: (k:string,v:string)=>Promise<any> }; namespace: string }) {
  const key = (userId: string) => `${opts.namespace}:prefs:${userId}`;
  return {
    async get(userId: string): Promise<UserPreferences> {
      const raw = await opts.redis.get(key(userId));
      if (!raw) return { frequency: 'immediate', channels: {} };
      try { return JSON.parse(raw) as UserPreferences; } catch { return { frequency: 'immediate', channels: {} }; }
    },
    async put(userId: string, prefs: UserPreferences): Promise<void> {
      await opts.redis.set(key(userId), JSON.stringify(prefs));
    }
  };
}
