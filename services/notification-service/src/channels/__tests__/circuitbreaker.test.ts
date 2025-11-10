import { describe, it, expect, vi } from 'vitest';
import { createCircuitBreaker } from '../circuitbreaker.js';

function nowSeq() {
  let t = 0;
  return () => (t += 1000);
}

describe('channels/circuitbreaker', () => {
  it('opens after threshold and blocks calls', async () => {
    const now = nowSeq();
    const breaker = createCircuitBreaker({ failureThreshold: 2, halfOpenAfterMs: 5000, windowSize: 3, now });

    try { await breaker.execute(async () => { throw new Error('x'); }); } catch {}
    try { await breaker.execute(async () => { throw new Error('y'); }); } catch {}

    await expect(breaker.execute(async () => 'ok')).rejects.toThrow(/breaker open/);
  });

  it('half-open permits a probe and closes on success', async () => {
    let t = 0; const now = () => t;
    const breaker = createCircuitBreaker({ failureThreshold: 1, halfOpenAfterMs: 1000, windowSize: 2, now });

    // fail once to open
    try { await breaker.execute(async () => { throw new Error('x'); }); } catch {}
    t = 1500; // advance

    const res = await breaker.execute(async () => 'ok');
    expect(res).toBe('ok');

    // closed again; subsequent calls allowed
    await expect(breaker.execute(async () => 'ok2')).resolves.toBe('ok2');
  });
});
