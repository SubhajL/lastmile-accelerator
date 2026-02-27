import type { FastifyInstance } from 'fastify';
import { createHash } from 'node:crypto';
import { mkdir, writeFile } from 'node:fs/promises';
import { join } from 'node:path';

import { createSnapshotViaOrchestrator } from '../lib/snapshot-orchestrator-client.js';

type ZipIngestBody = {
  filename: string;
  sizeBytes: number;
  sha256: string;
};

export function registerZipIngestRoutes(
  app: FastifyInstance,
  opts?: { snapshotOrchestratorUrl?: string },
): void {
  const snapshotOrchestratorUrl =
    opts?.snapshotOrchestratorUrl ??
    process.env.SNAPSHOT_ORCHESTRATOR_URL ??
    'http://localhost:7054';
  const uploadMaxBytes = Number(process.env.MAX_UPLOAD_BYTES ?? 50 * 1024 * 1024);

  app.post<{ Params: { projectId: string }; Body: ZipIngestBody }>(
    '/v1/projects/:projectId/ingest/zip',
    {
      schema: {
        params: {
          type: 'object',
          required: ['projectId'],
          properties: { projectId: { type: 'string', minLength: 1 } },
        },
        body: {
          type: 'object',
          required: ['filename', 'sizeBytes', 'sha256'],
          properties: {
            filename: { type: 'string', minLength: 1 },
            sizeBytes: { type: 'integer', minimum: 1 },
            sha256: { type: 'string', pattern: '^[a-f0-9]{64}$' },
          },
        },
      },
    },
    async (req, reply) => {
      try {
        const { snapshotId } = await createSnapshotViaOrchestrator({
          baseUrl: snapshotOrchestratorUrl,
          projectId: req.params.projectId,
          body: {
            mode: 'C',
            sourceRef: {
              zip: {
                filename: req.body.filename,
                sizeBytes: req.body.sizeBytes,
                sha256: req.body.sha256,
              },
            },
          },
        });

        return reply.code(201).send({ snapshotId });
      } catch (err) {
        req.log.error({ err }, 'snapshot-orchestrator request failed');
        return reply.code(502).send({ error: 'snapshot_orchestrator_unavailable' });
      }
    },
  );

  app.post<{ Params: { projectId: string }; Body: Buffer }>(
    '/v1/projects/:projectId/ingest/zip/upload',
    {
      bodyLimit: uploadMaxBytes,
      schema: {
        params: {
          type: 'object',
          required: ['projectId'],
          properties: { projectId: { type: 'string', minLength: 1 } },
        },
      },
    },
    async (req, reply) => {
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
          : 'upload.zip';

      const storageDir = join(process.env.TMPDIR ?? '/tmp', 'lma-zip-uploads');
      await mkdir(storageDir, { recursive: true });
      const storagePath = join(
        storageDir,
        `${Date.now()}-${Math.random().toString(16).slice(2)}.zip`,
      );
      await writeFile(storagePath, req.body);

      const sha256 = createHash('sha256').update(req.body).digest('hex');
      const sizeBytes = req.body.length;

      try {
        const { snapshotId } = await createSnapshotViaOrchestrator({
          baseUrl: snapshotOrchestratorUrl,
          projectId: req.params.projectId,
          body: {
            mode: 'C',
            sourceRef: {
              zip: {
                filename,
                sizeBytes,
                sha256,
                storageKey: `local:${storagePath}`,
              },
            },
          },
        });

        return reply.code(201).send({ snapshotId });
      } catch (err) {
        req.log.error({ err }, 'snapshot-orchestrator request failed');
        return reply.code(502).send({ error: 'snapshot_orchestrator_unavailable' });
      }
    },
  );
}
