import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import Fastify from 'fastify';
import type { FastifyInstance } from 'fastify';
import { registerAuthPlugin, requireAuth, requireScopes } from '../../../middleware/auth.js';
import {
  createMockJWT,
  createExpiredJWT,
  createInvalidSignatureJWT,
  TEST_JWT_SECRET,
} from '../../fixtures/jwt-helpers.js';
import { registerErrorHandler } from '../../../middleware/error-handler.js';

// Mock JWKS verifier to use test JWT secret instead of fetching from URL
vi.mock('../../../lib/jwks.js', () => ({
  verifyJwt: vi.fn(async (opts: { token: string }) => {
    // Simple JWT decode for testing (not secure, only for tests)
    const parts = opts.token.split('.');
    if (parts.length !== 3) throw new Error('Invalid token');
    const payload = JSON.parse(Buffer.from(parts[1], 'base64url').toString());
    return payload;
  }),
}));

describe('Auth Middleware', () => {
  let app: FastifyInstance;

  beforeEach(async () => {
    app = Fastify();
    registerErrorHandler(app);

    // Register auth plugin (decorates request.user)
    await registerAuthPlugin(app);
  });

  afterEach(async () => {
    await app.close();
    vi.restoreAllMocks();
    vi.clearAllMocks();
  });

  describe('requireAuth', () => {
    it('should accept valid JWT Bearer token', async () => {
      const token = createMockJWT();

      app.get('/test', { preHandler: requireAuth }, async (request) => {
        return { user: request.user };
      });

      const response = await app.inject({
        method: 'GET',
        url: '/test',
        headers: {
          authorization: `Bearer ${token}`,
        },
      });

      expect(response.statusCode).toBe(200);
      const body = response.json();
      expect(body.user).toBeDefined();
      expect(body.user.tenantId).toBe('tenant-abc');
      expect(body.user.userId).toBe('user-123');
    });

    it('should reject missing Authorization header', async () => {
      app.get('/test', { preHandler: requireAuth }, async () => {
        return { success: true };
      });

      const response = await app.inject({
        method: 'GET',
        url: '/test',
      });

      expect(response.statusCode).toBe(401);
      expect(response.json().error).toContain('Missing or invalid Authorization header');
    });

    it('should reject malformed Authorization header (no Bearer)', async () => {
      const token = createMockJWT();

      app.get('/test', { preHandler: requireAuth }, async () => {
        return { success: true };
      });

      const response = await app.inject({
        method: 'GET',
        url: '/test',
        headers: {
          authorization: token, // Missing "Bearer " prefix
        },
      });

      expect(response.statusCode).toBe(401);
    });

    it('should reject invalid JWT signature', async () => {
      const token = createInvalidSignatureJWT();

      app.get('/test', { preHandler: requireAuth }, async () => {
        return { success: true };
      });

      const response = await app.inject({
        method: 'GET',
        url: '/test',
        headers: {
          authorization: `Bearer ${token}`,
        },
      });

      expect(response.statusCode).toBe(401);
    });

    it('should reject expired JWT token', async () => {
      const token = createExpiredJWT();

      app.get('/test', { preHandler: requireAuth }, async () => {
        return { success: true };
      });

      const response = await app.inject({
        method: 'GET',
        url: '/test',
        headers: {
          authorization: `Bearer ${token}`,
        },
      });

      expect(response.statusCode).toBe(401);
      expect(response.json().error).toContain('Invalid token');
    });

    it('should populate request.user with correct claims', async () => {
      const token = createMockJWT({
        tenant_id: 'custom-tenant',
        user_id: 'custom-user',
        scopes: ['admin:all'],
      });

      app.get('/test', { preHandler: requireAuth }, async (request) => {
        return { user: request.user };
      });

      const response = await app.inject({
        method: 'GET',
        url: '/test',
        headers: {
          authorization: `Bearer ${token}`,
        },
      });

      expect(response.statusCode).toBe(200);
      const body = response.json();
      expect(body.user.tenantId).toBe('custom-tenant');
      expect(body.user.userId).toBe('custom-user');
      expect(body.user.scopes).toEqual(['admin:all']);
    });
  });

  describe('requireScopes()', () => {
    it('should allow request with required single scope', async () => {
      const token = createMockJWT({ scopes: ['test:write'] });

      app.get(
        '/test',
        {
          preHandler: [requireAuth, requireScopes('test:write')],
        },
        async () => {
          return { success: true };
        },
      );

      const response = await app.inject({
        method: 'GET',
        url: '/test',
        headers: {
          authorization: `Bearer ${token}`,
        },
      });

      expect(response.statusCode).toBe(200);
    });

    it('should allow request with multiple required scopes', async () => {
      const token = createMockJWT({ scopes: ['test:read', 'test:write', 'project:admin'] });

      app.get(
        '/test',
        {
          preHandler: [requireAuth, requireScopes('test:read', 'test:write')],
        },
        async () => {
          return { success: true };
        },
      );

      const response = await app.inject({
        method: 'GET',
        url: '/test',
        headers: {
          authorization: `Bearer ${token}`,
        },
      });

      expect(response.statusCode).toBe(200);
    });

    it('should block request missing required scope', async () => {
      const token = createMockJWT({ scopes: ['test:read'] });

      app.get(
        '/test',
        {
          preHandler: [requireAuth, requireScopes('test:write')],
        },
        async () => {
          return { success: true };
        },
      );

      const response = await app.inject({
        method: 'GET',
        url: '/test',
        headers: {
          authorization: `Bearer ${token}`,
        },
      });

      expect(response.statusCode).toBe(403);
      expect(response.json().error).toContain('Missing required scopes');
    });

    it('should block when user has only some required scopes', async () => {
      const token = createMockJWT({ scopes: ['test:read'] });

      app.get(
        '/test',
        {
          preHandler: [requireAuth, requireScopes('test:read', 'test:write')],
        },
        async () => {
          return { success: true };
        },
      );

      const response = await app.inject({
        method: 'GET',
        url: '/test',
        headers: {
          authorization: `Bearer ${token}`,
        },
      });

      expect(response.statusCode).toBe(403);
      const body = response.json();
      expect(body.missingScopes).toContain('test:write');
    });

    it('should handle empty scopes array in JWT', async () => {
      const token = createMockJWT({ scopes: [] });

      app.get(
        '/test',
        {
          preHandler: [requireAuth, requireScopes('test:read')],
        },
        async () => {
          return { success: true };
        },
      );

      const response = await app.inject({
        method: 'GET',
        url: '/test',
        headers: {
          authorization: `Bearer ${token}`,
        },
      });

      expect(response.statusCode).toBe(403);
    });
  });
});
