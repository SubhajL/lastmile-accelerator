import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import Fastify from 'fastify';
import type { FastifyInstance } from 'fastify';
import { registerErrorHandler } from '../../../middleware/error-handler.js';
import { AuthError, AuthorizationError, ValidationError } from '../../../types/auth.js';

describe('Error Handler', () => {
  let app: FastifyInstance;

  beforeEach(async () => {
    app = Fastify();
    registerErrorHandler(app);
  });

  afterEach(async () => {
    await app.close();
  });

  describe('errorHandler()', () => {
    it('should map AuthError to 401 Unauthorized', async () => {
      app.get('/test', async () => {
        throw new AuthError('Token is invalid');
      });

      const response = await app.inject({
        method: 'GET',
        url: '/test',
      });

      expect(response.statusCode).toBe(401);
      expect(response.json().error).toContain('Token is invalid');
    });

    it('should include WWW-Authenticate header for AuthError', async () => {
      app.get('/test', async () => {
        throw new AuthError('Token missing');
      });

      const response = await app.inject({
        method: 'GET',
        url: '/test',
      });

      expect(response.statusCode).toBe(401);
      expect(response.headers['www-authenticate']).toBe('Bearer');
    });

    it('should map AuthorizationError to 403 Forbidden', async () => {
      app.get('/test', async () => {
        throw new AuthorizationError('Insufficient permissions');
      });

      const response = await app.inject({
        method: 'GET',
        url: '/test',
      });

      expect(response.statusCode).toBe(403);
      expect(response.json().error).toContain('Insufficient permissions');
    });

    it('should include missing scopes in error body', async () => {
      app.get('/test', async () => {
        throw new AuthorizationError(
          'Missing required scopes',
          ['test:write', 'project:admin']
        );
      });

      const response = await app.inject({
        method: 'GET',
        url: '/test',
      });

      expect(response.statusCode).toBe(403);
      const body = response.json();
      expect(body.missingScopes).toEqual(['test:write', 'project:admin']);
    });

    it('should map ValidationError to 400 Bad Request', async () => {
      app.get('/test', async () => {
        throw new ValidationError('Invalid UUID format', 'projectId');
      });

      const response = await app.inject({
        method: 'GET',
        url: '/test',
      });

      expect(response.statusCode).toBe(400);
      const body = response.json();
      expect(body.error).toContain('Invalid UUID');
      expect(body.field).toBe('projectId');
    });

    it('should map generic Error to 500 Internal Server Error', async () => {
      app.get('/test', async () => {
        throw new Error('Unexpected error');
      });

      const response = await app.inject({
        method: 'GET',
        url: '/test',
      });

      expect(response.statusCode).toBe(500);
      expect(response.json().error).toBe('Internal Server Error');
    });

    it('should sanitize error message for generic errors', async () => {
      app.get('/test', async () => {
        throw new Error('Database connection failed with password: secret123');
      });

      const response = await app.inject({
        method: 'GET',
        url: '/test',
      });

      expect(response.statusCode).toBe(500);
      // Generic errors shouldn't leak internal details
      expect(response.json().error).toBe('Internal Server Error');
    });

    it('should handle errors without message gracefully', async () => {
      app.get('/test', async () => {
        throw new Error();
      });

      const response = await app.inject({
        method: 'GET',
        url: '/test',
      });

      expect(response.statusCode).toBe(500);
      expect(response.json().error).toBe('Internal Server Error');
    });
  });
});
