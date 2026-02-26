import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { createApp } from '../../app.js';
import { authenticateRequest, requireScopes } from '../../middleware/auth.js';
import { registerErrorHandler } from '../../middleware/error-handler.js';
import { createMockJWT, TEST_JWT_SECRET } from '../fixtures/jwt-helpers.js';
import { applyTestEnv } from '../fixtures/env.js';

// Mock Redis so tests run without a live Redis instance
vi.mock('../../clients/redis.js', () => ({
  createRedisClient: vi.fn().mockResolvedValue({
    get: vi.fn().mockResolvedValue(null),
    set: vi.fn().mockResolvedValue(undefined),
    del: vi.fn().mockResolvedValue(undefined),
    incr: vi.fn().mockResolvedValue(1),
    incrWithExpire: vi.fn().mockResolvedValue(1),
    expire: vi.fn().mockResolvedValue(undefined),
    close: vi.fn().mockResolvedValue(undefined),
  }),
}));

// Set minimal env for config loading with safe placeholders
applyTestEnv();
process.env.JWT_JWKS_URL = TEST_JWT_SECRET; // use shared secret in tests
process.env.JWT_ISSUER = 'https://auth.example.com/';
process.env.JWT_AUDIENCE = 'test-lab-service';

// Mock JWKS verifier to decode JWT for testing
vi.mock('../../lib/jwks.js', () => ({
  verifyJwt: vi.fn(async (opts: { token: string }) => {
    const parts = opts.token.split('.');
    if (parts.length !== 3) throw new Error('Invalid token');
    const payload = JSON.parse(Buffer.from(parts[1], 'base64url').toString());
    return payload;
  }),
}));

describe('Server bootstrap', () => {
  let app: Awaited<ReturnType<typeof createApp>>;

  beforeEach(async () => {
    app = await createApp();
  });

  afterEach(async () => {
    await app.close();
    vi.restoreAllMocks();
    vi.clearAllMocks();
  });

  it('exposes /healthz', async () => {
    const res = await app.inject({ method: 'GET', url: '/healthz' });
    expect(res.statusCode).toBe(200);
    expect(res.body).toBe('ok');
  });

  it('exposes /metrics in Prometheus format', async () => {
    const res = await app.inject({ method: 'GET', url: '/metrics' });
    expect(res.statusCode).toBe(200);
    expect(res.headers['content-type']).toContain('text/plain');
    expect(res.body).toContain('# HELP');
    expect(res.body).toContain('# TYPE');
  });

  it('supports protected routes with JWT + scopes', async () => {
    // Mount a protected route for testing
    app.get(
      '/_protected',
      { preHandler: [authenticateRequest, requireScopes('test:read')] },
      async (req) => {
      const u = (req as any).user;
      return { ok: true, uid: u?.userId };
      }
    );

    const token = createMockJWT({ scopes: ['test:read'] });

    const ok = await app.inject({ method: 'GET', url: '/_protected', headers: { authorization: `Bearer ${token}` } });
    expect(ok.statusCode).toBe(200);

    const denied = await app.inject({ method: 'GET', url: '/_protected' });
    expect(denied.statusCode).toBe(401);
  });

  it('global error handler returns 500 for unhandled errors', async () => {
    app.get('/boom', async () => {
      throw new Error('kaboom');
    });
    const res = await app.inject({ method: 'GET', url: '/boom' });
    expect(res.statusCode).toBe(500);
    expect(res.json().error).toBe('Internal Server Error');
  });
});
