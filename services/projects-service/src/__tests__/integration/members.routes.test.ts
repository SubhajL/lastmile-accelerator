import { describe, it, expect, vi, beforeEach } from 'vitest';
import Fastify from 'fastify';
import request from 'supertest';
import { registerMemberRoutes } from '../../routes/members';

const deps = {
  services: {
    addMember: vi.fn(async (tenantId: string, input: any) => ({ id: 'm1', tenant_id: tenantId, ...input })),
    updateMemberRole: vi.fn(async (_tenantId: string, memberId: string, newRole: string) => ({ id: memberId, role: newRole })),
    removeMember: vi.fn(async () => {}),
  },
};

describe('routes/members.ts (integration-lite)', () => {
  let app: any;

  beforeEach(async () => {
    app = Fastify();
    const { errorHandler } = await import('../../middleware/errorHandler');
app.setErrorHandler(errorHandler as any);
    // attach a user with members:write scope and owner role so route guards pass
    app.addHook('preHandler', async (req: any) => { req.user = { scope: 'members:write', role: 'owner' }; });
    registerMemberRoutes(app, deps);
    await app.ready();
  });

  it('POST /v1/tenants/:tenantId/members creates member', async () => {
    const res = await request(app.server).post('/v1/tenants/t1/members').send({ email: 'u@x.com', role: 'developer' });
    expect(res.status).toBe(201);
    expect(res.body.email).toBe('u@x.com');
  });

  it('PUT /v1/members/:memberId updates role', async () => {
    const res = await request(app.server).put('/v1/members/m1').send({ role: 'admin' });
    expect(res.status).toBe(200);
    expect(res.body.role).toBe('admin');
  });

  it('POST /v1/tenants/:tenantId/members invalid role 400', async () => {
    const res = await request(app.server).post('/v1/tenants/t1/members').send({ email: 'u@x.com', role: 'boss' });
    expect(res.status).toBe(400);
  });

  it('DELETE /v1/members/:memberId removes member', async () => {
    const res = await request(app.server).delete('/v1/members/m1');
    expect(res.status).toBe(204);
  });
});