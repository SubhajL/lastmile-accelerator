import { describe, expect, test } from 'vitest';
import http from 'node:http';

import { createApp } from '../app.js';

function startStubOrchestrator(args: {
  handler: (req: http.IncomingMessage) => { statusCode?: number; body: unknown };
}) {
  const server = http.createServer(async (req, res) => {
    const response = args.handler(req);
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

describe('GET /v1/snapshots/:snapshotId/fix-list', () => {
  test('returns deterministic fix list for existing snapshot', async () => {
    const snapshotId = 'snap_0123456789abcdef0123456789abcdef';
    const stub = await startStubOrchestrator({
      handler: (req) => {
        expect(req.method).toBe('GET');
        expect(req.url).toBe(`/v1/snapshots/${snapshotId}`);
        return { body: { snapshotId } };
      },
    });

    const app = await createApp({ logger: false, snapshotOrchestratorUrl: stub.url });
    const res = await app.inject({
      method: 'GET',
      url: `/v1/snapshots/${snapshotId}/fix-list`,
    });

    expect(res.statusCode).toBe(200);
    expect(res.json()).toEqual({
      snapshotId,
      fixes: [
        {
          id: `fix_${snapshotId}`,
          tier: 3,
          title: 'MVP placeholder fix',
          description: 'This is a deterministic placeholder fix-list entry.',
        },
      ],
    });

    await Promise.all([app.close(), stub.close()]);
  });

  test('returns 400 for invalid snapshotId', async () => {
    const app = await createApp({ logger: false });
    const res = await app.inject({
      method: 'GET',
      url: '/v1/snapshots/not-a-snapshot/fix-list',
    });

    expect(res.statusCode).toBe(400);
    expect(res.json()).toEqual({ error: 'invalid_snapshot_id' });

    await app.close();
  });

  test('returns 404 when snapshot does not exist', async () => {
    const snapshotId = 'snap_0123456789abcdef0123456789abcdef';
    const stub = await startStubOrchestrator({
      handler: () => ({ statusCode: 404, body: { error: 'not found' } }),
    });

    const app = await createApp({ logger: false, snapshotOrchestratorUrl: stub.url });
    const res = await app.inject({
      method: 'GET',
      url: `/v1/snapshots/${snapshotId}/fix-list`,
    });

    expect(res.statusCode).toBe(404);
    expect(res.json()).toEqual({ error: 'snapshot_not_found' });

    await Promise.all([app.close(), stub.close()]);
  });
});

describe('GET /v1/fixes/:fixId/manifest', () => {
  test('returns manifest stub', async () => {
    const app = await createApp({ logger: false });
    const res = await app.inject({
      method: 'GET',
      url: '/v1/fixes/fix_123/manifest',
    });

    expect(res.statusCode).toBe(200);
    expect(res.json()).toEqual({
      fixId: 'fix_123',
      summary: 'MVP placeholder manifest',
      files: [],
      locDelta: 0,
      tests: [],
    });

    await app.close();
  });
});
