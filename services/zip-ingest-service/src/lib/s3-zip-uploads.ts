import { Sha256 } from '@aws-crypto/sha256-js';
import { formatUrl } from '@aws-sdk/util-format-url';
import { HttpRequest } from '@smithy/protocol-http';
import { SignatureV4 } from '@smithy/signature-v4';
import { HeadObjectCommand, S3Client } from '@aws-sdk/client-s3';

export type S3ZipUploadsConfig = {
  bucket: string;
  endpoint: string;
  region: string;
  accessKeyId: string;
  secretAccessKey: string;
  forcePathStyle: boolean;
  prefix: string;
};

function parseBoolean(value: string | undefined, defaultValue: boolean): boolean {
  if (value === undefined) return defaultValue;
  const normalized = value.trim().toLowerCase();
  if (normalized === 'true' || normalized === '1' || normalized === 'yes') return true;
  if (normalized === 'false' || normalized === '0' || normalized === 'no') return false;
  return defaultValue;
}

function normalizePrefix(prefix: string): string {
  if (!prefix) return '';
  return prefix.endsWith('/') ? prefix : `${prefix}/`;
}

export function getS3ZipUploadsConfigFromEnv(): S3ZipUploadsConfig | null {
  const bucket = process.env.SNAPSHOT_BUCKET;
  const endpoint = process.env.SNAPSHOT_S3_ENDPOINT;
  const accessKeyId = process.env.SNAPSHOT_S3_ACCESS_KEY;
  const secretAccessKey = process.env.SNAPSHOT_S3_SECRET_KEY;

  if (!bucket || !endpoint || !accessKeyId || !secretAccessKey) return null;

  const region = process.env.SNAPSHOT_S3_REGION ?? 'us-east-1';
  const forcePathStyle = parseBoolean(process.env.SNAPSHOT_S3_FORCE_PATH_STYLE, true);
  const prefix = normalizePrefix(process.env.ZIP_UPLOAD_PREFIX ?? 'zip-uploads/');

  return { bucket, endpoint, region, accessKeyId, secretAccessKey, forcePathStyle, prefix };
}

export function buildZipUploadObjectKey(args: {
  prefix: string;
  projectId: string;
  uploadId: string;
}): string {
  return `${normalizePrefix(args.prefix)}${encodeURIComponent(args.projectId)}/${args.uploadId}.zip`;
}

export function createS3Client(cfg: S3ZipUploadsConfig): S3Client {
  return new S3Client({
    region: cfg.region,
    endpoint: cfg.endpoint,
    forcePathStyle: cfg.forcePathStyle,
    credentials: {
      accessKeyId: cfg.accessKeyId,
      secretAccessKey: cfg.secretAccessKey,
    },
  });
}

export async function createPresignedZipPutUrl(args: {
  cfg: S3ZipUploadsConfig;
  objectKey: string;
  expiresInSeconds: number;
}): Promise<string> {
  const endpoint = new URL(args.cfg.endpoint);
  const hostname = endpoint.hostname;
  const port = endpoint.port ? Number(endpoint.port) : undefined;
  const basePath = endpoint.pathname && endpoint.pathname !== '/' ? endpoint.pathname : '';

  const pathStylePath = `${basePath}/${encodeURIComponent(args.cfg.bucket)}/${args.objectKey}`;
  const path = args.cfg.forcePathStyle ? pathStylePath : `${basePath}/${args.objectKey}`;

  const request = new HttpRequest({
    protocol: endpoint.protocol,
    hostname,
    port,
    method: 'PUT',
    path,
    headers: {
      host: port ? `${hostname}:${port}` : hostname,
      'content-type': 'application/zip',
    },
  });

  const signer = new SignatureV4({
    credentials: {
      accessKeyId: args.cfg.accessKeyId,
      secretAccessKey: args.cfg.secretAccessKey,
    },
    region: args.cfg.region,
    service: 's3',
    sha256: Sha256,
  });

  const signed = await signer.presign(request, { expiresIn: args.expiresInSeconds });
  return formatUrl(signed);
}

export async function headZipUploadSizeBytes(args: {
  client: S3Client;
  cfg: S3ZipUploadsConfig;
  objectKey: string;
}): Promise<number | null> {
  try {
    const res = await args.client.send(
      new HeadObjectCommand({ Bucket: args.cfg.bucket, Key: args.objectKey }),
    );
    const sizeBytes = res.ContentLength;
    return typeof sizeBytes === 'number' ? sizeBytes : null;
  } catch (err) {
    const statusCode =
      typeof err === 'object' && err !== null
        ? (err as { $metadata?: { httpStatusCode?: number } }).$metadata?.httpStatusCode
        : null;
    if (statusCode === 404) return null;
    throw err;
  }
}
