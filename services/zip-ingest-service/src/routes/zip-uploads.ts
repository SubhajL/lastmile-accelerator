import { createHash, createHmac, randomBytes, timingSafeEqual } from 'node:crypto';
import { mkdir, writeFile } from 'node:fs/promises';
import { join } from 'node:path';

import type { FastifyInstance } from 'fastify';

import { createSnapshotViaOrchestrator } from '../lib/snapshot-orchestrator-client.js';
import type { ZipSourceRefWithSha256 } from '../lib/snapshot-orchestrator-client.js';
import {
  buildZipUploadObjectKey,
  createS3Client,
  createPresignedZipPutUrl,
  getS3ZipUploadsConfigFromEnv,
  headZipUploadSizeBytes,
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

type CompleteZipIngestBody = {
  uploadId: string;
  filename: string;
  sha256?: string;
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

type S3UploadSession = {
  projectId: string;
  bucket: string;
  objectKey: string;
  expiresAtUnixSeconds: number;
};

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
  const s3Client = uploadBackend === 's3' && s3Cfg ? createS3Client(s3Cfg) : null;
  const s3Sessions = new Map<string, S3UploadSession>();
  const s3SessionPruneGraceSeconds = uploadTokenTtlSeconds;

  function pruneStaleS3Sessions(nowUnixSeconds: number): void {
    const hardExpiryUnixSeconds = nowUnixSeconds - s3SessionPruneGraceSeconds;
    for (const [uploadId, session] of s3Sessions.entries()) {
      if (session.expiresAtUnixSeconds < hardExpiryUnixSeconds) {
        s3Sessions.delete(uploadId);
      }
    }
  }

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
      pruneStaleS3Sessions(Math.floor(Date.now() / 1000));

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

        s3Sessions.set(uploadId, {
          projectId: req.params.projectId,
          bucket: s3Cfg.bucket,
          objectKey,
          expiresAtUnixSeconds: expires,
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

  app.post<{ Params: { projectId: string }; Body: CompleteZipIngestBody }>(
    '/v1/projects/:projectId/ingest/zip/complete',
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
          required: ['uploadId', 'filename'],
          properties: {
            uploadId: { type: 'string', minLength: 1 },
            filename: { type: 'string', minLength: 1 },
            sha256: { type: 'string', pattern: '^[a-f0-9]{64}$' },
          },
        },
      },
    },
    async (req, reply) => {
      if (uploadBackend !== 's3') {
        return reply.code(400).send({ error: 'zip_upload_backend_not_s3' });
      }

      if (!s3Cfg || !s3Client) {
        req.log.error('ZIP_UPLOAD_BACKEND=s3 but S3 env is not configured');
        return reply.code(500).send({ error: 's3_not_configured' });
      }

      pruneStaleS3Sessions(Math.floor(Date.now() / 1000));

      const session = s3Sessions.get(req.body.uploadId);
      if (!session || session.projectId !== req.params.projectId) {
        return reply.code(404).send({ error: 'zip_upload_not_found' });
      }
      const now = Math.floor(Date.now() / 1000);
      if (session.expiresAtUnixSeconds < now) {
        s3Sessions.delete(req.body.uploadId);
        return reply.code(401).send({ error: 'zip_upload_expired' });
      }

      let sizeBytes: number | null = null;
      try {
        sizeBytes = await headZipUploadSizeBytes({
          client: s3Client,
          cfg: s3Cfg,
          objectKey: session.objectKey,
        });
      } catch (err) {
        req.log.error({ err }, 's3 head object failed');
        return reply.code(502).send({ error: 'zip_storage_unavailable' });
      }
      if (sizeBytes === null) {
        return reply.code(404).send({ error: 'zip_upload_not_found' });
      }
      if (sizeBytes < 1) {
        return reply.code(400).send({ error: 'zip_upload_empty' });
      }
      if (sizeBytes > uploadMaxBytes) {
        return reply.code(413).send({ error: 'zip_upload_too_large' });
      }

      const storageKey = `s3://${session.bucket}/${session.objectKey}`;

      try {
        const { snapshotId } = await createSnapshotViaOrchestrator({
          baseUrl: snapshotOrchestratorUrl,
          projectId: req.params.projectId,
          body: {
            mode: 'C',
            sourceRef: {
              zip: {
                filename: req.body.filename,
                sizeBytes,
                sha256: req.body.sha256,
                storageKey,
                uploadId: req.body.uploadId,
              },
            },
          },
        });

        s3Sessions.delete(req.body.uploadId);
        return reply.code(201).send({ snapshotId });
      } catch (err) {
        req.log.error({ err }, 'snapshot-orchestrator request failed');
        return reply.code(502).send({ error: 'snapshot_orchestrator_unavailable' });
      }
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
      const zipSourceRef: ZipSourceRefWithSha256 = {
        filename,
        sizeBytes,
        sha256,
        storageKey: `local:${storagePath}`,
        uploadId,
      };

      try {
        const { snapshotId } = await createSnapshotViaOrchestrator({
          baseUrl: snapshotOrchestratorUrl,
          projectId,
          body: {
            mode: 'C',
            sourceRef: {
              zip: zipSourceRef,
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
