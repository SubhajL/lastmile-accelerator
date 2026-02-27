import { createHash } from 'node:crypto';
import http from 'node:http';

import { describe, expect, test } from 'vitest';

import { createApp } from '../app.js';

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

describe('zip signed-upload flow', () => {
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
