import { afterEach, describe, expect, test, vi } from 'vitest';

describe('zip uploads (redis store init retry)', () => {
  afterEach(() => {
    vi.unstubAllEnvs();
    vi.resetModules();
    vi.clearAllMocks();
  });

  test('initiate retries Redis store initialization after a transient failure', async () => {
    vi.stubEnv('ZIP_UPLOAD_BACKEND', 's3');
    vi.stubEnv('SNAPSHOT_BUCKET', 'snapshots');
    vi.stubEnv('SNAPSHOT_S3_ENDPOINT', 'http://minio.example:9000');
    vi.stubEnv('SNAPSHOT_S3_REGION', 'us-east-1');
    vi.stubEnv('SNAPSHOT_S3_ACCESS_KEY', 'minio-access-key');
    vi.stubEnv('SNAPSHOT_S3_SECRET_KEY', 'minio-secret-key');
    vi.stubEnv('SNAPSHOT_S3_FORCE_PATH_STYLE', 'true');
    vi.stubEnv('ZIP_UPLOAD_PREFIX', 'zip-uploads/');
    vi.stubEnv('REDIS_URL', 'redis://redis.example:6379');

    let initCalls = 0;
    vi.doMock('../lib/redis-zip-upload-session-store.js', async () => {
      const { createInMemoryZipUploadSessionStore } = await import(
        '../lib/zip-upload-session-store.js'
      );
      return {
        createRedisZipUploadSessionStoreFromUrl: async () => {
          initCalls += 1;
          if (initCalls === 1) throw new Error('redis temporarily unavailable');
          return createInMemoryZipUploadSessionStore();
        },
      };
    });

    const { createApp } = await import('../app.js');
    const app = await createApp({ logger: false });

    const first = await app.inject({
      method: 'POST',
      url: '/v1/projects/p123/ingest/zip/initiate',
      payload: { filename: 'repo.zip' },
    });
    expect(first.statusCode).toBe(502);
    expect(first.json()).toEqual({ error: 'redis_unavailable' });

    const second = await app.inject({
      method: 'POST',
      url: '/v1/projects/p123/ingest/zip/initiate',
      payload: { filename: 'repo.zip' },
    });
    expect(second.statusCode).toBe(201);
    expect(initCalls).toBe(2);

    await app.close();
  });
});

