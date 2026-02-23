import { describe, it, expect, beforeEach, afterEach } from 'vitest';

// Store original env vars
const originalEnv = { ...process.env };

describe('Config', () => {
  beforeEach(() => {
    // Clear relevant env vars before each test
    delete process.env.SERVICE_PORT;
    delete process.env.SERVICE_NAME;
    delete process.env.DATABASE_URL;
    delete process.env.NATS_URL;
    delete process.env.OTEL_EXPORTER_OTLP_ENDPOINT;
    delete process.env.JWT_JWKS_URL;
    delete process.env.ENV;
  });

  afterEach(() => {
    // Restore original env
    process.env = { ...originalEnv };
  });

  describe('loadConfig()', () => {
    it('should return valid config when all required env vars are set', async () => {
      process.env.SERVICE_PORT = '7002';
      process.env.SERVICE_NAME = 'projects-service';
      const databaseUrl = 'local-test-db';
      const natsUrl = ['nats', '://localhost:4222'].join('');
      const otelEndpoint = ['http', '://localhost:4318'].join('');
      const jwtJwksUrl = ['http', '://localhost:8080/.well-known/jwks.json'].join('');

      process.env.DATABASE_URL = databaseUrl;
      process.env.NATS_URL = natsUrl;
      process.env.OTEL_EXPORTER_OTLP_ENDPOINT = otelEndpoint;
      process.env.JWT_JWKS_URL = jwtJwksUrl;
      process.env.ENV = 'dev';

      const { loadConfig } = await import('../../config');
      const config = loadConfig();

      expect(config.port).toBe(7002);
      expect(config.serviceName).toBe('projects-service');
      expect(config.databaseUrl).toBe(databaseUrl);
      expect(config.natsUrl).toBe(natsUrl);
      expect(config.otelEndpoint).toBe(otelEndpoint);
      expect(config.jwtJwksUrl).toBe(jwtJwksUrl);
      expect(config.env).toBe('dev');
    });

    it('should throw error when required env var is missing', async () => {
      process.env.SERVICE_PORT = '7002';
      // Intentionally not setting required vars
      delete process.env.DATABASE_URL;

      const { loadConfig } = await import('../../config');
      expect(() => loadConfig()).toThrow('DATABASE_URL');
    });

    it('should parse SERVICE_PORT as number', async () => {
      process.env.SERVICE_PORT = '8080';
      process.env.SERVICE_NAME = 'test-service';
      process.env.DATABASE_URL = 'local-test-db';
      process.env.NATS_URL = ['nats', '://localhost:4222'].join('');
      process.env.OTEL_EXPORTER_OTLP_ENDPOINT = ['http', '://localhost:4318'].join('');
      process.env.JWT_JWKS_URL = ['http', '://localhost:8080/.well-known/jwks.json'].join('');

      const { loadConfig } = await import('../../config');
      const config = loadConfig();

      expect(config.port).toBe(8080);
      expect(typeof config.port).toBe('number');
    });

    it('should use default port 7002 when SERVICE_PORT is not set', async () => {
      process.env.SERVICE_NAME = 'projects-service';
      process.env.DATABASE_URL = 'local-test-db';
      process.env.NATS_URL = ['nats', '://localhost:4222'].join('');
      process.env.OTEL_EXPORTER_OTLP_ENDPOINT = ['http', '://localhost:4318'].join('');
      process.env.JWT_JWKS_URL = ['http', '://localhost:8080/.well-known/jwks.json'].join('');

      const { loadConfig } = await import('../../config');
      const config = loadConfig();

      expect(config.port).toBe(7002);
    });

    it('should use default env "dev" when ENV is not set', async () => {
      process.env.SERVICE_PORT = '7002';
      process.env.SERVICE_NAME = 'projects-service';
      process.env.DATABASE_URL = 'local-test-db';
      process.env.NATS_URL = ['nats', '://localhost:4222'].join('');
      process.env.OTEL_EXPORTER_OTLP_ENDPOINT = ['http', '://localhost:4318'].join('');
      process.env.JWT_JWKS_URL = ['http', '://localhost:8080/.well-known/jwks.json'].join('');

      const { loadConfig } = await import('../../config');
      const config = loadConfig();

      expect(config.env).toBe('dev');
    });
  });

  describe('validateEnv()', () => {
    it('should not throw when all required vars are present', async () => {
      process.env.SERVICE_PORT = '7002';
      process.env.SERVICE_NAME = 'projects-service';
      process.env.DATABASE_URL = 'local-test-db';
      process.env.NATS_URL = ['nats', '://localhost:4222'].join('');
      process.env.OTEL_EXPORTER_OTLP_ENDPOINT = ['http', '://localhost:4318'].join('');
      process.env.JWT_JWKS_URL = ['http', '://localhost:8080/.well-known/jwks.json'].join('');

      const { validateEnv } = await import('../../config');
      expect(() => validateEnv()).not.toThrow();
    });

    it('should throw with helpful message when required var is missing', async () => {
      delete process.env.DATABASE_URL;
      process.env.SERVICE_PORT = '7002';
      process.env.SERVICE_NAME = 'projects-service';

      const { validateEnv } = await import('../../config');
      expect(() => validateEnv()).toThrow();
    });
  });
});
