import { describe, it, expect } from 'vitest';
import { metricsMiddleware, getMetricsSnapshot, resetMetrics } from '../../middleware/metrics';

function makeReply() {
  const listeners: Record<string, Function[]> = {};
  return {
    statusCode: 0,
    on: (event: string, cb: Function) => {
      listeners[event] = listeners[event] || [];
      listeners[event].push(cb);
    },
    emit: (event: string) => {
      (listeners[event] || []).forEach((cb) => cb());
    },
  } as any;
}

describe('middleware/metrics.ts', () => {
  it('records request count and duration on finish', async () => {
    resetMetrics();
    const req: any = { method: 'GET', url: '/v1/projects' };
    const reply: any = makeReply();

    const mw = metricsMiddleware();
    const before = Date.now();
    await mw(req, reply);
    reply.statusCode = 200;
    reply.emit('finish');

    const snap = getMetricsSnapshot();
    expect(snap.totalRequests).toBe(1);
    expect(snap.totalErrors).toBe(0);
    expect(snap.durations.count).toBe(1);
    expect(snap.durations.min).toBeGreaterThanOrEqual(0);
    expect(snap.durations.max).toBeGreaterThanOrEqual(snap.durations.min);
  });

  it('increments error counter for 5xx', async () => {
    resetMetrics();
    const req: any = { method: 'POST', url: '/v1/projects' };
    const reply: any = makeReply();

    const mw = metricsMiddleware();
    await mw(req, reply);
    reply.statusCode = 500;
    reply.emit('finish');

    const snap = getMetricsSnapshot();
    expect(snap.totalRequests).toBe(1);
    expect(snap.totalErrors).toBe(1);
  });
});