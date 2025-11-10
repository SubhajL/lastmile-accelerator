import { describe, it, expect } from 'vitest';
import { createRoutingEngine } from './engine.js';

function mockPrefs(p: any) {
  return { get: async () => p } as any;
}

describe('routing/engine', () => {
  it('blocks when channel disabled', async () => {
    const engine = createRoutingEngine({ prefs: mockPrefs({ channels: { email: false } }), now: () => Date.now() });
    const d = await engine.evaluate({ userId: 'u', channel: 'email' } as any);
    expect(d).toEqual({ status: 'block', reason: 'channel_disabled' });
  });

  it('defers during quiet hours', async () => {
    // 23:00 local time
    const fixed = new Date(); fixed.setHours(23, 0, 0, 0);
    const engine = createRoutingEngine({ prefs: mockPrefs({ quietHours: { start: '22:00', end: '07:00' } }), now: () => fixed.getTime() });
    const d = await engine.evaluate({ userId: 'u', channel: 'email' } as any);
    expect(d).toEqual({ status: 'defer', reason: 'quiet_hours' });
  });

  it('applies digest and mute frequencies', async () => {
    const e1 = createRoutingEngine({ prefs: mockPrefs({ frequency: 'mute' }), now: () => Date.now() });
    const b = await e1.evaluate({ userId: 'u', channel: 'email' } as any);
    expect(b).toEqual({ status: 'block', reason: 'muted' });

    const e2 = createRoutingEngine({ prefs: mockPrefs({ frequency: 'digest_hourly' }), now: () => Date.now() });
    const d = await e2.evaluate({ userId: 'u', channel: 'email' } as any);
    expect(d).toEqual({ status: 'defer', reason: 'digest_hourly' });
  });
});
