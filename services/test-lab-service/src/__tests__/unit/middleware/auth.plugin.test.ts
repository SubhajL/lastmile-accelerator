import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import Fastify from 'fastify';
import type { FastifyInstance } from 'fastify';
import { registerAuthPlugin, requireAuth } from '../../../middleware/auth.js';
import { registerErrorHandler } from '../../../middleware/error-handler.js';
import { applyTestEnv } from '../../fixtures/env.js';

vi.mock('../../../lib/jwks.js', () => ({
  verifyJwt: vi.fn(async () => ({
    sub: 'sub-1', tenant_id: 't-1', user_id: 'u-1', scopes: ['x:read'],
    iss: 'https://auth.example.com/', aud: 'test-lab-service',
    exp: Math.floor(Date.now() / 1000) + 60, iat: Math.floor(Date.now() / 1000),
  })),
}));

describe('registerAuthPlugin (JWKS-only decorator)', () => {
  let app: FastifyInstance;

  beforeEach(async () => {
    applyTestEnv();
    app = Fastify();
    registerErrorHandler(app);
  });

  afterEach(async () => {
    await app.close();
    vi.clearAllMocks();
  });

  it('decorates request.user without installing fastify-jwt', async () => {
    await registerAuthPlugin(app as any);

    app.get('/check', async (req) => {
      const hasUser = Object.prototype.hasOwnProperty.call(req, 'user');
      const hasJwtVerify = typeof (req as any).jwtVerify !== 'undefined';
      return { hasUser, hasJwtVerify };
    });

    const res = await app.inject({ method: 'GET', url: '/check' });
    expect(res.statusCode).toBe(200);
    const body = res.json();
    expect(body.hasUser).toBe(true);
    expect(body.hasJwtVerify).toBe(false);
  });

  it('works with requireAuth path (JWKS verify)', async () => {
    // minimal env for config used by requireAuth
    process.env.JWT_ISSUER = 'https://auth.example.com/';
    process.env.JWT_AUDIENCE = 'test-lab-service';
    process.env.JWT_JWKS_URL = 'https://jwks.local/.well-known/jwks.json';
    process.env.JWT_ALG = 'RS256';
    process.env.JWT_CLOCK_SKEW_SEC = '60';
    const { __resetConfigForTests } = await import('../../../config.js');
    __resetConfigForTests();

    await registerAuthPlugin(app as any);
    app.get('/secure', { preHandler: requireAuth as any }, async (req) => {
      const u: any = (req as any).user;
      return { uid: u?.userId, scopes: u?.scopes };
    });

    const res = await app.inject({ method: 'GET', url: '/secure', headers: { authorization: 'Bearer token-abc' } });
    expect(res.statusCode).toBe(200);
    const body = res.json();
    expect(body.uid).toBe('u-1');
    expect(body.scopes).toEqual(['x:read']);
  });
});
