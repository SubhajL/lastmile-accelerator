import type { FastifyInstance, FastifyReply, FastifyRequest } from 'fastify';
import { z } from 'zod';
import { requireAuth, requireScopes } from '../middleware/auth.js';
import {
  CreateTestRunSchema,
  ListTestRunsQuerySchema,
  UpdateTestRunStatusSchema,
  CreateBrowserTestRunSchema,
  UpdateBrowserRunStatusSchema,
} from '../schemas/test-runs.js';

const ParamsWithProject = z.object({ projectId: z.string().uuid() });
const ParamsWithId = z.object({ id: z.string().uuid() });
const ParamsWithRunId = z.object({ testRunId: z.string().uuid() });

export function registerTestRunRoutes(app: FastifyInstance) {
  // Test Runs
  app.post(
    '/v1/projects/:projectId/test-runs',
    { preHandler: [requireAuth, requireScopes('run:write')] },
    async (request: FastifyRequest, reply: FastifyReply) => {
      const { projectId } = ParamsWithProject.parse(request.params);
      const body = CreateTestRunSchema.parse(request.body);
      const rec = await app.repos.testRuns!.create(projectId, body);
      return reply.code(201).send(rec);
    }
  );

  app.get(
    '/v1/projects/:projectId/test-runs',
    { preHandler: [requireAuth, requireScopes('run:read')] },
    async (request, reply) => {
      const { projectId } = ParamsWithProject.parse(request.params);
      const query = ListTestRunsQuerySchema.parse(request.query);
      const page = await app.repos.testRuns!.listByProject(projectId, {
        limit: query.limit!,
        cursor: query.cursor,
        status: query.status,
      });
      return page;
    }
  );

  app.get(
    '/v1/test-runs/:id',
    { preHandler: [requireAuth, requireScopes('run:read')] },
    async (request, reply) => {
      const { id } = ParamsWithId.parse(request.params);
      const rec = await app.repos.testRuns!.getById(id);
      if (!rec) return reply.code(404).send({ error: 'Not Found' });
      return rec;
    }
  );

  app.patch(
    '/v1/test-runs/:id/status',
    { preHandler: [requireAuth, requireScopes('run:write')] },
    async (request, reply) => {
      const { id } = ParamsWithId.parse(request.params);
      const body = UpdateTestRunStatusSchema.parse(request.body);
      const rec = await app.repos.testRuns!.updateStatus(id, body.status, {
        startedAt: body.startedAt,
        finishedAt: body.finishedAt,
        results: body.results,
        artifacts: body.artifacts,
      });
      if (!rec) return reply.code(404).send({ error: 'Not Found' });
      return rec;
    }
  );

  // Browser Test Runs
  app.post(
    '/v1/test-runs/:testRunId/browser-runs',
    { preHandler: [requireAuth, requireScopes('browser-run:write')] },
    async (request, reply) => {
      const { testRunId } = ParamsWithRunId.parse(request.params);
      const body = CreateBrowserTestRunSchema.parse(request.body);
      const rec = await app.repos.browserTestRuns!.create(testRunId, body);
      return reply.code(201).send(rec);
    }
  );

  app.get(
    '/v1/test-runs/:testRunId/browser-runs',
    { preHandler: [requireAuth, requireScopes('browser-run:read')] },
    async (request, reply) => {
      const { testRunId } = ParamsWithRunId.parse(request.params);
      const page = await app.repos.browserTestRuns!.listByTestRun(testRunId);
      return page;
    }
  );

  app.get(
    '/v1/browser-runs/:id',
    { preHandler: [requireAuth, requireScopes('browser-run:read')] },
    async (request, reply) => {
      const { id } = ParamsWithId.parse(request.params);
      const rec = await app.repos.browserTestRuns!.getById(id);
      if (!rec) return reply.code(404).send({ error: 'Not Found' });
      return rec;
    }
  );

  app.patch(
    '/v1/browser-runs/:id/status',
    { preHandler: [requireAuth, requireScopes('browser-run:write')] },
    async (request, reply) => {
      const { id } = ParamsWithId.parse(request.params);
      const body = UpdateBrowserRunStatusSchema.parse(request.body);
      const rec = await app.repos.browserTestRuns!.updateStatus(id, body.status, {
        startedAt: body.startedAt,
        finishedAt: body.finishedAt,
        screenshots: body.screenshots,
        logs: body.logs,
      });
      if (!rec) return reply.code(404).send({ error: 'Not Found' });
      return rec;
    }
  );
}
