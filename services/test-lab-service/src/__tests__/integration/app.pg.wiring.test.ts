import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { createApp } from '../../app.js';
import { createMockJWT, TEST_JWT_SECRET } from '../fixtures/jwt-helpers.js';

process.env.SERVICE_NAME = 'test-lab-service';
process.env.SERVICE_PORT = '7202';
process.env.DATABASE_URL = 'pgmem://scaffolds';
process.env.REDIS_URL = 'redis://localhost:6379';
process.env.NATS_URL = 'nats://localhost:4222';
process.env.OTEL_EXPORTER_OTLP_ENDPOINT = 'http://127.0.0.1:4318';
process.env.JWT_JWKS_URL = TEST_JWT_SECRET;
process.env.S3_BUCKET_PREVIEWS = 'test-previews';
process.env.BROWSER_GRID_URL = 'http://selenium-grid:4444';
process.env.VAULT_ADDR = 'http://vault:8200';
process.env.REPO_BACKEND = 'pg';

describe('app with pg repo', () => {
  let app: Awaited<ReturnType<typeof createApp>>;
  const projectId = '11111111-1111-1111-1111-111111111111';
  const token = createMockJWT({ scopes: ['scaffold:read', 'scaffold:write'] });

  beforeEach(async () => {
    app = await createApp();
  });

  afterEach(async () => {
    await app.close();
  });

  it('persists scaffold via pg backend', async () => {
    const create = await app.inject({
      method: 'POST',
      url: `/v1/projects/${projectId}/test-scaffolds`,
      headers: { authorization: `Bearer ${token}` },
      payload: { type: 'unit', framework: 'vitest', language: 'ts', config: {} },
    });
    expect(create.statusCode).toBe(201);

    const id = create.json().id as string;
    const get = await app.inject({ method: 'GET', url: `/v1/test-scaffolds/${id}`, headers: { authorization: `Bearer ${token}` } });
    expect(get.statusCode).toBe(200);
    expect(get.json().id).toBe(id);
  });
});
