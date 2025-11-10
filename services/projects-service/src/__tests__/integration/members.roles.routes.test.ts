import { describe, it, expect, beforeEach } from 'vitest';
import Fastify from 'fastify';
import request from 'supertest';
import { registerMemberRoutes } from '../../routes/members';

const deps = {
  services: {
    addMember: async (tenantId: string, input: any) => ({ id: 'm1', tenant_id: tenantId, ...input }),
    updateMemberRole: async (_tenantId: string, memberId: string, newRole: string) => ({ id: memberId, role: newRole }),
    removeMember: async () => {},
  },
};

describe('members routes role enforcement (owner-only)', () => {
  let app: any;

  beforeEach(async () => {
    app = Fastify();
    // user with members:write but NOT owner role
    app.addHook('preHandler', async (req: any) => { req.user = { scope: 'members:write', role: 'developer' }; });
    registerMemberRoutes(app, deps);
    await app.ready();
  });

  it('POST /v1/tenants/:tenantId/members requires owner role (403 for non-owner)', async () => {
    const res = await request(app.server).post('/v1/tenants/t1/members').send({ email: 'u@x.com', role: 'developer' });
    expect(res.status).toBe(403);
  });

  it('PUT /v1/members/:memberId requires owner role (403 for non-owner)', async () => {
    const res = await request(app.server).put('/v1/members/m1').send({ role: 'admin' });
    expect(res.status).toBe(403);
  });

  it('DELETE /v1/members/:memberId requires owner role (403 for non-owner)', async () => {
    const res = await request(app.server).delete('/v1/members/m1');
    expect(res.status).toBe(403);
  });
});