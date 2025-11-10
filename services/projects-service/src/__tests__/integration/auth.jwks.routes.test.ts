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

describe('JWKS-backed auth verification', () => {
let privKey: any;
  let pubJwk: JWK & { kid: string; alg: string; use?: string };
  let jwks: { app: any; url: string };

  beforeAll(async () => {
    const { publicKey, privateKey } = await generateKeyPair('RS256');
    privKey = privateKey;
    const jwk = await exportJWK(publicKey);
    pubJwk = { ...(jwk as JWK), kid: 'test-key', alg: 'RS256', use: 'sig' };
    jwks = await startJwksServer(pubJwk);
  });

  afterAll(async () => {
    await jwks.app.close();
  });

  async function sign(payload: Record<string, any>, overrides: { iss?: string; aud?: string; exp?: number; nbf?: number } = {}) {
    const now = Math.floor(Date.now() / 1000);
    const jwt = await new SignJWT(payload)
      .setProtectedHeader({ alg: 'RS256', kid: 'test-key', typ: 'JWT' })
      .setIssuer(overrides.iss ?? 'https://issuer.test')
      .setAudience(overrides.aud ?? 'projects-service')
      .setSubject('u1')
      .setIssuedAt(now)
      .setNotBefore(overrides.nbf ?? now - 1)
      .setExpirationTime(overrides.exp ?? now + 60 * 60)
      .sign(privKey);
    return jwt;
  }

  it('accepts a valid token via remote JWKS', async () => {
    const app = await createServer({
      serviceName: 'projects-service',
      // new auth option expected by implementation
      auth: { jwksUrl: jwks.url, issuer: 'https://issuer.test', audience: 'projects-service' },
    } as any);
    await app.ready();

    const token = await sign({ tenantId: 't1', scope: 'projects:read' });
    const res = await request(app.server)
      .get('/v1/projects')
      .set('Authorization', `Bearer ${token}`);

    expect([200, 500]).toContain(res.status);
    expect(res.status).not.toBe(401);
    expect(res.status).not.toBe(403);

    await app.close();
  });

  it('rejects when issuer or audience mismatch', async () => {
    const app = await createServer({
      serviceName: 'projects-service',
      auth: { jwksUrl: jwks.url, issuer: 'https://issuer.test', audience: 'projects-service' },
    } as any);
    await app.ready();

    const badIss = await sign({ tenantId: 't1', scope: 'projects:read' }, { iss: 'https://wrong-iss' });
    const res1 = await request(app.server)
      .get('/v1/projects')
      .set('Authorization', `Bearer ${badIss}`);
    expect(res1.status).toBe(401);

    const badAud = await sign({ tenantId: 't1', scope: 'projects:read' }, { aud: 'wrong-aud' });
    const res2 = await request(app.server)
      .get('/v1/projects')
      .set('Authorization', `Bearer ${badAud}`);
    expect(res2.status).toBe(401);

    await app.close();
  });
});