import { describe, it, expect, vi, beforeEach } from 'vitest';
import Fastify from 'fastify';
import request from 'supertest';
import { registerProjectRoutes } from '../../routes/projects';

const fakeTenant = 't1';
const deps = {
  extractTenantId: () => fakeTenant,
  services: {
    listProjects: vi.fn(async (tenantId: string) => [{ id: 'p1', tenant_id: tenantId, name: 'X' }]),
    getProject: vi.fn(async () => ({ id: 'p1', tenant_id: fakeTenant, name: 'X' })),
    createProject: vi.fn(async (_tenantId: string, input: any) => ({ id: 'p2', tenant_id: fakeTenant, ...input })),
    updateProject: vi.fn(async (_tenantId: string, _id: string, input: any) => ({ id: 'p1', tenant_id: fakeTenant, ...input })),
    deleteProject: vi.fn(async () => {}),
  },
};

describe('routes/projects.ts (integration-lite)', () => {
  let app: any;
beforeEach(async () => {
    app = Fastify();
    const { errorHandler } = await import('../../middleware/errorHandler');
    app.setErrorHandler(errorHandler as any);
    // attach a user with read+write scopes so route guards pass
    app.addHook('preHandler', async (req: any) => { req.user = { scope: 'projects:read projects:write' }; });
    registerProjectRoutes(app, deps);
    await app.ready();
  });

  it('GET /v1/projects returns list', async () => {
    const res = await request(app.server).get('/v1/projects');
    expect(res.status).toBe(200);
    expect(res.body[0].tenant_id).toBe(fakeTenant);
  });

  it('POST /v1/projects creates project', async () => {
    const res = await request(app.server).post('/v1/projects').send({ name: 'New' });
    expect(res.status).toBe(201);
    expect(res.body.name).toBe('New');
  });

  it('POST /v1/projects invalid payload returns 400', async () => {
    const res = await request(app.server).post('/v1/projects').send({});
    expect(res.status).toBe(400);
  });

  it('GET /v1/projects/:id returns project', async () => {
    const res = await request(app.server).get('/v1/projects/p1');
    expect(res.status).toBe(200);
    expect(res.body.id).toBe('p1');
  });

  it('PUT /v1/projects/:id updates project', async () => {
    const res = await request(app.server).put('/v1/projects/p1').send({ name: 'Renamed' });
    expect(res.status).toBe(200);
    expect(res.body.name).toBe('Renamed');
  });

  it('DELETE /v1/projects/:id returns 204', async () => {
    const res = await request(app.server).delete('/v1/projects/p1');
    expect(res.status).toBe(204);
  });
});