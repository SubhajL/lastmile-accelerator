import { describe, it, expect, vi, beforeEach } from 'vitest';
import Fastify from 'fastify';
import request from 'supertest';
import { registerProjectRoutes } from '../../routes/projects';
import * as events from '../../events/eventPublisher';
import * as db from '../../db';

describe('projects routes event publishing (integration-lite, real service)', () => {
  let app: any;

  beforeEach(async () => {
    vi.restoreAllMocks();
    app = Fastify();
    const { errorHandler } = await import('../../middleware/errorHandler');
    app.setErrorHandler(errorHandler as any);
// attach a fake user with sub and write scope
    app.addHook('preHandler', async (req: any) => { req.user = { sub: 'u1', scope: 'projects:write' }; });

    registerProjectRoutes(app, { extractTenantId: () => 't1' });
    await app.ready();
  });

  it('POST /v1/projects publishes project.created on success', async () => {
    // stub transaction to succeed without real DB
    vi.spyOn(db, 'transaction').mockImplementation(async (fn: any) => {
      const client = { query: vi.fn().mockResolvedValue({ rows: [{ id: 'pX' }] }) } as any;
      const result = await fn(client);
      return { id: 'pX', ...result };
    });
    const spy = vi.spyOn(events, 'publishProjectEvent').mockResolvedValue();

    const res = await request(app.server).post('/v1/projects').send({ name: 'New' });
    expect(res.status).toBe(201);
    expect(spy).toHaveBeenCalledWith('created', expect.objectContaining({ tenantId: 't1' }), undefined);
  });
});