import { describe, expect, test } from 'vitest';

import { createInMemoryZipUploadSessionStore } from './zip-upload-session-store.js';

describe('createInMemoryZipUploadSessionStore', () => {
  test('round-trips put/get/delete', async () => {
    const store = createInMemoryZipUploadSessionStore();
    const uploadId = 'upl_123';

    await store.putSession({
      uploadId,
      ttlSeconds: 60,
      session: {
        projectId: 'p123',
        bucket: 'snapshots',
        objectKey: 'zip-uploads/p123/upl_123.zip',
        expiresAtUnixSeconds: 123,
      },
    });

    await expect(store.getSession(uploadId)).resolves.toEqual({
      projectId: 'p123',
      bucket: 'snapshots',
      objectKey: 'zip-uploads/p123/upl_123.zip',
      expiresAtUnixSeconds: 123,
    });

    await store.deleteSession(uploadId);
    await expect(store.getSession(uploadId)).resolves.toBeNull();
    await store.close();
  });
});

