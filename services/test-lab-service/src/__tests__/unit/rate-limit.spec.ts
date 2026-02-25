import { describe, expect, test, beforeEach, vi } from 'vitest';
import Fastify from 'fastify';
import type { FastifyInstance } from 'fastify';
import type { RedisWrapper } from '../../clients/redis.js';
import { createRateLimitMiddleware } from '../../middleware/rate-limit.js';

describe('createRateLimitMiddleware', () => {
  let app: FastifyInstance;
  let mockRedis: RedisWrapper;
  let requestCounter: Map<string, number>;

  beforeEach(async () => {
    app = Fastify();
    requestCounter = new Map();

    mockRedis = {
      get: vi.fn(async (key: string) => {
        return String(requestCounter.get(key) ?? 0);
      }),
      set: vi.fn(async () => {}),
      del: vi.fn(async () => {}),
      incr: vi.fn(async (key: string) => {
        const current = requestCounter.get(key) ?? 0;
        const newValue = current + 1;
        requestCounter.set(key, newValue);
        return newValue;
      }),
      incrWithExpire: vi.fn(async (key: string) => {
        const current = requestCounter.get(key) ?? 0;
        const newValue = current + 1;
        requestCounter.set(key, newValue);
        return newValue;
      }),
      expire: vi.fn(async () => {}),
      close: vi.fn(async () => {}),
    };
  });

  test('allows requests below limit threshold', async () => {
    const rateLimitConfig = {
      requestsPerMinute: 10,
      windowMs: 60000,
      exemptRoutes: ['/healthz'],
      keyPrefix: 'rate-limit:',
    };

    app.addHook('preHandler', createRateLimitMiddleware(mockRedis, rateLimitConfig));

    app.get('/api/test', async () => ({ success: true }));

    // Make 9 requests (below limit of 10)
    for (let i = 0; i < 9; i++) {
      const response = await app.inject({
        method: 'GET',
        url: '/api/test',
        remoteAddress: '192.168.1.100',
      });

      expect(response.statusCode).toBe(200);
      expect(response.json()).toEqual({ success: true });
    }
  });

  test('blocks 11th request with 429 status when limit is 10', async () => {
    const rateLimitConfig = {
      requestsPerMinute: 10,
      windowMs: 60000,
      exemptRoutes: [],
      keyPrefix: 'rate-limit:',
    };

    app.addHook('preHandler', createRateLimitMiddleware(mockRedis, rateLimitConfig));

    app.get('/api/test', async () => ({ success: true }));

    // Make 10 requests (at limit)
    for (let i = 0; i < 10; i++) {
      const response = await app.inject({
        method: 'GET',
        url: '/api/test',
        remoteAddress: '192.168.1.100',
      });
      expect(response.statusCode).toBe(200);
    }

    // 11th request should be rate limited
    const blockedResponse = await app.inject({
      method: 'GET',
      url: '/api/test',
      remoteAddress: '192.168.1.100',
    });

    expect(blockedResponse.statusCode).toBe(429);
    expect(blockedResponse.json()).toMatchObject({
      error: expect.objectContaining({
        code: 'RATE_LIMIT_EXCEEDED',
        message: expect.stringContaining('Too many requests'),
      }),
    });
  });

  test('includes Retry-After header in 429 response', async () => {
    const rateLimitConfig = {
      requestsPerMinute: 5,
      windowMs: 60000,
      exemptRoutes: [],
      keyPrefix: 'rate-limit:',
    };

    app.addHook('preHandler', createRateLimitMiddleware(mockRedis, rateLimitConfig));

    app.get('/api/test', async () => ({ success: true }));

    // Exceed limit
    for (let i = 0; i < 6; i++) {
      await app.inject({
        method: 'GET',
        url: '/api/test',
        remoteAddress: '192.168.1.100',
      });
    }

    const response = await app.inject({
      method: 'GET',
      url: '/api/test',
      remoteAddress: '192.168.1.100',
    });

    expect(response.statusCode).toBe(429);
    expect(response.headers['retry-after']).toBe('60');
  });

  test('isolates limits per IP address (different IPs have separate quotas)', async () => {
    const rateLimitConfig = {
      requestsPerMinute: 3,
      windowMs: 60000,
      exemptRoutes: [],
      keyPrefix: 'rate-limit:',
    };

    app.addHook('preHandler', createRateLimitMiddleware(mockRedis, rateLimitConfig));

    app.get('/api/test', async () => ({ success: true }));

    // IP A makes 3 requests (at limit)
    for (let i = 0; i < 3; i++) {
      const response = await app.inject({
        method: 'GET',
        url: '/api/test',
        remoteAddress: '192.168.1.1',
      });
      expect(response.statusCode).toBe(200);
    }

    // IP B should have separate quota
    const ipBResponse = await app.inject({
      method: 'GET',
      url: '/api/test',
      remoteAddress: '192.168.1.2',
    });

    // IP B should not be rate limited (different IP)
    expect(ipBResponse.statusCode).toBe(200);
  });

  test('skips exempt routes like /healthz', async () => {
    const rateLimitConfig = {
      requestsPerMinute: 2,
      windowMs: 60000,
      exemptRoutes: ['/healthz', '/metrics'],
      keyPrefix: 'rate-limit:',
    };

    app.addHook('preHandler', createRateLimitMiddleware(mockRedis, rateLimitConfig));

    app.get('/healthz', async () => 'ok');

    // Make 10 requests to /healthz (should never be rate limited)
    for (let i = 0; i < 10; i++) {
      const response = await app.inject({
        method: 'GET',
        url: '/healthz',
      });

      expect(response.statusCode).toBe(200);
      expect(response.body).toBe('ok');
    }

    // Redis incr should never be called for exempt routes
    expect(mockRedis.incrWithExpire).not.toHaveBeenCalled();
  });

  test('falls back to IP when no tenant claim in JWT', async () => {
    const rateLimitConfig = {
      requestsPerMinute: 5,
      windowMs: 60000,
      exemptRoutes: [],
      keyPrefix: 'rate-limit:',
    };

    app.addHook('preHandler', createRateLimitMiddleware(mockRedis, rateLimitConfig));

    app.get('/api/test', async () => ({ success: true }));

    // Make requests without user (no JWT)
    for (let i = 0; i < 3; i++) {
      const response = await app.inject({
        method: 'GET',
        url: '/api/test',
        remoteAddress: '203.0.113.42',
      });

      expect(response.statusCode).toBe(200);
    }

    // Verify rate limit key uses IP address
    expect(mockRedis.incrWithExpire).toHaveBeenCalledWith(expect.stringContaining('203.0.113.42'), expect.any(Number));
  });

  test('uses JWT tenantId when user context is present', async () => {
    const rateLimitConfig = {
      requestsPerMinute: 5,
      windowMs: 60000,
      exemptRoutes: [],
      keyPrefix: 'rate-limit:',
    };

    app.addHook('preHandler', async (req) => {
      req.user = { sub: 'u-1', tenantId: 'tenant-123', userId: 'u-1', scopes: [] };
    });
    app.addHook('preHandler', createRateLimitMiddleware(mockRedis, rateLimitConfig));
    app.get('/api/test', async () => ({ success: true }));

    const response = await app.inject({
      method: 'GET',
      url: '/api/test',
      remoteAddress: '203.0.113.42',
    });

    expect(response.statusCode).toBe(200);
    expect(mockRedis.incrWithExpire).toHaveBeenCalledWith(expect.stringContaining('tenant-123'), expect.any(Number));
  });

  test('allows requests when Redis fails (fail-open)', async () => {
    const faultyRedis: RedisWrapper = {
      get: vi.fn(async () => {
        throw new Error('Redis connection timeout');
      }),
      set: vi.fn(async () => {}),
      del: vi.fn(async () => {}),
      incr: vi.fn(async () => {
        throw new Error('Redis connection timeout');
      }),
      incrWithExpire: vi.fn(async () => {
        throw new Error('Redis connection timeout');
      }),
      expire: vi.fn(async () => {}),
      close: vi.fn(async () => {}),
    };

    const rateLimitConfig = {
      requestsPerMinute: 10,
      windowMs: 60000,
      exemptRoutes: [],
      keyPrefix: 'rate-limit:',
    };

    app.addHook('preHandler', createRateLimitMiddleware(faultyRedis, rateLimitConfig));

    app.get('/api/test', async () => ({ success: true }));

    // Request should succeed despite Redis failure (fail-open)
    const response = await app.inject({
      method: 'GET',
      url: '/api/test',
      remoteAddress: '192.168.1.100',
    });

    expect(response.statusCode).toBe(200);
    expect(response.json()).toEqual({ success: true });
  });
});
