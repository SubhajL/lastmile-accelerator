import http from 'node:http';

import { describe, expect, test } from 'vitest';

import { createApp } from '../app.js';

function startStubOrchestrator(args: {
  handler: (req: http.IncomingMessage, body: string) => { statusCode?: number; body: unknown };
}) {
  const server = http.createServer(async (req, res) => {
    const chunks: Buffer[] = [];
    for await (const chunk of req) {
      chunks.push(Buffer.from(chunk));
    }
    const body = Buffer.concat(chunks).toString('utf-8');
    const response = args.handler(req, body);
    res.statusCode = response.statusCode ?? 200;
    res.setHeader('content-type', 'application/json');
    res.end(JSON.stringify(response.body));
  });

  return new Promise<{ url: string; close: () => Promise<void> }>((resolve) => {
    server.listen(0, '127.0.0.1', () => {
      const address = server.address();
      if (!address || typeof address === 'string') {
        throw new Error('unexpected server address');
      }
      resolve({
        url: `http://127.0.0.1:${address.port}`,
        close: () => new Promise((r) => server.close(() => r())),
      });
    });
  });
}

describe('POST /v1/projects/:projectId/ingest/zip', () => {
  test('delegates to snapshot-orchestrator and returns snapshotId', async () => {
    const stub = await startStubOrchestrator({
      handler: (req, body) => {
        expect(req.method).toBe('POST');
        expect(req.url).toBe('/v1/projects/p123/snapshots');
        expect(JSON.parse(body)).toEqual({
          mode: 'C',
          sourceRef: {
            zip: { filename: 'repo.zip', sizeBytes: 123, sha256: 'abc' },
          },
        });

        return {
          body: { snapshotId: 'snap_123', projectId: 'p123', mode: 'C', status: 'created' },
        };
      },
    });

    const app = await createApp({ snapshotOrchestratorUrl: stub.url });
    const res = await app.inject({
      method: 'POST',
      url: '/v1/projects/p123/ingest/zip',
      payload: { filename: 'repo.zip', sizeBytes: 123, sha256: 'abc' },
    });

    expect(res.statusCode).toBe(200);
    expect(res.json()).toEqual({ snapshotId: 'snap_123' });

    await Promise.all([app.close(), stub.close()]);
  });

  test('returns 502 when snapshot-orchestrator fails', async () => {
    const stub = await startStubOrchestrator({
      handler: () => ({ statusCode: 500, body: { error: 'nope' } }),
    });

    const app = await createApp({ snapshotOrchestratorUrl: stub.url });
    const res = await app.inject({
      method: 'POST',
      url: '/v1/projects/p123/ingest/zip',
      payload: { filename: 'repo.zip', sizeBytes: 1, sha256: 'abc' },
    });

    expect(res.statusCode).toBe(502);
    expect(res.json()).toEqual({ error: 'snapshot_orchestrator_unavailable' });

    await Promise.all([app.close(), stub.close()]);
  });
});
