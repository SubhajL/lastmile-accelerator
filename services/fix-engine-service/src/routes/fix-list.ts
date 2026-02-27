import type { FastifyInstance } from 'fastify';

import { getSnapshotOrchestratorSnapshot } from '../lib/snapshot-orchestrator-client.js';

function isValidSnapshotId(snapshotId: string): boolean {
  return /^snap_[0-9a-f]{32}$/.test(snapshotId);
}

export function registerFixListRoutes(
  app: FastifyInstance,
  opts?: { snapshotOrchestratorUrl?: string },
): void {
  const snapshotOrchestratorUrl =
    opts?.snapshotOrchestratorUrl ??
    process.env.SNAPSHOT_ORCHESTRATOR_URL ??
    'http://localhost:7054';

  app.get<{ Params: { snapshotId: string } }>(
    '/v1/snapshots/:snapshotId/fix-list',
    {
      schema: {
        params: {
          type: 'object',
          required: ['snapshotId'],
          properties: {
            snapshotId: {
              type: 'string',
              minLength: 1,
            },
          },
        },
      },
    },
    async (req, reply) => {
      if (!isValidSnapshotId(req.params.snapshotId)) {
        return reply.code(400).send({ error: 'invalid_snapshot_id' });
      }

      const snapshot = await getSnapshotOrchestratorSnapshot({
        baseUrl: snapshotOrchestratorUrl,
        snapshotId: req.params.snapshotId,
      });
      if (snapshot === 'not_found') {
        return reply.code(404).send({ error: 'snapshot_not_found' });
      }
      if (snapshot === 'unavailable') {
        return reply.code(502).send({ error: 'snapshot_orchestrator_unavailable' });
      }

      const fixId = `fix_${req.params.snapshotId}`;
      return reply.send({
        snapshotId: req.params.snapshotId,
        fixes: [
          {
            id: fixId,
            tier: 3,
            title: 'MVP placeholder fix',
            description: 'This is a deterministic placeholder fix-list entry.',
          },
        ],
      });
    },
  );
}
