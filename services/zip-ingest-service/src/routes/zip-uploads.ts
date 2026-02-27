import { createHash, createHmac, randomBytes, timingSafeEqual } from 'node:crypto';
import { mkdir, writeFile } from 'node:fs/promises';
import { join } from 'node:path';

import type { FastifyInstance } from 'fastify';

import { createSnapshotViaOrchestrator } from '../lib/snapshot-orchestrator-client.js';
import {
  buildZipUploadObjectKey,
  createPresignedZipPutUrl,
  getS3ZipUploadsConfigFromEnv,
} from '../lib/s3-zip-uploads.js';

type InitiateZipIngestBody = {
  filename?: string;
};

type UploadQuery = {
  expires: number;
  token: string;
};

type UploadParams = {
  projectId: string;
  uploadId: string;
};

function hmacToken(args: { secret: string; uploadId: string; expires: number }): string {
  const mac = createHmac('sha256', args.secret)
    .update(`${args.uploadId}:${args.expires}`)
    .digest('hex');
  return mac;
}

function isValidToken(args: {
  secret: string;
  uploadId: string;
  expires: number;
  token: string;
}): boolean {
  const expected = hmacToken({
    secret: args.secret,
    uploadId: args.uploadId,
    expires: args.expires,
  });
  const token = args.token.toLowerCase();
  if (token.length !== expected.length) return false;
  return timingSafeEqual(Buffer.from(token, 'utf8'), Buffer.from(expected, 'utf8'));
}

function sha256Hex(buf: Buffer): string {
  return createHash('sha256').update(buf).digest('hex');
}

function newUploadId(): string {
  return `upl_${randomBytes(16).toString('hex')}`;
}

type UploadBackend = 'service' | 's3';

function parseUploadBackend(value: string | undefined): UploadBackend {
  if (!value) return 'service';
  const normalized = value.trim().toLowerCase();
  if (normalized === 's3') return 's3';
  return 'service';
}

export function registerZipUploadRoutes(
  app: FastifyInstance,
  opts?: {
    snapshotOrchestratorUrl?: string;
    uploadBaseUrl?: string;
    uploadSigningSecret?: string;
    uploadTokenTtlSeconds?: number;
    uploadMaxBytes?: number;
  },
): void {
  const snapshotOrchestratorUrl =
    opts?.snapshotOrchestratorUrl ??
    process.env.SNAPSHOT_ORCHESTRATOR_URL ??
    'http://localhost:7054';

  const uploadBaseUrl =
    opts?.uploadBaseUrl ?? process.env.UPLOAD_BASE_URL ?? 'http://localhost:7052';

  const uploadSigningSecret =
    opts?.uploadSigningSecret ??
    process.env.ZIP_UPLOAD_SIGNING_SECRET ??
    randomBytes(32).toString('hex');

  const uploadTokenTtlSeconds =
    opts?.uploadTokenTtlSeconds ?? Number(process.env.ZIP_UPLOAD_TTL_SEC ?? 15 * 60);
  const uploadMaxBytes =
    opts?.uploadMaxBytes ?? Number(process.env.MAX_UPLOAD_BYTES ?? 50 * 1024 * 1024);

  const uploadBackend = parseUploadBackend(process.env.ZIP_UPLOAD_BACKEND);
  const s3Cfg = uploadBackend === 's3' ? getS3ZipUploadsConfigFromEnv() : null;

  app.post<{ Params: { projectId: string }; Body: InitiateZipIngestBody }>(
    '/v1/projects/:projectId/ingest/zip/initiate',
    {
      schema: {
        params: {
          type: 'object',
          required: ['projectId'],
          properties: { projectId: { type: 'string', minLength: 1 } },
        },
        body: {
          type: 'object',
          additionalProperties: false,
          properties: {
            filename: { type: 'string', minLength: 1 },
          },
        },
      },
    },
    async (req, reply) => {
      const uploadId = newUploadId();
      const expires = Math.floor(Date.now() / 1000) + uploadTokenTtlSeconds;
      const filename = req.body?.filename ?? `${uploadId}.zip`;

      if (uploadBackend === 's3') {
        if (!s3Cfg) {
          req.log.error('ZIP_UPLOAD_BACKEND=s3 but S3 env is not configured');
          return reply.code(500).send({ error: 's3_not_configured' });
        }

        const objectKey = buildZipUploadObjectKey({
          prefix: s3Cfg.prefix,
          projectId: req.params.projectId,
          uploadId,
        });

        const uploadUrl = await createPresignedZipPutUrl({
          cfg: s3Cfg,
          objectKey,
          expiresInSeconds: uploadTokenTtlSeconds,
        });

        return reply.code(201).send({
          uploadId,
          uploadUrl,
          uploadBackend: 's3',
          bucket: s3Cfg.bucket,
          objectKey,
          method: 'PUT',
          contentType: 'application/zip',
          expiresAt: new Date(expires * 1000).toISOString(),
          filename,
          maxBytes: uploadMaxBytes,
        });
      }

      const token = hmacToken({ secret: uploadSigningSecret, uploadId, expires });
      const uploadUrl = `${uploadBaseUrl}/v1/projects/${req.params.projectId}/ingest/zip/uploads/${uploadId}?expires=${expires}&token=${token}`;

      return reply.code(201).send({
        uploadId,
        uploadUrl,
        uploadBackend: 'service',
        method: 'PUT',
        contentType: 'application/zip',
        expiresAt: new Date(expires * 1000).toISOString(),
        filename,
        maxBytes: uploadMaxBytes,
      });
    },
  );

  app.put<{ Params: UploadParams; Querystring: UploadQuery; Body: Buffer }>(
    '/v1/projects/:projectId/ingest/zip/uploads/:uploadId',
    {
      bodyLimit: uploadMaxBytes,
      schema: {
        params: {
          type: 'object',
          required: ['projectId', 'uploadId'],
          properties: {
            projectId: { type: 'string', minLength: 1 },
            uploadId: { type: 'string', minLength: 1 },
          },
        },
        querystring: {
          type: 'object',
          required: ['expires', 'token'],
          properties: {
            expires: { type: 'integer', minimum: 1 },
            token: { type: 'string', minLength: 1 },
          },
        },
      },
    },
    async (req, reply) => {
      const { projectId, uploadId } = req.params;
      const { expires, token } = req.query;
      const now = Math.floor(Date.now() / 1000);
      if (expires < now) {
        return reply.code(401).send({ error: 'upload_token_expired' });
      }
      if (!isValidToken({ secret: uploadSigningSecret, uploadId, expires, token })) {
        return reply.code(401).send({ error: 'invalid_upload_token' });
      }
      if (!Buffer.isBuffer(req.body)) {
        return reply.code(415).send({ error: 'unsupported_media_type' });
      }
      if (req.body.length < 1) {
        return reply.code(400).send({ error: 'empty_upload' });
      }

      const filenameHeader = req.headers['x-filename'];
      const filename =
        typeof filenameHeader === 'string' && filenameHeader.trim()
          ? filenameHeader.trim()
          : `${uploadId}.zip`;

      const storageDir = join(process.env.TMPDIR ?? '/tmp', 'lma-zip-uploads');
      await mkdir(storageDir, { recursive: true });
      const storagePath = join(storageDir, `${uploadId}.zip`);
      await writeFile(storagePath, req.body);

      const sha256 = sha256Hex(req.body);
      const sizeBytes = req.body.length;

      try {
        const { snapshotId } = await createSnapshotViaOrchestrator({
          baseUrl: snapshotOrchestratorUrl,
          projectId,
          body: {
            mode: 'C',
            sourceRef: {
              zip: {
                filename,
                sizeBytes,
                sha256,
                storageKey: `local:${storagePath}`,
                uploadId,
              },
            },
          },
        });

        return reply.code(201).send({ snapshotId, sha256, sizeBytes });
      } catch (err) {
        req.log.error({ err }, 'snapshot-orchestrator request failed');
        return reply.code(502).send({ error: 'snapshot_orchestrator_unavailable' });
      }
    },
  );
}
