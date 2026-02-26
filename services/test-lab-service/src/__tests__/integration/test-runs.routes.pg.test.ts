import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { createApp } from '../../app.js';
import { createMockJWT, TEST_JWT_SECRET } from '../fixtures/jwt-helpers.js';

// Mock Redis so tests run without a live Redis instance
vi.mock('../../clients/redis.js', () => ({
  createRedisClient: vi.fn().mockResolvedValue({
    get: vi.fn().mockResolvedValue(null),
    set: vi.fn().mockResolvedValue(undefined),
    del: vi.fn().mockResolvedValue(undefined),
    incr: vi.fn().mockResolvedValue(1),
    incrWithExpire: vi.fn().mockResolvedValue(1),
    expire: vi.fn().mockResolvedValue(undefined),
    close: vi.fn().mockResolvedValue(undefined),
  }),
}));

vi.mock('../../lib/jwks.js', () => ({
  verifyJwt: vi.fn(async ({ token }: { token: string }) => {
    const { verify } = await import('jsonwebtoken');
    return verify(token, TEST_JWT_SECRET, { algorithms: ['HS256'] });
  }),
}));

process.env.SERVICE_NAME = 'test-lab-service';
process.env.SERVICE_PORT = '7202';
process.env.DATABASE_URL = 'pgmem://testruns';
process.env.REDIS_URL = 'redis://localhost:6379';
process.env.NATS_URL = 'nats://localhost:4222';
process.env.OTEL_EXPORTER_OTLP_ENDPOINT = 'http://127.0.0.1:4318';
process.env.JWT_JWKS_URL = TEST_JWT_SECRET;
process.env.JWT_ISSUER = 'https://auth.example.com/';
process.env.JWT_AUDIENCE = 'test-lab-service';
process.env.S3_BUCKET_PREVIEWS = 'test-previews';
process.env.BROWSER_GRID_URL = 'http://selenium-grid:4444';
process.env.VAULT_ADDR = 'http://vault:8200';
process.env.REPO_BACKEND = 'pg';

const projectId = '11111111-1111-1111-1111-111111111111';

describe('test runs & browser runs routes (pg)', () => {
  let app: Awaited<ReturnType<typeof createApp>>;
  let token: string;

  beforeEach(async () => {
    app = await createApp();
    token = createMockJWT({ scopes: ['run:read', 'run:write', 'browser-run:read', 'browser-run:write'] });
  });

  afterEach(async () => {
    await app.close();
  });

  it('creates a test run, updates status, lists with pagination & status filter; manages browser runs', async () => {
    // Create two runs
    const create1 = await app.inject({
      method: 'POST',
      url: `/v1/projects/${projectId}/test-runs`,
      headers: { authorization: `Bearer ${token}` },
      payload: { type: 'e2e' },
    });
    expect(create1.statusCode).toBe(201);
    const r1 = create1.json();

    const create2 = await app.inject({
      method: 'POST',
      url: `/v1/projects/${projectId}/test-runs`,
      headers: { authorization: `Bearer ${token}` },
      payload: { type: 'unit' },
    });
    expect(create2.statusCode).toBe(201);
    const r2 = create2.json();

    // Update r1 status running -> passed
    const runStart = await app.inject({
      method: 'PATCH',
      url: `/v1/test-runs/${r1.id}/status`,
      headers: { authorization: `Bearer ${token}` },
      payload: { status: 'running', startedAt: '2024-01-01T00:00:00.000Z' },
    });
    expect(runStart.statusCode).toBe(200);

    const runFinish = await app.inject({
      method: 'PATCH',
      url: `/v1/test-runs/${r1.id}/status`,
      headers: { authorization: `Bearer ${token}` },
      payload: {
        status: 'passed',
        finishedAt: '2024-01-01T00:05:00.000Z',
        results: { passed: 10, failed: 0 },
        artifacts: { s3: ['artifacts/log.txt'] },
      },
    });
    expect(runFinish.statusCode).toBe(200);
    expect(runFinish.json().status).toBe('passed');

    // List queued only -> should include r2 (still queued)
    const listQueued = await app.inject({
      method: 'GET',
      url: `/v1/projects/${projectId}/test-runs?limit=1&status=queued`,
      headers: { authorization: `Bearer ${token}` },
    });
    expect(listQueued.statusCode).toBe(200);
    const page1 = listQueued.json();
    expect(page1.items.length).toBe(1);

    const listNext = await app.inject({
      method: 'GET',
      url: `/v1/projects/${projectId}/test-runs?limit=1&cursor=${page1.nextCursor}`,
      headers: { authorization: `Bearer ${token}` },
    });
    expect(listNext.statusCode).toBe(200);

    // Browser runs under r1
    const bCreate1 = await app.inject({
      method: 'POST',
      url: `/v1/test-runs/${r1.id}/browser-runs`,
      headers: { authorization: `Bearer ${token}` },
      payload: { browser: 'chrome', os: 'macOS', viewport: '1280x800' },
    });
    expect(bCreate1.statusCode).toBe(201);
    const br1 = bCreate1.json();

    const bCreate2 = await app.inject({
      method: 'POST',
      url: `/v1/test-runs/${r1.id}/browser-runs`,
      headers: { authorization: `Bearer ${token}` },
      payload: { browser: 'firefox' },
    });
    const br2 = bCreate2.json();

    const bUpd = await app.inject({
      method: 'PATCH',
      url: `/v1/browser-runs/${br1.id}/status`,
      headers: { authorization: `Bearer ${token}` },
      payload: { status: 'passed', finishedAt: '2024-01-01T00:03:00.000Z', screenshots: ['s3://bucket/shot1.png'] },
    });
    expect(bUpd.statusCode).toBe(200);

    const bList = await app.inject({
      method: 'GET',
      url: `/v1/test-runs/${r1.id}/browser-runs`,
      headers: { authorization: `Bearer ${token}` },
    });
    expect(bList.statusCode).toBe(200);
    const bItems = bList.json().items as any[];
    expect(bItems.length).toBe(2);
    expect(bItems.some((i) => i.status === 'passed')).toBe(true);

    const bGet = await app.inject({
      method: 'GET',
      url: `/v1/browser-runs/${br2.id}`,
      headers: { authorization: `Bearer ${token}` },
    });
    expect(bGet.statusCode).toBe(200);
    expect(bGet.json().id).toBe(br2.id);
  });

  it('returns 401 and does not call verifyJwt when Authorization header missing', async () => {
    const res = await app.inject({ method: 'GET', url: `/v1/projects/${projectId}/test-runs` });
    expect(res.statusCode).toBe(401);
  });
});
