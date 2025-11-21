import { describe, it, expect, vi } from 'vitest';
import { bootstrap } from './worker.js';

// Capture options passed to circuit breaker
const captured: any[] = [];
vi.mock('./channels/circuitbreaker.js', () => ({
  createCircuitBreaker: vi.fn((opts) => {
    captured.push(opts);
    return { execute: (fn: any) => fn() };
  }),
}));

vi.mock('./config.js', () => ({
  loadConfig: vi.fn().mockReturnValue({
    env: 'dev',
    service: { name: 'notification-service', port: 7902, env: 'dev' },
    observability: {
      serviceName: 'notification-service',
      serviceVersion: '1.0.0',
      otlpEndpoint: 'http://localhost:4318',
      environment: 'dev',
    },
    nats: { url: 'nats://localhost:4222', subjects: { notifications: 'snapshots' } },
    redis: { url: 'redis://localhost:6379', maxRetriesPerRequest: 3 },
    postgres: {
      host: 'localhost',
      port: 5432,
      database: 'lma',
      user: 'postgres',
      password: 'postgres',
      maxConnections: 10,
    },
    smtp: {
      host: 'localhost',
      port: 1025,
      user: 'user',
      password: 'pass',
      from: 'noreply@example.com',
      secure: false,
    },
    vault: { addr: 'http://localhost:8200', roleId: 'role', secretId: 'secret' },
    auth: {},
    channels: {
      webhookUrl: 'https://example.com/hook',
      slackWebhookUrl: 'https://hooks.slack.com/xyz',
    },
    templates: { dir: 'src/templates' },
    queue: { batchSize: 10, tickMs: 200, defaultMaxAttempts: 3 },
    reliability: {
      timeoutMs: 5000,
      retry: { max: 1, baseMs: 1, jitterPct: 0 },
      breaker: { failureThreshold: 7, windowSize: 11, halfOpenAfterMs: 12345 },
    },
    features: { outboxEnqueueDedup: false },
    worker: {
      enabled: false,
      natsSubjects: ['snapshot.*', 'fixes.*', 'publish.*', 'checks.*', 'slo.*', 'errors.*'],
    },
  }),
}));

vi.mock('./events/nats-subscribe.js', () => ({
  createNatsSubscribe: vi
    .fn()
    .mockReturnValue(() => ({ once: vi.fn().mockResolvedValue({ processed: 0, failed: 0 }) })),
}));
vi.mock('./redis/client.js', () => ({
  createRedisClient: vi.fn().mockReturnValue({ quit: vi.fn() }),
}));
vi.mock('./notifications/queue.js', () => ({
  createNotificationQueue: vi
    .fn()
    .mockReturnValue({ dequeue: vi.fn(), ack: vi.fn(), nack: vi.fn() }),
}));
vi.mock(
  'nats',
  () => ({
    connect: vi.fn().mockResolvedValue({ subscribe: vi.fn().mockReturnValue({}), close: vi.fn() }),
  }),
  { virtual: true },
);
vi.mock('ioredis', () => ({ default: vi.fn() }), { virtual: true });
vi.mock('./channels/renderer.js', () => ({
  createHandlebarsRenderer: vi
    .fn()
    .mockReturnValue(vi.fn().mockResolvedValue({ subject: 's', html: '<p/>', text: 't' })),
}));
vi.mock('./templates/loader.js', () => ({ createTemplateLoader: vi.fn().mockReturnValue({}) }));
vi.mock('./templates/store.js', () => ({ createTemplateStore: vi.fn().mockReturnValue({}) }));
vi.mock('./bootstrap/runloop.js', () => ({
  createRunLoop: vi.fn().mockReturnValue({ start: vi.fn(), stop: vi.fn() }),
}));
vi.mock('./bootstrap/runtime.js', () => ({
  createRuntime: vi.fn().mockReturnValue({
    startOnce: vi.fn().mockResolvedValue({ processed: 0, failed: 0 }),
    router: { route: vi.fn().mockResolvedValue({ ok: true }) },
    processNextBatch: vi.fn().mockResolvedValue({ dispatched: 0 }),
  }),
  startNatsSubscriptions: vi.fn().mockResolvedValue({ stop: vi.fn().mockResolvedValue(undefined) }),
}));

describe('worker breaker wiring from config', () => {
  it('uses cfg.reliability.breaker values when creating circuit breaker', async () => {
    const ctl = await bootstrap();
    expect(captured.length).toBeGreaterThan(0);
    const opts = captured[0]!;
    expect(opts.failureThreshold).toBe(7);
    expect(opts.windowSize).toBe(11);
    expect(opts.halfOpenAfterMs).toBe(12345);
    await ctl.stop();
  });
});
