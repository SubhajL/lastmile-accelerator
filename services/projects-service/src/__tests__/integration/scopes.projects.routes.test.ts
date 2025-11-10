import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import Fastify from 'fastify';
import request from 'supertest';
import { createServer } from '../../index';
import { generateKeyPair, exportJWK, SignJWT, type JWK } from 'jose';

async function startJwksServer(jwk: JWK) {
  const app = Fastify({ logger: false });
  app.get('/.well-known/jwks.json', async (_req, reply) => {
    return reply.send({ keys: [jwk] });
  });
  await app.listen({ port: 0, host: '127.0.0.1' });
  const address = app.server.address();
  const port = typeof address === 'object' && address ? address.port : 0;
  const url = `http://127.0.0.1:${port}/.well-known/jwks.json`;
  return { app, url };
}

describe('scope enforcement on projects routes', () => {
let privKey: any;
  let jwks: { app: any; url: string };

  beforeAll(async () => {
    const { publicKey, privateKey } = await generateKeyPair('RS256');
    privKey = privateKey;
    const jwk = await exportJWK(publicKey);
    const pubJwk = { ...(jwk as JWK), kid: 'test-key', alg: 'RS256', use: 'sig' };
    jwks = await startJwksServer(pubJwk);
  });

  afterAll(async () => {
    await jwks.app.close();
  });

  async function tokenWithScope(scope: string) {
    const now = Math.floor(Date.now() / 1000);
    return new SignJWT({ tenantId: 't1', scope })
      .setProtectedHeader({ alg: 'RS256', kid: 'test-key', typ: 'JWT' })
      .setIssuer('https://issuer.test')
      .setAudience('projects-service')
      .setSubject('u1')
      .setIssuedAt(now)
      .setNotBefore(now - 1)
      .setExpirationTime(now + 3600)
      .sign(privKey);
  }

  it('GET /v1/projects requires projects:read', async () => {
    const app = await createServer({
      serviceName: 'projects-service',
      auth: { jwksUrl: jwks.url, issuer: 'https://issuer.test', audience: 'projects-service' },
    } as any);
    await app.ready();

    const noRead = await tokenWithScope('projects:write');
    const res1 = await request(app.server)
      .get('/v1/projects')
      .set('Authorization', `Bearer ${noRead}`);
    expect(res1.status).toBe(403);

    const hasRead = await tokenWithScope('projects:read');
    const res2 = await request(app.server)
      .get('/v1/projects')
      .set('Authorization', `Bearer ${hasRead}`);
    expect([200, 500]).toContain(res2.status);
    expect(res2.status).not.toBe(401);

    await app.close();
  });

  it('POST /v1/projects requires projects:write', async () => {
    const app = await createServer({
      serviceName: 'projects-service',
      auth: { jwksUrl: jwks.url, issuer: 'https://issuer.test', audience: 'projects-service' },
    } as any);
    await app.ready();

    const noWrite = await tokenWithScope('projects:read');
    const res1 = await request(app.server)
      .post('/v1/projects')
      .send({ name: 'p1' })
      .set('Authorization', `Bearer ${noWrite}`);
    expect(res1.status).toBe(403);

    const hasWrite = await tokenWithScope('projects:write');
    const res2 = await request(app.server)
      .post('/v1/projects')
      .send({ name: 'p2' })
      .set('Authorization', `Bearer ${hasWrite}`);
    expect([201, 500]).toContain(res2.status);
    expect(res2.status).not.toBe(401);

    await app.close();
  });
});