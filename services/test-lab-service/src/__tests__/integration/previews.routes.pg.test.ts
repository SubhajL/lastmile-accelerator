import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { createApp } from '../../app.js';
import { createMockJWT, TEST_JWT_SECRET } from '../fixtures/jwt-helpers.js';

process.env.SERVICE_NAME = 'test-lab-service';
process.env.SERVICE_PORT = '7202';
process.env.DATABASE_URL = 'pgmem://previews';
process.env.REDIS_URL = 'redis://localhost:6379';
process.env.NATS_URL = 'nats://localhost:4222';
process.env.OTEL_EXPORTER_OTLP_ENDPOINT = 'http://127.0.0.1:4318';
process.env.JWT_JWKS_URL = TEST_JWT_SECRET;
process.env.S3_BUCKET_PREVIEWS = 'test-previews';
process.env.BROWSER_GRID_URL = 'http://selenium-grid:4444';
process.env.VAULT_ADDR = 'http://vault:8200';
process.env.REPO_BACKEND = 'pg';

const projectId = '11111111-1111-1111-1111-111111111111';

describe('preview environments routes (pg)', () => {
  let app: Awaited<ReturnType<typeof createApp>>;
  let token: string;

  beforeEach(async () => {
    app = await createApp();
    token = createMockJWT({ scopes: ['preview:read', 'preview:write'] });
  });

  afterEach(async () => {
    await app.close();
  });

  it('CRUD and pagination', async () => {
    // create
    const c1 = await app.inject({
      method: 'POST', url: `/v1/projects/${projectId}/previews`,
      headers: { authorization: `Bearer ${token}` },
      payload: { url: 'https://p1.example' },
    });
    expect(c1.statusCode).toBe(201);
    const p1 = c1.json();

    const c2 = await app.inject({
      method: 'POST', url: `/v1/projects/${projectId}/previews`,
      headers: { authorization: `Bearer ${token}` },
      payload: { url: 'https://p2.example' },
    });
    const p2 = c2.json();

    // get
    const g = await app.inject({ method: 'GET', url: `/v1/previews/${p1.id}`, headers: { authorization: `Bearer ${token}` } });
    expect(g.statusCode).toBe(200);

    // update
    const u = await app.inject({
      method: 'PUT', url: `/v1/previews/${p1.id}`,
      headers: { authorization: `Bearer ${token}` },
      payload: { status: 'ready', expiresAt: '2024-01-02T00:00:00.000Z' },
    });
    expect(u.statusCode).toBe(200);
    expect(u.json().status).toBe('ready');

    // list with pagination
    const l1 = await app.inject({ method: 'GET', url: `/v1/projects/${projectId}/previews?limit=1`, headers: { authorization: `Bearer ${token}` } });
    expect(l1.statusCode).toBe(200);
    const page1 = l1.json();
    expect(page1.items.length).toBe(1);

    const l2 = await app.inject({ method: 'GET', url: `/v1/projects/${projectId}/previews?limit=1&cursor=${page1.nextCursor}`, headers: { authorization: `Bearer ${token}` } });
    expect(l2.statusCode).toBe(200);

    // delete
    const d = await app.inject({ method: 'DELETE', url: `/v1/previews/${p2.id}`, headers: { authorization: `Bearer ${token}` } });
    expect(d.statusCode).toBe(204);
  });
});