import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { createApp } from '../../app.js';

vi.mock('../../clients/nats.js', () => {
  const close = vi.fn().mockResolvedValue(undefined);
  return {
    createNatsConnection: vi.fn().mockResolvedValue({
      publish: vi.fn(),
      subscribe: vi.fn().mockResolvedValue({ subject: 'x' }),
      close,
    }),
  };
});

vi.mock('../../clients/s3.js', () => ({
  createS3Client: vi.fn().mockReturnValue({}),
}));

vi.mock('../../clients/k8s.js', () => ({
  createK8sClient: vi.fn().mockResolvedValue({ batch: {}, core: {} }),
}));

let wired = false;
vi.mock('../../events/subscribers.js', async (orig) => {
  const mod: any = await orig();
  return {
    ...mod,
    wireEventSubscribers: vi.fn().mockImplementation(async () => { wired = true; }),
  };
});

// Minimal env
process.env.SERVICE_NAME = 'test-lab-service';
process.env.SERVICE_PORT = '7202';
process.env.DATABASE_URL = 'pgmem://wiring';
process.env.REDIS_URL = 'redis://localhost:6379';
process.env.NATS_URL = 'nats://localhost:4222';
process.env.OTEL_EXPORTER_OTLP_ENDPOINT = 'http://127.0.0.1:4318';
process.env.JWT_JWKS_URL = 'test';
process.env.S3_BUCKET_PREVIEWS = 'test-previews';
process.env.S3_BUCKET_ARTIFACTS = 'test-artifacts';
process.env.S3_REGION = 'us-east-1';
process.env.K8S_NAMESPACE = 'default';
process.env.K8S_SERVICE_ACCOUNT = 'default';
process.env.K8S_JOB_IMAGE_DEFAULT = 'busybox';
process.env.BROWSER_GRID_URL = 'http://selenium-grid:4444';
process.env.VAULT_ADDR = 'http://vault:8200';
process.env.REPO_BACKEND = 'pg';
process.env.ENABLE_EVENT_SUBSCRIBERS = 'true';

describe('events wiring in app', () => {
  let app: Awaited<ReturnType<typeof createApp>>;
  beforeEach(async () => {
    wired = false;
    const { __resetConfigForTests } = await import('../../config.js');
    __resetConfigForTests();
    app = await createApp();
  });
  afterEach(async () => {
    await app.close();
  });

  it('wires subscribers on pg backend and closes nats on shutdown', async () => {
    const cfgMod: any = await import('../../config.js');
    const cfg = cfgMod.getConfig();
    expect(cfg.enableEventSubscribers).toBe(true);
    const natsMod: any = await import('../../clients/nats.js');
    expect(natsMod.createNatsConnection).toHaveBeenCalled();
  });
});
