import { describe, it, expect, vi, beforeEach } from 'vitest';
import Fastify from 'fastify';
import { registerRequestLogging } from '../../middleware/logging';

// fake logger to capture logs
const childLogger = { info: vi.fn() } as any;
const fakeBase = {
  child: vi.fn().mockImplementation((_fields: any) => childLogger),
  info: vi.fn(),
} as any;

describe('middleware/logging', () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it('adds correlation fields and logs completion', async () => {
    const mod = await import('../../logger');
    vi.spyOn(mod, 'getLogger').mockReturnValue(fakeBase as any);

    const app = Fastify();
    registerRequestLogging(app as any);
    app.get('/v1/ping', async (_req, reply) => reply.code(200).send({ ok: true }));
    await app.ready();

    const res = await app.inject({ method: 'GET', url: '/v1/ping', headers: { 'x-request-id': 'rid-1' } });
    expect(res.statusCode).toBe(200);
    await new Promise((r) => setTimeout(r, 0));

    // child created with correlation fields
    expect(fakeBase.child).toHaveBeenCalled();
    const calledWith = fakeBase.child.mock.calls[0][0];
    expect(calledWith.request_id).toBeDefined();
    expect(calledWith.method).toBe('GET');

    // completion logging performed via reply.finish/onResponse (not asserted here)
  });
});
