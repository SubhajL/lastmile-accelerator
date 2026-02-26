import { describe, expect, test } from 'vitest';

import { createApp } from '../app.js';

describe('GET /v1/snapshots/:snapshotId/fix-list', () => {
  test('returns deterministic empty fix list', async () => {
    const app = await createApp();
    const res = await app.inject({
      method: 'GET',
      url: '/v1/snapshots/snap_0123456789abcdef0123456789abcdef/fix-list',
    });

    expect(res.statusCode).toBe(200);
    expect(res.json()).toEqual({
      snapshotId: 'snap_0123456789abcdef0123456789abcdef',
      fixes: [],
    });

    await app.close();
  });

  test('returns 400 for invalid snapshotId', async () => {
    const app = await createApp();
    const res = await app.inject({
      method: 'GET',
      url: '/v1/snapshots/not-a-snapshot/fix-list',
    });

    expect(res.statusCode).toBe(400);
    expect(res.json()).toEqual({ error: 'invalid_snapshot_id' });

    await app.close();
  });
});
