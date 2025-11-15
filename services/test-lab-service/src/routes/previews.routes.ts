import type { FastifyInstance, FastifyReply, FastifyRequest } from 'fastify';
import { z } from 'zod';
import { requireAuth, requireScopes } from '../middleware/auth.js';
import { CreatePreviewSchema, ListPreviewsQuerySchema, UpdatePreviewSchema } from '../schemas/previews.js';

const ParamsWithProject = z.object({ projectId: z.string().uuid() });
const ParamsWithId = z.object({ id: z.string().uuid() });

export function registerPreviewRoutes(app: FastifyInstance) {
  // Create preview env
  app.post(
    '/v1/projects/:projectId/previews',
    { preHandler: [requireAuth, requireScopes('preview:write')] },
    async (request: FastifyRequest, reply: FastifyReply) => {
      const { projectId } = ParamsWithProject.parse(request.params);
      const body = CreatePreviewSchema.parse(request.body);
      const rec = await app.repos.previewEnvs!.create(projectId, body);
      return reply.code(201).send(rec);
    }
  );

  // List previews
  app.get(
    '/v1/projects/:projectId/previews',
    { preHandler: [requireAuth, requireScopes('preview:read')] },
    async (request, reply) => {
      const { projectId } = ParamsWithProject.parse(request.params);
      const query = ListPreviewsQuerySchema.parse(request.query);
      const { items, nextCursor } = await app.repos.previewEnvs!.listByProject(projectId, { limit: query.limit!, cursor: query.cursor });
      return { items, nextCursor };
    }
  );

  // Get preview by id
  app.get(
    '/v1/previews/:id',
    { preHandler: [requireAuth, requireScopes('preview:read')] },
    async (request, reply) => {
      const { id } = ParamsWithId.parse(request.params);
      const rec = await app.repos.previewEnvs!.getById(id);
      if (!rec) return reply.code(404).send({ error: 'Not Found' });
      return rec;
    }
  );

  // Update preview
  app.put(
    '/v1/previews/:id',
    { preHandler: [requireAuth, requireScopes('preview:write')] },
    async (request, reply) => {
      const { id } = ParamsWithId.parse(request.params);
      const patch = UpdatePreviewSchema.parse(request.body);
      const rec = await app.repos.previewEnvs!.update(id, patch);
      if (!rec) return reply.code(404).send({ error: 'Not Found' });
      return rec;
    }
  );

  // Delete preview
  app.delete(
    '/v1/previews/:id',
    { preHandler: [requireAuth, requireScopes('preview:write')] },
    async (request, reply) => {
      const { id } = ParamsWithId.parse(request.params);
      const ok = await app.repos.previewEnvs!.delete(id);
      if (!ok) return reply.code(404).send({ error: 'Not Found' });
      return reply.code(204).send();
    }
  );
}
