import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { createApp } from './app.js';
import type { Config } from './types.js';
import type { Pool } from 'pg';

describe('app', () => {
  let mockConfig: Config;
  let mockDbPool: Pool;

  beforeEach(() => {
    mockConfig = {
      env: 'dev',
      service: {
        name: 'notification-service',
        port: 7902,
        env: 'dev'
      },
      observability: {
        serviceName: 'notification-service',
        serviceVersion: '1.0.0',
        otlpEndpoint: 'http://localhost:4318',
        environment: 'dev'
      },
      nats: { url: 'nats://localhost:4222' },
      redis: { url: 'redis://localhost:6379', maxRetriesPerRequest: 3 },
      postgres: {
        host: 'localhost',
        port: 5432,
        database: 'test',
        user: 'test',
        password: 'test',
        maxConnections: 10
      },
      smtp: {
        host: 'localhost',
        port: 1025,
        user: 'test',
        password: 'test',
        from: 'test@test.com',
        secure: false
      },
      vault: {
        addr: 'http://localhost:8200',
        roleId: 'test',
        secretId: 'test'
      },
      auth: {
        jwtPublicKey: 'test-key'
      },
      channels: {}
    };

    mockDbPool = {
      query: vi.fn().mockResolvedValue({ rows: [{ result: 1 }] })
    } as unknown as Pool;
  });

  afterEach(async () => {
    // Cleanup
  });

  describe('createApp', () => {
    it('should create Fastify instance', async () => {
      const app = await createApp(mockConfig, mockDbPool);

      expect(app).toBeDefined();
      expect(typeof app.listen).toBe('function');

      await app.close();
    });

    it('should register CORS plugin', async () => {
      const app = await createApp(mockConfig, mockDbPool);

      const response = await app.inject({
        method: 'OPTIONS',
        url: '/healthz',
        headers: {
          origin: 'http://localhost:3000'
        }
      });

      expect(response.headers['access-control-allow-origin']).toBeDefined();

      await app.close();
    });

    it('should register helmet security headers', async () => {
      const app = await createApp(mockConfig, mockDbPool);

      const response = await app.inject({
        method: 'GET',
        url: '/healthz'
      });

      expect(response.headers['x-frame-options']).toBeDefined();

      await app.close();
    });

    it('should register /healthz endpoint', async () => {
      const app = await createApp(mockConfig, mockDbPool);

      const response = await app.inject({
        method: 'GET',
        url: '/healthz'
      });

      expect(response.statusCode).toBe(200);
      const body = JSON.parse(response.body);
      expect(body.status).toBe('ok');
      expect(body.timestamp).toBeDefined();

      await app.close();
    });

    it('should return 200 on /healthz when database healthy', async () => {
      const healthyPool = {
        query: vi.fn().mockResolvedValue({ rows: [{ result: 1 }] })
      } as unknown as Pool;

      const app = await createApp(mockConfig, healthyPool);

      const response = await app.inject({
        method: 'GET',
        url: '/healthz'
      });

      expect(response.statusCode).toBe(200);

      await app.close();
    });

    it('should return 503 on /healthz when database unhealthy', async () => {
      const unhealthyPool = {
        query: vi.fn().mockRejectedValue(new Error('DB Error'))
      } as unknown as Pool;

      const app = await createApp(mockConfig, unhealthyPool);

      const response = await app.inject({
        method: 'GET',
        url: '/healthz'
      });

      expect(response.statusCode).toBe(503);

      await app.close();
    });

    it('should register /metrics endpoint', async () => {
      const app = await createApp(mockConfig, mockDbPool);

      const response = await app.inject({
        method: 'GET',
        url: '/metrics'
      });

      expect(response.statusCode).toBe(200);
      expect(response.headers['content-type']).toContain('text/plain');

      await app.close();
    });
  });
});
