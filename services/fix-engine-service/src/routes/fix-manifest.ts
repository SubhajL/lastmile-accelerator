import type { FastifyInstance } from 'fastify';

export function registerFixManifestRoutes(app: FastifyInstance): void {
  app.get<{ Params: { fixId: string } }>(
    '/v1/fixes/:fixId/manifest',
    {
      schema: {
        params: {
          type: 'object',
          required: ['fixId'],
          properties: {
            fixId: { type: 'string', minLength: 1 },
          },
        },
      },
    },
    async (req) => ({
      fixId: req.params.fixId,
      summary: 'MVP placeholder manifest',
      files: [],
      locDelta: 0,
      tests: [],
    }),
  );
}
