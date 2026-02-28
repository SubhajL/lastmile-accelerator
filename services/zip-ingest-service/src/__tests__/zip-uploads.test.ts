import { createHash } from 'node:crypto';
import http from 'node:http';

import { afterEach, describe, expect, test, vi } from 'vitest';

import { createApp } from '../app.js';
import { createInMemoryZipUploadSessionStore } from '../lib/zip-upload-session-store.js';

function startStubOrchestrator(args: {
  handler: (req: http.IncomingMessage, body: string) => { statusCode?: number; body: unknown };
}) {
  const server = http.createServer(async (req, res) => {
    let body = '';
    for await (const chunk of req) body += chunk;
    const response = args.handler(req, body);
    res.statusCode = response.statusCode ?? 200;
    res.setHeader('content-type', 'application/json');
    res.end(JSON.stringify(response.body));
  });

  return new Promise<{ url: string; close: () => Promise<void> }>((resolve) => {
    server.listen(0, '127.0.0.1', () => {
      const address = server.address();
      if (!address || typeof address === 'string') throw new Error('unexpected server address');
      resolve({
        url: `http://127.0.0.1:${address.port}`,
        close: () => new Promise((r) => server.close(() => r())),
      });
    });
  });
}

function startStubS3(args: {
  handler: (req: http.IncomingMessage) => { statusCode?: number; headers?: Record<string, string>; body?: string };
}) {
  const server = http.createServer((req, res) => {
    const response = args.handler(req);
    res.statusCode = response.statusCode ?? 200;
    for (const [k, v] of Object.entries(response.headers ?? {})) res.setHeader(k, v);
    res.end(response.body ?? '');
  });

  return new Promise<{ url: string; close: () => Promise<void> }>((resolve) => {
    server.listen(0, '127.0.0.1', () => {
      const address = server.address();
      if (!address || typeof address === 'string') throw new Error('unexpected server address');
      resolve({
        url: `http://127.0.0.1:${address.port}`,
        close: () => new Promise((r) => server.close(() => r())),
      });
    });
  });
}

describe('zip signed-upload flow', () => {
  afterEach(() => {
    vi.unstubAllEnvs();
  });

  test('initiate returns uploadUrl and upload creates snapshot', async () => {
    const snapshotId = 'snap_0123456789abcdef0123456789abcdef';
    const zipBytes = Buffer.from('not really a zip but good enough for MVP tests', 'utf8');
    const sha256 = createHash('sha256').update(zipBytes).digest('hex');

    const stub = await startStubOrchestrator({
      handler: (req, body) => {
        expect(req.method).toBe('POST');
        expect(req.url).toBe('/v1/projects/p123/snapshots');

        const parsed = JSON.parse(body) as {
          mode: string;
          sourceRef: {
            zip: {
              filename: string;
              sizeBytes: number;
              sha256: string;
              storageKey?: string;
              uploadId?: string;
            };
          };
        };
        expect(parsed.mode).toBe('C');
        expect(parsed.sourceRef.zip.filename).toBe('repo.zip');
        expect(parsed.sourceRef.zip.sizeBytes).toBe(zipBytes.length);
        expect(parsed.sourceRef.zip.sha256).toBe(sha256);
        expect(typeof parsed.sourceRef.zip.storageKey).toBe('string');
        expect(parsed.sourceRef.zip.storageKey).toMatch(/^local:/);
        expect(parsed.sourceRef.zip.uploadId).toMatch(/^upl_/);

        return { body: { snapshotId } };
      },
    });

    const app = await createApp({
      snapshotOrchestratorUrl: stub.url,
      uploadBaseUrl: 'http://zip-ingest.example',
      uploadSigningSecret: 'test-secret',
      logger: false,
    });

    const initRes = await app.inject({
      method: 'POST',
      url: '/v1/projects/p123/ingest/zip/initiate',
      payload: { filename: 'repo.zip' },
    });
    expect(initRes.statusCode).toBe(201);
    const initJson = initRes.json() as {
      uploadId: string;
      uploadUrl: string;
    };
    expect(initJson.uploadId).toMatch(/^upl_/);
    expect(initJson.uploadUrl).toMatch(
      /^http:\/\/zip-ingest\.example\/v1\/projects\/p123\/ingest\/zip\/uploads\/upl_/,
    );

    const uploadUrl = new URL(initJson.uploadUrl);
    const uploadRes = await app.inject({
      method: 'PUT',
      url: uploadUrl.pathname + uploadUrl.search,
      payload: zipBytes,
      headers: {
        'content-type': 'application/zip',
        'x-filename': 'repo.zip',
      },
    });

    expect(uploadRes.statusCode).toBe(201);
    expect(uploadRes.json()).toEqual({
      snapshotId,
      sha256,
      sizeBytes: zipBytes.length,
    });

    await Promise.all([app.close(), stub.close()]);
  });

  test('initiate returns S3 presigned URL when ZIP_UPLOAD_BACKEND=s3', async () => {
    vi.stubEnv('ZIP_UPLOAD_BACKEND', 's3');
    vi.stubEnv('SNAPSHOT_BUCKET', 'snapshots');
    vi.stubEnv('SNAPSHOT_S3_ENDPOINT', 'http://minio.example:9000');
    vi.stubEnv('SNAPSHOT_S3_REGION', 'us-east-1');
    vi.stubEnv('SNAPSHOT_S3_ACCESS_KEY', 'minio-access-key');
    vi.stubEnv('SNAPSHOT_S3_SECRET_KEY', 'minio-secret-key');
    vi.stubEnv('SNAPSHOT_S3_FORCE_PATH_STYLE', 'true');
    vi.stubEnv('ZIP_UPLOAD_PREFIX', 'zip-uploads/');

    const sessionStore = createInMemoryZipUploadSessionStore();
    const app = await createApp({ zipUploadSessionStore: sessionStore, logger: false });

    const initRes = await app.inject({
      method: 'POST',
      url: '/v1/projects/p123/ingest/zip/initiate',
      payload: { filename: 'repo.zip' },
    });
    expect(initRes.statusCode).toBe(201);
    const initJson = initRes.json() as {
      uploadId: string;
      uploadUrl: string;
      uploadBackend?: string;
      bucket?: string;
      objectKey?: string;
      method?: string;
      contentType?: string;
      filename: string;
      maxBytes: number;
    };
    expect(initJson.uploadId).toMatch(/^upl_/);
    expect(initJson.uploadBackend).toBe('s3');
    expect(initJson.bucket).toBe('snapshots');
    expect(initJson.objectKey).toBe(`zip-uploads/p123/${initJson.uploadId}.zip`);
    expect(initJson.method).toBe('PUT');
    expect(initJson.contentType).toBe('application/zip');

    const uploadUrl = new URL(initJson.uploadUrl);
    expect(uploadUrl.origin).toBe('http://minio.example:9000');
    expect(uploadUrl.pathname).toBe(`/snapshots/${initJson.objectKey}`);
    expect(uploadUrl.searchParams.get('X-Amz-Algorithm')).toBe('AWS4-HMAC-SHA256');
    expect(uploadUrl.searchParams.get('X-Amz-Expires')).toBeTruthy();
    expect(uploadUrl.searchParams.get('X-Amz-Signature')).toBeTruthy();

    await expect(sessionStore.getSession(initJson.uploadId)).resolves.toEqual({
      projectId: 'p123',
      bucket: 'snapshots',
      objectKey: `zip-uploads/p123/${initJson.uploadId}.zip`,
      expiresAtUnixSeconds: expect.any(Number),
    });

    await app.close();
  });

  test('initiate returns 500 when ZIP_UPLOAD_BACKEND=s3 but S3 env is missing', async () => {
    vi.stubEnv('ZIP_UPLOAD_BACKEND', 's3');

    const app = await createApp({ logger: false });

    const initRes = await app.inject({
      method: 'POST',
      url: '/v1/projects/p123/ingest/zip/initiate',
      payload: { filename: 'repo.zip' },
    });
    expect(initRes.statusCode).toBe(500);
    expect(initRes.json()).toEqual({ error: 's3_not_configured' });

    await app.close();
  });

  test('initiate returns 500 when ZIP_UPLOAD_BACKEND=s3 but REDIS_URL is missing (no injected store)', async () => {
    vi.stubEnv('ZIP_UPLOAD_BACKEND', 's3');
    vi.stubEnv('SNAPSHOT_BUCKET', 'snapshots');
    vi.stubEnv('SNAPSHOT_S3_ENDPOINT', 'http://minio.example:9000');
    vi.stubEnv('SNAPSHOT_S3_REGION', 'us-east-1');
    vi.stubEnv('SNAPSHOT_S3_ACCESS_KEY', 'minio-access-key');
    vi.stubEnv('SNAPSHOT_S3_SECRET_KEY', 'minio-secret-key');
    vi.stubEnv('SNAPSHOT_S3_FORCE_PATH_STYLE', 'true');
    vi.stubEnv('ZIP_UPLOAD_PREFIX', 'zip-uploads/');

    const app = await createApp({ logger: false });

    const initRes = await app.inject({
      method: 'POST',
      url: '/v1/projects/p123/ingest/zip/initiate',
      payload: { filename: 'repo.zip' },
    });
    expect(initRes.statusCode).toBe(500);
    expect(initRes.json()).toEqual({ error: 'redis_not_configured' });

    await app.close();
  });

  test('complete HEADs object and creates snapshot (S3 backend)', async () => {
    const snapshotId = 'snap_0123456789abcdef0123456789abcdef';
    const sizeBytes = 1234;
    const sha256 = 'a'.repeat(64);

    const s3 = await startStubS3({
      handler: (req) => {
        expect(req.method).toBe('HEAD');
        expect(req.url).toContain('/snapshots/zip-uploads/p123/');
        return { headers: { 'content-length': String(sizeBytes) } };
      },
    });

    const stub = await startStubOrchestrator({
      handler: (req, body) => {
        expect(req.method).toBe('POST');
        expect(req.url).toBe('/v1/projects/p123/snapshots');

        const parsed = JSON.parse(body) as {
          mode: string;
          sourceRef: {
            zip: {
              filename: string;
              sizeBytes: number;
              sha256?: string;
              storageKey?: string;
              uploadId?: string;
            };
          };
        };
        expect(parsed).toEqual({
          mode: 'C',
          sourceRef: {
            zip: {
              filename: 'repo.zip',
              sizeBytes,
              sha256,
              storageKey: expect.any(String),
              uploadId: expect.any(String),
            },
          },
        });
        expect(parsed.sourceRef.zip.storageKey).toMatch(/^s3:\/\//);

        return { body: { snapshotId } };
      },
    });

    vi.stubEnv('ZIP_UPLOAD_BACKEND', 's3');
    vi.stubEnv('SNAPSHOT_BUCKET', 'snapshots');
    vi.stubEnv('SNAPSHOT_S3_ENDPOINT', s3.url);
    vi.stubEnv('SNAPSHOT_S3_REGION', 'us-east-1');
    vi.stubEnv('SNAPSHOT_S3_ACCESS_KEY', 'minio-access-key');
    vi.stubEnv('SNAPSHOT_S3_SECRET_KEY', 'minio-secret-key');
    vi.stubEnv('SNAPSHOT_S3_FORCE_PATH_STYLE', 'true');
    vi.stubEnv('ZIP_UPLOAD_PREFIX', 'zip-uploads/');

    const sessionStore = createInMemoryZipUploadSessionStore();
    const app = await createApp({
      snapshotOrchestratorUrl: stub.url,
      zipUploadSessionStore: sessionStore,
      logger: false,
    });

    const initRes = await app.inject({
      method: 'POST',
      url: '/v1/projects/p123/ingest/zip/initiate',
      payload: { filename: 'repo.zip' },
    });
    expect(initRes.statusCode).toBe(201);
    const initJson = initRes.json() as { uploadId: string; objectKey?: string };
    expect(initJson.uploadId).toMatch(/^upl_/);
    expect(initJson.objectKey).toContain(`/p123/${initJson.uploadId}.zip`);

    const completeRes = await app.inject({
      method: 'POST',
      url: '/v1/projects/p123/ingest/zip/complete',
      payload: { uploadId: initJson.uploadId, filename: 'repo.zip', sha256 },
    });

    expect(completeRes.statusCode).toBe(201);
    expect(completeRes.json()).toEqual({ snapshotId });

    await Promise.all([app.close(), stub.close(), s3.close()]);
  });

  test('complete returns 400 when ZIP_UPLOAD_BACKEND is not s3', async () => {
    const app = await createApp({ logger: false });

    const res = await app.inject({
      method: 'POST',
      url: '/v1/projects/p123/ingest/zip/complete',
      payload: { uploadId: 'upl_deadbeefdeadbeefdeadbeefdeadbeef', filename: 'repo.zip' },
    });

    expect(res.statusCode).toBe(400);
    expect(res.json()).toEqual({ error: 'zip_upload_backend_not_s3' });

    await app.close();
  });

  test('complete returns 404 when uploadId is unknown', async () => {
    const s3 = await startStubS3({
      handler: () => ({ statusCode: 200, headers: { 'content-length': '1' } }),
    });

    vi.stubEnv('ZIP_UPLOAD_BACKEND', 's3');
    vi.stubEnv('SNAPSHOT_BUCKET', 'snapshots');
    vi.stubEnv('SNAPSHOT_S3_ENDPOINT', s3.url);
    vi.stubEnv('SNAPSHOT_S3_REGION', 'us-east-1');
    vi.stubEnv('SNAPSHOT_S3_ACCESS_KEY', 'minio-access-key');
    vi.stubEnv('SNAPSHOT_S3_SECRET_KEY', 'minio-secret-key');
    vi.stubEnv('SNAPSHOT_S3_FORCE_PATH_STYLE', 'true');

    const sessionStore = createInMemoryZipUploadSessionStore();
    const app = await createApp({ zipUploadSessionStore: sessionStore, logger: false });
    const res = await app.inject({
      method: 'POST',
      url: '/v1/projects/p123/ingest/zip/complete',
      payload: { uploadId: 'upl_deadbeefdeadbeefdeadbeefdeadbeef', filename: 'repo.zip' },
    });

    expect(res.statusCode).toBe(404);
    expect(res.json()).toEqual({ error: 'zip_upload_not_found' });

    await Promise.all([app.close(), s3.close()]);
  });

  test('complete returns 404 when S3 object is missing', async () => {
    const s3 = await startStubS3({
      handler: () => ({ statusCode: 404 }),
    });

    vi.stubEnv('ZIP_UPLOAD_BACKEND', 's3');
    vi.stubEnv('SNAPSHOT_BUCKET', 'snapshots');
    vi.stubEnv('SNAPSHOT_S3_ENDPOINT', s3.url);
    vi.stubEnv('SNAPSHOT_S3_REGION', 'us-east-1');
    vi.stubEnv('SNAPSHOT_S3_ACCESS_KEY', 'minio-access-key');
    vi.stubEnv('SNAPSHOT_S3_SECRET_KEY', 'minio-secret-key');
    vi.stubEnv('SNAPSHOT_S3_FORCE_PATH_STYLE', 'true');
    vi.stubEnv('ZIP_UPLOAD_PREFIX', 'zip-uploads/');

    const sessionStore = createInMemoryZipUploadSessionStore();
    const app = await createApp({ zipUploadSessionStore: sessionStore, logger: false });

    const initRes = await app.inject({
      method: 'POST',
      url: '/v1/projects/p123/ingest/zip/initiate',
      payload: { filename: 'repo.zip' },
    });
    const initJson = initRes.json() as { uploadId: string };

    const res = await app.inject({
      method: 'POST',
      url: '/v1/projects/p123/ingest/zip/complete',
      payload: { uploadId: initJson.uploadId, filename: 'repo.zip' },
    });

    expect(res.statusCode).toBe(404);
    expect(res.json()).toEqual({ error: 'zip_upload_not_found' });

    await Promise.all([app.close(), s3.close()]);
  });

  test('complete returns 502 when snapshot-orchestrator fails', async () => {
    const s3 = await startStubS3({
      handler: () => ({ statusCode: 200, headers: { 'content-length': '1' } }),
    });

    const stub = await startStubOrchestrator({
      handler: () => ({ statusCode: 500, body: { error: 'nope' } }),
    });

    vi.stubEnv('ZIP_UPLOAD_BACKEND', 's3');
    vi.stubEnv('SNAPSHOT_BUCKET', 'snapshots');
    vi.stubEnv('SNAPSHOT_S3_ENDPOINT', s3.url);
    vi.stubEnv('SNAPSHOT_S3_REGION', 'us-east-1');
    vi.stubEnv('SNAPSHOT_S3_ACCESS_KEY', 'minio-access-key');
    vi.stubEnv('SNAPSHOT_S3_SECRET_KEY', 'minio-secret-key');
    vi.stubEnv('SNAPSHOT_S3_FORCE_PATH_STYLE', 'true');
    vi.stubEnv('ZIP_UPLOAD_PREFIX', 'zip-uploads/');

    const sessionStore = createInMemoryZipUploadSessionStore();
    const app = await createApp({
      snapshotOrchestratorUrl: stub.url,
      zipUploadSessionStore: sessionStore,
      logger: false,
    });

    const initRes = await app.inject({
      method: 'POST',
      url: '/v1/projects/p123/ingest/zip/initiate',
      payload: { filename: 'repo.zip' },
    });
    const initJson = initRes.json() as { uploadId: string };

    const res = await app.inject({
      method: 'POST',
      url: '/v1/projects/p123/ingest/zip/complete',
      payload: { uploadId: initJson.uploadId, filename: 'repo.zip' },
    });

    expect(res.statusCode).toBe(502);
    expect(res.json()).toEqual({ error: 'snapshot_orchestrator_unavailable' });

    await Promise.all([app.close(), stub.close(), s3.close()]);
  });

  test('upload rejects invalid token', async () => {
    const app = await createApp({
      uploadBaseUrl: 'http://zip-ingest.example',
      uploadSigningSecret: 'test-secret',
      logger: false,
    });

    const initRes = await app.inject({
      method: 'POST',
      url: '/v1/projects/p123/ingest/zip/initiate',
      payload: { filename: 'repo.zip' },
    });
    const initJson = initRes.json() as { uploadUrl: string };
    const uploadUrl = new URL(initJson.uploadUrl);
    uploadUrl.searchParams.set('token', 'deadbeef');

    const uploadRes = await app.inject({
      method: 'PUT',
      url: uploadUrl.pathname + uploadUrl.search,
      payload: Buffer.from('x', 'utf8'),
      headers: { 'content-type': 'application/zip' },
    });

    expect(uploadRes.statusCode).toBe(401);
    expect(uploadRes.json()).toEqual({ error: 'invalid_upload_token' });

    await app.close();
  });
});

describe('zip direct-upload endpoint', () => {
  test('POST upload computes sha/size and creates snapshot', async () => {
    const snapshotId = 'snap_0123456789abcdef0123456789abcdef';
    const zipBytes = Buffer.from('hello', 'utf8');
    const sha256 = createHash('sha256').update(zipBytes).digest('hex');

    const stub = await startStubOrchestrator({
      handler: (req, body) => {
        expect(req.method).toBe('POST');
        expect(req.url).toBe('/v1/projects/p123/snapshots');
        expect(JSON.parse(body)).toEqual({
          mode: 'C',
          sourceRef: {
            zip: {
              filename: 'repo.zip',
              sizeBytes: zipBytes.length,
              sha256,
              storageKey: expect.any(String),
            },
          },
        });

        return { body: { snapshotId } };
      },
    });

    const app = await createApp({ snapshotOrchestratorUrl: stub.url, logger: false });
    const res = await app.inject({
      method: 'POST',
      url: '/v1/projects/p123/ingest/zip/upload',
      payload: zipBytes,
      headers: { 'content-type': 'application/zip', 'x-filename': 'repo.zip' },
    });

    expect(res.statusCode).toBe(201);
    expect(res.json()).toEqual({ snapshotId });

    await Promise.all([app.close(), stub.close()]);
  });
});
