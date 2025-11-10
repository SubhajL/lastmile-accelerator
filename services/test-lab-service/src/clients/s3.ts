import { S3Client, PutObjectCommand } from '@aws-sdk/client-s3';

export function createS3Client(region: string) {
  return new S3Client({ region });
}

export async function putArtifact(
  s3: S3Client,
  {
    bucket,
    key,
    body,
    contentType,
  }: { bucket: string; key: string; body: Uint8Array | string | Buffer; contentType?: string }
): Promise<{ bucket: string; key: string }> {
  await s3.send(new PutObjectCommand({ Bucket: bucket, Key: key, Body: body, ContentType: contentType }));
  return { bucket, key };
}
