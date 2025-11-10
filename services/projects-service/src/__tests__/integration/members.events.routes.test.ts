import { describe, it, expect, vi, beforeEach } from 'vitest';
import Fastify from 'fastify';
import request from 'supertest';
import { registerMemberRoutes } from '../../routes/members';
import * as events from '../../events/eventPublisher';
import * as db from '../../db';

describe('members routes event publishing (integration-lite, real service)', () => {
  let app: any;

  beforeEach(async () => {
    vi.restoreAllMocks();
    app = Fastify();
    const { errorHandler } = await import('../../middleware/errorHandler');
app.setErrorHandler(errorHandler as any);
    app.addHook('preHandler', async (req: any) => { req.user = { role: 'owner', scope: 'members:write' }; });

    registerMemberRoutes(app);
    await app.ready();
  });

  it('POST /v1/tenants/:tenantId/members publishes member.added', async () => {
    vi.spyOn(db, 'query').mockResolvedValue({ rows: [{ id: 'm1', tenant_id: 't1', email: 'u@x.com', role: 'developer' }] } as any);
    const spy = vi.spyOn(events, 'publishMemberEvent').mockResolvedValue();

    const res = await request(app.server).post('/v1/tenants/t1/members').send({ email: 'u@x.com', role: 'developer' });
    expect(res.status).toBe(201);
    expect(spy).toHaveBeenCalledWith('added', expect.objectContaining({ tenantId: 't1' }), undefined);
  });
});