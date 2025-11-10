import { describe, it, expect, vi, beforeEach } from 'vitest';
import Fastify from 'fastify';
import request from 'supertest';
import { registerTenantRoutes } from '../../routes/tenants';

describe('routes/tenants.ts (integration-lite)', () => {
  let app: any;
  const deps = {
    services: {
      getTenant: vi.fn(async (tenantId: string) => ({ id: tenantId, name: 'Acme' })),
      listTenantMembers: vi.fn(async (tenantId: string) => ([{ id: 'm1', tenant_id: tenantId, email: 'a@x.com' }])),
    },
  };

beforeEach(async () => {
    app = Fastify();
    // attach a user with tenants:read scope so route guards pass
    app.addHook('preHandler', async (req: any) => { req.user = { scope: 'tenants:read' }; });
    registerTenantRoutes(app, { extractTenantId: () => 't1', services: deps.services });
    await app.ready();
  });

  it('GET /v1/tenants/:tenantId returns tenant', async () => {
    const res = await request(app.server).get('/v1/tenants/t1');
    expect(res.status).toBe(200);
    expect(res.body.id).toBe('t1');
  });

  it('GET /v1/tenants/:tenantId/members returns list', async () => {
    const res = await request(app.server).get('/v1/tenants/t1/members');
    expect(res.status).toBe(200);
    expect(res.body[0].tenant_id).toBe('t1');
  });
});