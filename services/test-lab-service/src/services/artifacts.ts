import type { S3Client } from '@aws-sdk/client-s3';
import { putArtifact } from '../clients/s3.js';

export interface ArtifactInput {
  name: string;
  content: string | Uint8Array | Buffer;
  contentType?: string;
}

export async function uploadArtifactsMap(
  s3: S3Client,
  {
    bucket,
    prefix,
  }: { bucket: string; prefix: string },
  items: ArtifactInput[]
): Promise<{ bucket: string; key: string }[]> {
  const results: { bucket: string; key: string }[] = [];
  for (const it of items) {
    const key = `${prefix}/${it.name}`;
    const body = typeof it.content === 'string' ? Buffer.from(it.content) : it.content;
    const res = await putArtifact(s3 as any, { bucket, key, body, contentType: it.contentType });
    results.push(res);
  }
  return results;
}
