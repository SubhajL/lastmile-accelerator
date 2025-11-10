import { describe, it, expect, beforeEach, vi } from 'vitest';
import Fastify from 'fastify';
import request from 'supertest';
import { registerReadyRoute } from '../../routes/ready';
import * as db from '../../db';
import * as nats from '../../nats';

describe('/readyz readiness route', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('returns 200 when DB and NATS are healthy', async () => {
    vi.spyOn(db, 'query').mockResolvedValue({ rows: [{ ok: 1 }] } as any);
    vi.spyOn(nats, 'getNats').mockReturnValue({ flush: vi.fn().mockResolvedValue(undefined) } as any);

    const app = Fastify();
    registerReadyRoute(app);
    await app.ready();

    const res = await request(app.server).get('/readyz');
    expect(res.status).toBe(200);
    expect(res.body.ok).toBe(true);
  });

  it('returns 503 when DB is down', async () => {
    vi.spyOn(db, 'query').mockRejectedValue(new Error('db down'));
    vi.spyOn(nats, 'getNats').mockReturnValue({ flush: vi.fn().mockResolvedValue(undefined) } as any);

    const app = Fastify();
    registerReadyRoute(app);
    await app.ready();

    const res = await request(app.server).get('/readyz');
    expect(res.status).toBe(503);
    expect(res.body.ok).toBe(false);
    expect(res.body.details.db.healthy).toBe(false);
  });

  it('returns 503 when NATS is down', async () => {
    vi.spyOn(db, 'query').mockResolvedValue({ rows: [{ ok: 1 }] } as any);
    vi.spyOn(nats, 'getNats').mockImplementation(() => { throw new Error('nats down'); });

    const app = Fastify();
    registerReadyRoute(app);
    await app.ready();

    const res = await request(app.server).get('/readyz');
    expect(res.status).toBe(503);
    expect(res.body.ok).toBe(false);
    expect(res.body.details.nats.healthy).toBe(false);
  });
});