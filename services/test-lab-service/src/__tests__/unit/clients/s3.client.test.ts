import { describe, it, expect, vi, beforeEach } from 'vitest';
import { S3Client, PutObjectCommand } from '@aws-sdk/client-s3';
import { createS3Client, putArtifact } from '../../../clients/s3.js';

(vi as any).mock('@aws-sdk/client-s3', () => ({
  S3Client: vi.fn().mockImplementation(() => ({ send: vi.fn() })),
  PutObjectCommand: vi.fn().mockImplementation((input: any) => ({ input })),
}), { virtual: true });

describe('S3 client', () => {
  let s3: S3Client & { send: ReturnType<typeof vi.fn> };

  beforeEach(() => {
    s3 = createS3Client('us-east-1') as any;
  });

  it('uploads artifact and returns bucket/key', async () => {
    (s3.send as any).mockResolvedValueOnce({});
    const res = await putArtifact(s3 as any, { bucket: 'b', key: 'k', body: Buffer.from('x'), contentType: 'text/plain' });
    expect(res).toEqual({ bucket: 'b', key: 'k' });
    expect((PutObjectCommand as any).mock.calls[0][0]).toMatchObject({ Bucket: 'b', Key: 'k' });
  });

  it('propagates sdk error message', async () => {
    (s3.send as any).mockRejectedValueOnce(new Error('boom'));
    await expect(putArtifact(s3 as any, { bucket: 'b', key: 'k', body: 'x' })).rejects.toThrow('boom');
  });
});
