import type { FastifyInstance } from 'fastify';

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
            sizeBytes: { type: 'integer', minimum: 0 },
            sha256: { type: 'string', minLength: 1 },
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

        return reply.send({ snapshotId });
      } catch {
        return reply.code(502).send({ error: 'snapshot_orchestrator_unavailable' });
      }
    },
  );
}
