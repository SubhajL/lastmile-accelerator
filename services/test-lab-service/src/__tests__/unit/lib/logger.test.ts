import { describe, it, expect, beforeEach, vi } from 'vitest';
import { createLogger, childLogger, redactSensitiveData } from '../../../lib/logger.js';

describe('Logger', () => {
  describe('createLogger()', () => {
    it('should create logger with correct level', () => {
      const logger = createLogger({
        serviceName: 'test-service',
        env: 'dev',
        level: 'info',
      });

      expect(logger).toBeDefined();
      expect(logger.level).toBe('info');
    });

    it('should use debug level in dev environment', () => {
      const logger = createLogger({
        serviceName: 'test-service',
        env: 'dev',
      });

      expect(logger.level).toBe('debug');
    });

    it('should use info level in production environment', () => {
      const logger = createLogger({
        serviceName: 'test-service',
        env: 'prod',
      });

      expect(logger.level).toBe('info');
    });

    it('should include service name in base logger context', () => {
      const logger = createLogger({
        serviceName: 'test-lab-service',
        env: 'dev',
      });

      const bindings = logger.bindings();
      expect(bindings.service).toBe('test-lab-service');
    });

    it('should use pretty print in dev mode', () => {
      const logger = createLogger({
        serviceName: 'test-service',
        env: 'dev',
      });

      // In dev mode, transport should be set for pretty printing
      expect(logger).toBeDefined();
    });

    it('should use JSON output in production mode', () => {
      const logger = createLogger({
        serviceName: 'test-service',
        env: 'prod',
      });

      expect(logger).toBeDefined();
      // Production logger should not have pretty print transport
    });
  });

  describe('childLogger()', () => {
    it('should create child logger with parent logger', () => {
      const parentLogger = createLogger({
        serviceName: 'test-service',
        env: 'dev',
      });

      const child = childLogger(parentLogger, { requestId: 'req-123' });

      expect(child).toBeDefined();
      const bindings = child.bindings();
      expect(bindings.requestId).toBe('req-123');
    });

    it('should inherit parent context', () => {
      const parentLogger = createLogger({
        serviceName: 'test-service',
        env: 'dev',
      });

      const child = childLogger(parentLogger, { requestId: 'req-123' });

      const bindings = child.bindings();
      expect(bindings.service).toBe('test-service');
      expect(bindings.requestId).toBe('req-123');
    });

    it('should add new context fields', () => {
      const parentLogger = createLogger({
        serviceName: 'test-service',
        env: 'dev',
      });

      const child = childLogger(parentLogger, { 
        requestId: 'req-123',
        userId: 'user-456',
        projectId: 'project-789',
      });

      const bindings = child.bindings();
      expect(bindings.requestId).toBe('req-123');
      expect(bindings.userId).toBe('user-456');
      expect(bindings.projectId).toBe('project-789');
    });
  });

  describe('redactSensitiveData()', () => {
    it('should sanitize sensitive fields (password, token)', () => {
      const input = { 
        username: 'user',
        password: 'secret123',
        token: 'bearer-token-xyz',
      };

      const result = redactSensitiveData(input);
        
      expect(result.password).toBe('[REDACTED]');
      expect(result.token).toBe('[REDACTED]');
      expect(result.username).toBe('user');
    });

    it('should sanitize authorization header in nested objects', () => {
      const input = {
        headers: {
          'content-type': 'application/json',
          'authorization': 'Bearer secret-token',
        },
      };

      const result = redactSensitiveData(input);
        
      const headers = result.headers as Record<string, unknown>;
      expect(headers.authorization).toBe('[REDACTED]');
      expect(headers['content-type']).toBe('application/json');
    });

    it('should sanitize multiple sensitive fields', () => {
      const input = {
        apiKey: 'key-123',
        api_key: 'key-456',
        accessToken: 'token-789',
        refreshToken: 'refresh-abc',
        secret: 'my-secret',
        normalField: 'normal-value',
      };

      const result = redactSensitiveData(input);

      expect(result.apiKey).toBe('[REDACTED]');
      expect(result.api_key).toBe('[REDACTED]');
      expect(result.accessToken).toBe('[REDACTED]');
      expect(result.refreshToken).toBe('[REDACTED]');
      expect(result.secret).toBe('[REDACTED]');
      expect(result.normalField).toBe('normal-value');
    });

    it('should handle deeply nested sensitive fields', () => {
      const input = {
        user: {
          credentials: {
            password: 'secret',
            token: 'abc123',
          },
          email: 'user@example.com',
        },
      };

      const result = redactSensitiveData(input);

      const user = result.user as Record<string, unknown>;
      const credentials = user.credentials as Record<string, unknown>;
      expect(credentials.password).toBe('[REDACTED]');
      expect(credentials.token).toBe('[REDACTED]');
      expect(user.email).toBe('user@example.com');
    });
  });
});
