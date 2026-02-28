import { describe, expect, test } from 'vitest';

import { createRedisZipUploadSessionStore } from './redis-zip-upload-session-store.js';

describe('createRedisZipUploadSessionStore', () => {
  test('stores sessions by keyPrefix and respects ttlSeconds', async () => {
    const kv = new Map<string, string>();
    const ttlByKey = new Map<string, number | undefined>();

    const store = createRedisZipUploadSessionStore({
      keyPrefix: 'zip-upload-sessions:',
      client: {
        async get(key) {
          return kv.get(key) ?? null;
        },
        async set(key, value, ttlSeconds) {
          kv.set(key, value);
          ttlByKey.set(key, ttlSeconds);
        },
        async del(key) {
          kv.delete(key);
          ttlByKey.delete(key);
        },
        async close() {},
      },
    });

    await store.putSession({
      uploadId: 'upl_123',
      ttlSeconds: 120,
      session: {
        projectId: 'p123',
        bucket: 'snapshots',
        objectKey: 'zip-uploads/p123/upl_123.zip',
        expiresAtUnixSeconds: 456,
      },
    });

    expect(ttlByKey.get('zip-upload-sessions:upl_123')).toBe(120);
    await expect(store.getSession('upl_123')).resolves.toEqual({
      projectId: 'p123',
      bucket: 'snapshots',
      objectKey: 'zip-uploads/p123/upl_123.zip',
      expiresAtUnixSeconds: 456,
    });

    await store.deleteSession('upl_123');
    await expect(store.getSession('upl_123')).resolves.toBeNull();
  });

  test('getSession returns null and deletes key on invalid JSON', async () => {
    const kv = new Map<string, string>();
    const deleted = new Set<string>();

    kv.set('zip-upload-sessions:upl_bad', 'not-json');

    const store = createRedisZipUploadSessionStore({
      keyPrefix: 'zip-upload-sessions:',
      client: {
        async get(key) {
          return kv.get(key) ?? null;
        },
        async set(key, value) {
          kv.set(key, value);
        },
        async del(key) {
          kv.delete(key);
          deleted.add(key);
        },
        async close() {},
      },
    });

    await expect(store.getSession('upl_bad')).resolves.toBeNull();
    expect(deleted.has('zip-upload-sessions:upl_bad')).toBe(true);
  });

  test('getSession returns null and deletes key on invalid shape', async () => {
    const kv = new Map<string, string>();
    const deleted = new Set<string>();

    kv.set(
      'zip-upload-sessions:upl_shape',
      JSON.stringify({
        projectId: 'p123',
        bucket: 'snapshots',
        objectKey: 'zip-uploads/p123/upl_shape.zip',
        expiresAtUnixSeconds: 'not-a-number',
      }),
    );

    const store = createRedisZipUploadSessionStore({
      keyPrefix: 'zip-upload-sessions:',
      client: {
        async get(key) {
          return kv.get(key) ?? null;
        },
        async set(key, value) {
          kv.set(key, value);
        },
        async del(key) {
          kv.delete(key);
          deleted.add(key);
        },
        async close() {},
      },
    });

    await expect(store.getSession('upl_shape')).resolves.toBeNull();
    expect(deleted.has('zip-upload-sessions:upl_shape')).toBe(true);
  });
});
