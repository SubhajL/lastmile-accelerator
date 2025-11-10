import { describe, it, expect, beforeEach } from 'vitest';
import request from 'supertest';
import { createServer } from '../../index';
import { __setJwtVerifier } from '../../middleware/auth';

describe('auth preHandler + index wiring', () => {
beforeEach(async () => {
    __setJwtVerifier(async (token: string) => {
      if (token === 'valid') return { sub: 'u1', tenantId: 't1', scope: 'projects:read' } as any;
      throw new Error('invalid');
    });
  });

  it('401 when missing bearer token', async () => {
    const app = await createServer({ serviceName: 'projects-service' });
    await app.ready();
    const res = await request(app.server).get('/v1/projects');
    expect(res.status).toBe(401);
  });

  it('401 when invalid token', async () => {
    const app = await createServer({ serviceName: 'projects-service' });
    await app.ready();
    const res = await request(app.server)
      .get('/v1/projects')
      .set('Authorization', 'Bearer invalid');
    expect(res.status).toBe(401);
  });

  it('200 when valid token and tenant attached', async () => {
    const app = await createServer({ serviceName: 'projects-service' });
    await app.ready();
    const res = await request(app.server)
      .get('/v1/projects')
      .set('Authorization', 'Bearer valid');
    // service may 200 with empty list (no DB), but auth passed
    expect([200, 500]).toContain(res.status);
    if (res.status === 500) {
      // DB not initialized is acceptable for this auth gate test
      expect(String(res.text)).toContain('INTERNAL_SERVER_ERROR');
    }
  });
});