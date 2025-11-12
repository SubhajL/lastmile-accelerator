import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { createApp } from '../../app.js';
import { authenticateRequest, requireScopes } from '../../middleware/auth.js';
import { registerErrorHandler } from '../../middleware/error-handler.js';
import { createMockJWT, TEST_JWT_SECRET } from '../fixtures/jwt-helpers.js';
import { applyTestEnv } from '../fixtures/env.js';

// Set minimal env for config loading with safe placeholders
applyTestEnv();
process.env.JWT_JWKS_URL = TEST_JWT_SECRET; // use shared secret in tests
process.env.JWT_ISSUER = 'https://auth.example.com/';
process.env.JWT_AUDIENCE = 'test-lab-service';

describe('Server bootstrap', () => {
  let app: Awaited<ReturnType<typeof createApp>>;

  beforeEach(async () => {
    app = await createApp();
  });

  afterEach(async () => {
    await app.close();
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
    app.get('/_protected', { preHandler: [authenticateRequest as any, requireScopes('test:read') as any] }, async (req) => {
      const u = (req as any).user;
      return { ok: true, uid: u?.userId };
    });

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
