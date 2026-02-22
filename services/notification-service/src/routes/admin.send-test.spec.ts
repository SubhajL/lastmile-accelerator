import { describe, it, expect, beforeAll } from 'vitest';
import Fastify from 'fastify';
import { jwtPlugin } from '../auth/jwt.js';
import { registerAdminRoutes } from './admin.send-test.js';

function makeApp(secret = 'testsecret', enqueue = async () => 'job-1') {
  const app = Fastify();
  app.register(jwtPlugin, { secret });
  registerAdminRoutes(app as any, { enqueue } as any);
  return app;
}

describe('POST /admin/send-test', () => {
  it('rejects without JWT (401)', async () => {
    const app = makeApp();
    await app.ready();
    const res = await app.inject({ method: 'POST', url: '/admin/send-test', payload: { tenantId: 't', userId: 'u' } });
    expect(res.statusCode).toBe(401);
  });

  it('rejects without admin role (403)', async () => {
    const app = makeApp();
    await app.ready();
    const token = (app as any).jwt.sign({ roles: ['user'] });
    const res = await app.inject({ method: 'POST', url: '/admin/send-test', headers: { authorization: `Bearer ${token}` }, payload: { tenantId: 't', userId: 'u' } });
    expect(res.statusCode).toBe(403);
  });

  it('enqueues job and returns 202 with jobId', async () => {
    const seen: any[] = [];
    const app = makeApp('testsecret', async (job: any) => { seen.push(job); return 'job-xyz'; });
    await app.ready();
    const token = (app as any).jwt.sign({ roles: ['admin'] });
    const res = await app.inject({ method: 'POST', url: '/admin/send-test', headers: { authorization: `Bearer ${token}` }, payload: { tenantId: 't1', userId: 'u1', channel: 'email', templateName: 'hello', payload: { x: 1 }, priority: 'high' } });
    expect(res.statusCode).toBe(202);
    const json = res.json();
    expect(json).toEqual({ ok: true, jobId: 'job-xyz' });
    expect(seen[0]).toEqual({ tenantId: 't1', userId: 'u1', channel: 'email', templateName: 'hello', payload: { x: 1 }, priority: 'high', maxAttempts: 3 });
  });

  it('rejects invalid priority (400)', async () => {
    const enqueue = async () => 'job-1';
    const app = makeApp('testsecret', enqueue);
    await app.ready();
    const token = (app as any).jwt.sign({ roles: ['admin'] });
    const res = await app.inject({
      method: 'POST',
      url: '/admin/send-test',
      headers: { authorization: `Bearer ${token}` },
      payload: { tenantId: 't1', userId: 'u1', priority: 'urgent' },
    });
    expect(res.statusCode).toBe(400);
  });

  it('rejects invalid channel (400)', async () => {
    const enqueue = async () => 'job-1';
    const app = makeApp('testsecret', enqueue);
    await app.ready();
    const token = (app as any).jwt.sign({ roles: ['admin'] });
    const res = await app.inject({
      method: 'POST',
      url: '/admin/send-test',
      headers: { authorization: `Bearer ${token}` },
      payload: { tenantId: 't1', userId: 'u1', channel: 'fax' },
    });
    expect(res.statusCode).toBe(400);
  });

  it('rejects invalid maxAttempts (400)', async () => {
    const enqueue = async () => 'job-1';
    const app = makeApp('testsecret', enqueue);
    await app.ready();
    const token = (app as any).jwt.sign({ roles: ['admin'] });
    const res = await app.inject({
      method: 'POST',
      url: '/admin/send-test',
      headers: { authorization: `Bearer ${token}` },
      payload: { tenantId: 't1', userId: 'u1', maxAttempts: 0 },
    });
    expect(res.statusCode).toBe(400);
  });

  it('rejects non-object payload (400)', async () => {
    const enqueue = async () => 'job-1';
    const app = makeApp('testsecret', enqueue);
    await app.ready();
    const token = (app as any).jwt.sign({ roles: ['admin'] });
    const res = await app.inject({
      method: 'POST',
      url: '/admin/send-test',
      headers: { authorization: `Bearer ${token}` },
      payload: { tenantId: 't1', userId: 'u1', payload: 'nope' },
    });
    expect(res.statusCode).toBe(400);
  });
});
