import { describe, it, expect, vi } from 'vitest';
import { createPreferenceStore, type UserPreferences } from './store.js';

function mockRedis() {
  const kv = new Map<string, string>();
  return {
    get: vi.fn(async (k: string) => kv.get(k) ?? null),
    set: vi.fn(async (k: string, v: string) => { kv.set(k, v); })
  } as any;
}

describe('prefs/store', () => {
  it('returns defaults when missing and persists prefs', async () => {
    const redis = mockRedis();
    const store = createPreferenceStore({ redis, namespace: 'np' });

    const def = await store.get('u1');
    expect(def.frequency).toBe('immediate');

    const prefs: UserPreferences = { channels: { email: true, sms: false }, quietHours: { start: '22:00', end: '07:00' }, frequency: 'digest_daily' };
    await store.put('u1', prefs);

    const got = await store.get('u1');
    expect(got.channels?.sms).toBe(false);
    expect(got.frequency).toBe('digest_daily');
  });
});
