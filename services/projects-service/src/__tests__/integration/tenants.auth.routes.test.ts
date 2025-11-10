import { describe, it, expect, beforeEach } from 'vitest';
import request from 'supertest';
import { createServer } from '../../index';
import { __setJwtVerifier } from '../../middleware/auth';

describe('tenants routes tenant isolation (auth)', () => {
beforeEach(() => {
    __setJwtVerifier(async (token: string) => {
      if (token === 'valid') return { sub: 'u1', tenantId: 't1', scope: 'tenants:read' } as any;
      throw new Error('invalid');
    });
  });

  it('GET /v1/tenants/:tenantId returns 403 when tenantId != JWT', async () => {
    const app = await createServer({ serviceName: 'projects-service' });
    await app.ready();
    const res = await request(app.server)
      .get('/v1/tenants/t2')
      .set('Authorization', 'Bearer valid');
    expect(res.status).toBe(403);
  });

  it('GET /v1/tenants/:tenantId/members returns 403 when tenantId != JWT', async () => {
    const app = await createServer({ serviceName: 'projects-service' });
    await app.ready();
    const res = await request(app.server)
      .get('/v1/tenants/t3/members')
      .set('Authorization', 'Bearer valid');
    expect(res.status).toBe(403);
  });
});