import { describe, it, expect, vi, beforeEach } from 'vitest';
import Fastify from 'fastify';
import request from 'supertest';
import { registerEnvironmentRoutes } from '../../routes/environments';

const deps = {
  extractTenantId: () => 't1',
  services: {
    listEnvironments: vi.fn(async (_tenant: string, projectId: string) => [{ id: 'e1', project_id: projectId, name: 'dev' }]),
    getEnvironment: vi.fn(async (_tenant: string, projectId: string, envId: string) => ({ id: envId, project_id: projectId, name: 'dev' })),
    createEnvironment: vi.fn(async (_tenant: string, projectId: string, input: any) => ({ id: 'e2', project_id: projectId, ...input })),
    updateEnvironment: vi.fn(async (_tenant: string, _pid: string, envId: string, input: any) => ({ id: envId, ...input })),
    deleteEnvironment: vi.fn(async () => {}),
    setIngestionModes: vi.fn(async () => {}),
  },
};

describe('routes/environments.ts (integration-lite)', () => {
  let app: any;
  beforeEach(async () => {
    app = Fastify();
    const { errorHandler } = await import('../../middleware/errorHandler');
app.setErrorHandler(errorHandler as any);
    // attach a user with env read+write scopes so route guards pass
    app.addHook('preHandler', async (req: any) => { req.user = { scope: 'environments:read environments:write' }; });
    registerEnvironmentRoutes(app, deps);
    await app.ready();
  });

  it('GET /v1/projects/:projectId/environments returns list', async () => {
    const res = await request(app.server).get('/v1/projects/p1/environments');
    expect(res.status).toBe(200);
    expect(res.body[0].project_id).toBe('p1');
  });

  it('POST /v1/projects/:projectId/environments creates environment', async () => {
    const res = await request(app.server).post('/v1/projects/p1/environments').send({ name: 'qa' });
    expect(res.status).toBe(201);
    expect(res.body.name).toBe('qa');
  });

  it('POST /v1/projects/:projectId/environments invalid payload 400', async () => {
    const res = await request(app.server).post('/v1/projects/p1/environments').send({});
    expect(res.status).toBe(400);
  });

  it('PUT /v1/projects/:projectId/environments/:envId updates environment', async () => {
    const res = await request(app.server).put('/v1/projects/p1/environments/e1').send({ config: { a: 2 } });
    expect(res.status).toBe(200);
    expect(res.body.config.a).toBe(2);
  });

  it('DELETE /v1/projects/:projectId/environments/:envId deletes environment', async () => {
    const res = await request(app.server).delete('/v1/projects/p1/environments/e1');
    expect(res.status).toBe(204);
  });

  it('POST /v1/projects/:projectId/ingestion-modes sets modes', async () => {
    const res = await request(app.server).post('/v1/projects/p1/ingestion-modes').send({ modes: ['A','B'], defaultMode: 'A' });
    expect(res.status).toBe(200);
  });
});