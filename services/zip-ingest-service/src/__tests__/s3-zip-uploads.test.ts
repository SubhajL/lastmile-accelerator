import { afterEach, describe, expect, test, vi } from 'vitest';

import { buildZipUploadObjectKey, getS3ZipUploadsConfigFromEnv } from '../lib/s3-zip-uploads.js';

describe('s3-zip-uploads', () => {
  afterEach(() => {
    vi.unstubAllEnvs();
  });

  describe('buildZipUploadObjectKey', () => {
    test('normalizes prefix and url-encodes projectId', () => {
      const objectKey = buildZipUploadObjectKey({
        prefix: 'zip-uploads',
        projectId: 'proj/evil',
        uploadId: 'upl_123',
      });

      expect(objectKey).toBe('zip-uploads/proj%2Fevil/upl_123.zip');
    });
  });

  describe('getS3ZipUploadsConfigFromEnv', () => {
    test('parses boolean and normalizes prefix', () => {
      vi.stubEnv('SNAPSHOT_BUCKET', 'snapshots');
      vi.stubEnv('SNAPSHOT_S3_ENDPOINT', 'http://minio.example:9000');
      vi.stubEnv('SNAPSHOT_S3_ACCESS_KEY', 'k');
      vi.stubEnv('SNAPSHOT_S3_SECRET_KEY', 's');
      vi.stubEnv('SNAPSHOT_S3_FORCE_PATH_STYLE', 'no');
      vi.stubEnv('ZIP_UPLOAD_PREFIX', 'zip-uploads');

      const cfg = getS3ZipUploadsConfigFromEnv();
      expect(cfg).toEqual({
        bucket: 'snapshots',
        endpoint: 'http://minio.example:9000',
        region: 'us-east-1',
        accessKeyId: 'k',
        secretAccessKey: 's',
        forcePathStyle: false,
        prefix: 'zip-uploads/',
      });
    });

    test('returns null when required env vars are missing', () => {
      vi.stubEnv('SNAPSHOT_BUCKET', 'snapshots');
      vi.stubEnv('SNAPSHOT_S3_ENDPOINT', 'http://minio.example:9000');
      vi.stubEnv('SNAPSHOT_S3_ACCESS_KEY', 'k');

      const cfg = getS3ZipUploadsConfigFromEnv();
      expect(cfg).toBeNull();
    });
  });
});
