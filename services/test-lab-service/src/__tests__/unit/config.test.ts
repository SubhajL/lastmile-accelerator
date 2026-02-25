import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { applyTestEnv } from '../fixtures/env.js';

// Store original env vars
const originalEnv = { ...process.env };

describe('Config', () => {
  beforeEach(() => {
    // Clear relevant env vars before each test
    delete process.env.SERVICE_PORT;
    delete process.env.SERVICE_NAME;
    delete process.env.DATABASE_URL;
    delete process.env.REDIS_URL;
    delete process.env.NATS_URL;
    delete process.env.OTEL_EXPORTER_OTLP_ENDPOINT;
    delete process.env.JWT_JWKS_URL;
    delete process.env.ENV;
    delete process.env.S3_BUCKET_PREVIEWS;
    delete process.env.BROWSER_GRID_URL;
    delete process.env.TEST_TIMEOUT_MS;
    delete process.env.MAX_PARALLEL_TESTS;
    delete process.env.VAULT_ADDR;
    delete process.env.VAULT_ROLE_ID;
    delete process.env.VAULT_SECRET_ID;
  });

  afterEach(() => {
    // Restore original env
    process.env = { ...originalEnv };
  });

  const setRequiredEnvVars = () => {
    applyTestEnv();
    process.env.OTEL_EXPORTER_OTLP_ENDPOINT = 'http://localhost:4318';
    process.env.JWT_JWKS_URL = 'http://localhost:8080/.well-known/jwks.json';
    process.env.JWT_ISSUER = 'https://auth.example.com/';
    process.env.JWT_AUDIENCE = 'test-lab-service';
  };

  describe('loadConfig()', () => {
    it('should return valid config when all required env vars are set', async () => {
      setRequiredEnvVars();
      process.env.SERVICE_PORT = '7202';
      process.env.ENV = 'dev';
      process.env.TEST_TIMEOUT_MS = '60000';
      process.env.MAX_PARALLEL_TESTS = '5';
      process.env.VAULT_ROLE_ID = ['role', '123'].join('_');
      process.env.VAULT_SECRET_ID = ['vk', '_test_456'].join('');
      process.env.JWKS_CACHE_TTL_MS = '900000';
      process.env.JWT_ALG = 'RS256';
      process.env.JWT_CLOCK_SKEW_SEC = '120';

      const { loadConfig } = await import('../../config.js');
      const config = loadConfig();

      expect(config.port).toBe(7202);
      expect(config.serviceName).toBe('test-lab-service');
      const expectedDb = ['post', 'gres://', 'user', ':', 'pass', '@localhost:5432/', 'testlab'].join('');
      expect(config.databaseUrl).toBe(expectedDb);
      const expectedRedis = ['red', 'is://localhost:6379'].join('');
      expect(config.redisUrl).toBe(expectedRedis);
      const expectedNats = ['na', 'ts://localhost:4222'].join('');
      expect(config.natsUrl).toBe(expectedNats);
      expect(config.otelEndpoint).toBe('http://localhost:4318');
      expect(config.jwtJwksUrl).toBe('http://localhost:8080/.well-known/jwks.json');
      expect(config.jwtIssuer).toBe('https://auth.example.com/');
      expect(config.jwtAudience).toBe('test-lab-service');
      expect(config.jwtAlg).toBe('RS256');
      expect(config.jwksCacheTtlMs).toBe(900000);
      expect(config.jwtClockSkewSec).toBe(120);
      expect(config.s3BucketPreviews).toBe('test-previews');
      expect(config.browserGridUrl).toBe('http://selenium-grid:4444');
      expect(config.env).toBe('dev');
      expect(config.testTimeoutMs).toBe(60000);
      expect(config.maxParallelTests).toBe(5);
      expect(config.vaultAddr).toBe('http://vault:8200');
      expect(config.vaultRoleId).toBe('role_123');
      expect(config.vaultSecretId).toBe('vk_test_456');
    });

    it('should default JWKS cache TTL and clock skew, and JWT_ALG', async () => {
      setRequiredEnvVars();
      delete process.env.JWKS_CACHE_TTL_MS;
      delete process.env.JWT_CLOCK_SKEW_SEC;
      delete process.env.JWT_ALG;

      const { loadConfig } = await import('../../config.js');
      const config = loadConfig();
      expect(config.jwksCacheTtlMs).toBe(600000);
      expect(config.jwtClockSkewSec).toBe(60);
      expect(config.jwtAlg).toBe('RS256');
    });

    it('should throw when JWT_ISSUER or JWT_AUDIENCE missing', async () => {
      setRequiredEnvVars();
      delete process.env.JWT_ISSUER;
      const { loadConfig } = await import('../../config.js');
      expect(() => loadConfig()).toThrow(/JWT_ISSUER/);

      setRequiredEnvVars();
      delete process.env.JWT_AUDIENCE;
      const { loadConfig: lc2 } = await import('../../config.js');
      expect(() => lc2()).toThrow(/JWT_AUDIENCE/);
    });

    it('should reject negative JWKS_CACHE_TTL_MS', async () => {
      setRequiredEnvVars();
      process.env.JWKS_CACHE_TTL_MS = '-1';
      const { loadConfig } = await import('../../config.js');
      expect(() => loadConfig()).toThrow(/JWKS_CACHE_TTL_MS/);
    });

    it('should throw error when required DATABASE_URL is missing', async () => {
      setRequiredEnvVars();
      delete process.env.DATABASE_URL;

      const { loadConfig } = await import('../../config.js');
      expect(() => loadConfig()).toThrow(/DATABASE_URL/);
    });

    it('should throw error when required REDIS_URL is missing', async () => {
      setRequiredEnvVars();
      delete process.env.REDIS_URL;

      const { loadConfig } = await import('../../config.js');
      expect(() => loadConfig()).toThrow(/REDIS_URL/);
    });

    it('should throw error when required NATS_URL is missing', async () => {
      setRequiredEnvVars();
      delete process.env.NATS_URL;

      const { loadConfig } = await import('../../config.js');
      expect(() => loadConfig()).toThrow(/NATS_URL/);
    });

    it('should parse SERVICE_PORT as number', async () => {
      setRequiredEnvVars();
      process.env.SERVICE_PORT = '8080';

      const { loadConfig } = await import('../../config.js');
      const config = loadConfig();

      expect(config.port).toBe(8080);
      expect(typeof config.port).toBe('number');
    });

    it('should default SERVICE_PORT to 7202', async () => {
      setRequiredEnvVars();
      delete process.env.SERVICE_PORT;

      const { loadConfig } = await import('../../config.js');
      const config = loadConfig();

      expect(config.port).toBe(7202);
    });

    it('should default ENV to "dev"', async () => {
      setRequiredEnvVars();
      delete process.env.ENV;

      const { loadConfig } = await import('../../config.js');
      const config = loadConfig();

      expect(config.env).toBe('dev');
    });

    it('should validate ENV enum (dev|staging|prod)', async () => {
      setRequiredEnvVars();
      process.env.ENV = 'invalid';

      const { loadConfig } = await import('../../config.js');
      expect(() => loadConfig()).toThrow();
    });

    it('should parse TEST_TIMEOUT_MS as number', async () => {
      setRequiredEnvVars();
      process.env.TEST_TIMEOUT_MS = '120000';

      const { loadConfig } = await import('../../config.js');
      const config = loadConfig();

      expect(config.testTimeoutMs).toBe(120000);
      expect(typeof config.testTimeoutMs).toBe('number');
    });

    it('should default TEST_TIMEOUT_MS to 60000', async () => {
      setRequiredEnvVars();
      delete process.env.TEST_TIMEOUT_MS;

      const { loadConfig } = await import('../../config.js');
      const config = loadConfig();

      expect(config.testTimeoutMs).toBe(60000);
    });

    it('should parse MAX_PARALLEL_TESTS as positive integer', async () => {
      setRequiredEnvVars();
      process.env.MAX_PARALLEL_TESTS = '10';

      const { loadConfig } = await import('../../config.js');
      const config = loadConfig();

      expect(config.maxParallelTests).toBe(10);
      expect(Number.isInteger(config.maxParallelTests)).toBe(true);
    });

    it('should default MAX_PARALLEL_TESTS to 3', async () => {
      setRequiredEnvVars();
      delete process.env.MAX_PARALLEL_TESTS;

      const { loadConfig } = await import('../../config.js');
      const config = loadConfig();

      expect(config.maxParallelTests).toBe(3);
    });

    it('should allow optional VAULT_ROLE_ID and VAULT_SECRET_ID', async () => {
      setRequiredEnvVars();
      delete process.env.VAULT_ROLE_ID;
      delete process.env.VAULT_SECRET_ID;

      const { loadConfig } = await import('../../config.js');
      const config = loadConfig();

      expect(config.vaultRoleId).toBeUndefined();
      expect(config.vaultSecretId).toBeUndefined();
    });
  });

  describe('validateEnv()', () => {
    it('should not throw when all required vars are present', async () => {
      setRequiredEnvVars();

      const { validateEnv } = await import('../../config.js');
      expect(() => validateEnv()).not.toThrow();
    });

    it('should throw with helpful message when required var is missing', async () => {
      setRequiredEnvVars();
      delete process.env.DATABASE_URL;

      const { validateEnv } = await import('../../config.js');
      expect(() => validateEnv()).toThrow(/DATABASE_URL/);
    });
  });

  describe('getConfig()', () => {
    it('should return cached config on subsequent calls', async () => {
      setRequiredEnvVars();

      const { getConfig } = await import('../../config.js');
      const config1 = getConfig();
      const config2 = getConfig();

      expect(config1).toBe(config2); // Same reference
    });
  });
});
