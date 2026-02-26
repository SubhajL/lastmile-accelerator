import type { FastifyInstance } from 'fastify';

function isValidSnapshotId(snapshotId: string): boolean {
  return /^snap_[0-9a-f]{32}$/i.test(snapshotId);
}

export function registerFixListRoutes(app: FastifyInstance): void {
  app.get<{ Params: { snapshotId: string } }>(
    '/v1/snapshots/:snapshotId/fix-list',
    {
      schema: {
        params: {
          type: 'object',
          required: ['snapshotId'],
          properties: { snapshotId: { type: 'string', minLength: 1 } },
        },
      },
    },
    async (req, reply) => {
      if (!isValidSnapshotId(req.params.snapshotId)) {
        return reply.code(400).send({ error: 'invalid_snapshot_id' });
      }

      return reply.send({ snapshotId: req.params.snapshotId, fixes: [] });
    },
  );
}
