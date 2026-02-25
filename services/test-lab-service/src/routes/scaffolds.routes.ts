import type { FastifyInstance, FastifyReply, FastifyRequest } from 'fastify';
import { z } from 'zod';
import { requireAuth, requireScopes } from '../middleware/auth.js';
import { CreateScaffoldSchema, UpdateScaffoldSchema, ListScaffoldsQuerySchema } from '../schemas/scaffolds.js';

const ParamsWithProject = z.object({ projectId: z.string().uuid() });
const ParamsWithId = z.object({ id: z.string().uuid() });

export function registerScaffoldRoutes(app: FastifyInstance) {
  // Create scaffold
  app.post('/v1/projects/:projectId/test-scaffolds', {
    preHandler: [requireAuth, requireScopes('scaffold:write')],
    schema: {},
  }, async (request: FastifyRequest, reply: FastifyReply) => {
    const params = ParamsWithProject.parse(request.params);
    const body = CreateScaffoldSchema.parse(request.body);
    const rec = await app.repos.scaffolds.create(params.projectId, body);
    return reply.code(201).send(rec);
  });

  // List scaffolds
  app.get('/v1/projects/:projectId/test-scaffolds', {
    preHandler: [requireAuth, requireScopes('scaffold:read')],
  }, async (request, reply) => {
    const params = ParamsWithProject.parse(request.params);
    const query = ListScaffoldsQuerySchema.parse(request.query);
    const { items, nextCursor } = await app.repos.scaffolds.listByProject(params.projectId, query.limit!, query.cursor);
    return { items, nextCursor };
  });

  // Get scaffold by id
  app.get('/v1/test-scaffolds/:id', {
    preHandler: [requireAuth, requireScopes('scaffold:read')],
  }, async (request, reply) => {
    const { id } = ParamsWithId.parse(request.params);
    const rec = await app.repos.scaffolds.getById(id);
    if (!rec) return reply.code(404).send({ error: 'Not Found' });
    return rec;
  });

  // Update scaffold
  app.put('/v1/test-scaffolds/:id', {
    preHandler: [requireAuth, requireScopes('scaffold:write')],
  }, async (request, reply) => {
    const { id } = ParamsWithId.parse(request.params);
    const patch = UpdateScaffoldSchema.parse(request.body);
    const rec = await app.repos.scaffolds.update(id, patch);
    if (!rec) return reply.code(404).send({ error: 'Not Found' });
    return rec;
  });

  // Delete scaffold
  app.delete('/v1/test-scaffolds/:id', {
    preHandler: [requireAuth, requireScopes('scaffold:write')],
  }, async (request, reply) => {
    const { id } = ParamsWithId.parse(request.params);
    const ok = await app.repos.scaffolds.delete(id);
    if (!ok) return reply.code(404).send({ error: 'Not Found' });
    return reply.code(204).send();
  });
}
