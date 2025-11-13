import { describe, it, expect, vi } from 'vitest';
import { bootstrap } from './worker.js';

vi.mock('./config.js', () => ({
  loadConfig: vi.fn().mockReturnValue({
    env: 'dev',
    service: { name: 'notification-service', port: 7902, env: 'dev' },
    observability: { serviceName: 'notification-service', serviceVersion: '1.0.0', otlpEndpoint: 'http://localhost:4318', environment: 'dev' },
    nats: { url: 'nats://localhost:4222', subjects: { notifications: 'snapshots' } },
    redis: { url: 'redis://localhost:6379', maxRetriesPerRequest: 3 },
    postgres: { host: 'localhost', port: 5432, database: 'lma', user: 'postgres', password: 'postgres', maxConnections: 10 },
    smtp: { host: 'localhost', port: 1025, user: 'user', password: 'pass', from: 'noreply@example.com', secure: false },
    vault: { addr: 'http://localhost:8200', roleId: 'role', secretId: 'secret' },
    auth: {},
    channels: {},
    templates: { dir: 'src/templates' },
    queue: { batchSize: 10, tickMs: 200, defaultMaxAttempts: 3 },
    reliability: { timeoutMs: 5000, retry: { max: 2, baseMs: 10, jitterPct: 0 }, breaker: { failureThreshold: 3, windowSize: 10, halfOpenAfterMs: 15000 } }
  })
}));

vi.mock('./bootstrap/runtime.js', () => ({
  createRuntime: vi.fn().mockImplementation(({ subjects, batchSize }) => ({
    startOnce: vi.fn().mockResolvedValue({ processed: 0, failed: 0 })
  }))
}));

vi.mock('./bootstrap/runloop.js', () => ({
  createRunLoop: vi.fn().mockReturnValue({ start: vi.fn(), stop: vi.fn() })
}));

vi.mock('./events/nats-subscribe.js', () => ({
  createNatsSubscribe: vi.fn().mockReturnValue(() => ({ once: vi.fn().mockResolvedValue({ processed: 0, failed: 0 }) }))
}));

vi.mock('./channels/renderer.js', () => ({
  createHandlebarsRenderer: vi.fn().mockReturnValue(vi.fn().mockResolvedValue({ subject: 'x', html: 'y', text: 'y' }))
}));

vi.mock('./channels/registry.js', () => ({
  createDefaultChannelRegistry: vi.fn().mockReturnValue({ get: vi.fn().mockReturnValue({ send: vi.fn() }) }),
  createChannelRegistry: vi.fn().mockReturnValue({ get: vi.fn() })
}));

vi.mock('./redis/client.js', () => ({
  createRedisClient: vi.fn().mockReturnValue({ on: vi.fn(), quit: vi.fn() })
}));

vi.mock('./notifications/queue.js', () => ({
  createNotificationQueue: vi.fn().mockReturnValue({
    enqueue: vi.fn(),
    dequeue: vi.fn(),
    ack: vi.fn(),
    nack: vi.fn()
  })
}));

vi.mock('nats', () => ({
  connect: vi.fn().mockResolvedValue({ subscribe: vi.fn().mockReturnValue({}), close: vi.fn() })
}), { virtual: true });

vi.mock('ioredis', () => ({ default: vi.fn() }), { virtual: true });

describe('worker/bootstrap', () => {
  it('bootstraps and starts run loop without throwing', async () => {
    const ctl = await bootstrap();
    expect(ctl).toBeDefined();
    expect(typeof ctl.stop).toBe('function');
    await ctl.stop();
  });
});
