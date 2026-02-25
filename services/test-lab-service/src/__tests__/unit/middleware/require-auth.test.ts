import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';
import Fastify from 'fastify';
import { applyTestEnv } from '../../fixtures/env.js';

vi.mock('../../../lib/jwks.js', () => {
  return { verifyJwt: vi.fn() };
});

const ISSUER = 'https://auth.example.com/';
const AUDIENCE = 'test-lab-service';
const JWKS_URL = 'https://jwks.local/.well-known/jwks.json';

async function buildApp() {
  const app = Fastify();
  const { registerErrorHandler } = await import('../../../middleware/error-handler.js');
  registerErrorHandler(app);
  const { requireAuth } = await import('../../../middleware/auth.js');
  app.get('/secure', { preHandler: requireAuth as any }, async (req) => {
    const u: any = (req as any).user;
    return { uid: u?.userId ?? null, scopes: u?.scopes ?? null };
  });
  return app;
}

describe('requireAuth preHandler', () => {
  beforeEach(async () => {
    applyTestEnv();
    process.env.JWT_ISSUER = ISSUER;
    process.env.JWT_AUDIENCE = AUDIENCE;
    process.env.JWT_JWKS_URL = JWKS_URL;
    process.env.JWT_ALG = 'RS256';
    process.env.JWT_CLOCK_SKEW_SEC = '60';
    const { __resetConfigForTests } = await import('../../../config.js');
    __resetConfigForTests();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('returns 200 and populates user on valid Bearer token', async () => {
    const app = await buildApp();
    const jwks: any = await import('../../../lib/jwks.js');
    jwks.verifyJwt.mockResolvedValue({
      sub: 'sub-1', tenant_id: 't-1', user_id: 'u-1', scopes: ['x:read'], iss: ISSUER, aud: AUDIENCE,
      exp: Math.floor(Date.now() / 1000) + 60, iat: Math.floor(Date.now() / 1000),
    });

    const res = await app.inject({ method: 'GET', url: '/secure', headers: { authorization: 'Bearer token-abc' } });
    expect(res.statusCode).toBe(200);
    const body = res.json();
    expect(body.uid).toBe('u-1');
    expect(body.scopes).toEqual(['x:read']);

    expect(jwks.verifyJwt).toHaveBeenCalled();
    const args = jwks.verifyJwt.mock.calls[0][0];
    expect(args.token).toBe('token-abc');
    expect(args.issuer).toBe(ISSUER);
    expect(args.audience).toBe(AUDIENCE);
    expect(args.alg).toBe('RS256');
    expect(args.clockSkewSec).toBe(60);
    expect(args.jwksUrl).toBe(JWKS_URL);

    await app.close();
  });

  it('401 when Authorization header missing', async () => {
    const app = await buildApp();
    const res = await app.inject({ method: 'GET', url: '/secure' });
    expect(res.statusCode).toBe(401);
    await app.close();
  });

  it('401 when header malformed (no Bearer prefix)', async () => {
    const app = await buildApp();
    const res = await app.inject({ method: 'GET', url: '/secure', headers: { authorization: 'Basic abc' } });
    expect(res.statusCode).toBe(401);
    await app.close();
  });

  it('401 when verification fails', async () => {
    const app = await buildApp();
    const jwks: any = await import('../../../lib/jwks.js');
    jwks.verifyJwt.mockRejectedValue(new Error('bad token'));
    const res = await app.inject({ method: 'GET', url: '/secure', headers: { authorization: 'Bearer bad' } });
    expect(res.statusCode).toBe(401);
    await app.close();
  });
});
