import type { FastifyInstance } from 'fastify';
import {
  listEnvironments as listSvc,
  getEnvironment as getSvc,
  createEnvironment as createSvc,
  updateEnvironment as updateSvc,
  deleteEnvironment as deleteSvc,
  setIngestionModes as setModesSvc,
} from '../services/environmentService';
import { extractTenantId as extractTenantIdDefault, requireScopes } from '../middleware/auth';
import { parseCreateEnvironment, parseSetIngestionModes } from '../utils/validators';

export function registerEnvironmentRoutes(app: FastifyInstance, opts?: {
  extractTenantId?: (req: any) => string;
  services?: {
    listEnvironments: typeof listSvc;
    getEnvironment: typeof getSvc;
    createEnvironment: typeof createSvc;
    updateEnvironment: typeof updateSvc;
    deleteEnvironment: typeof deleteSvc;
    setIngestionModes: typeof setModesSvc;
  };
}) {
  const extractTenantId = opts?.extractTenantId ?? extractTenantIdDefault;
  const services = opts?.services ?? {
    listEnvironments: listSvc,
    getEnvironment: getSvc,
    createEnvironment: createSvc,
    updateEnvironment: updateSvc,
    deleteEnvironment: deleteSvc,
    setIngestionModes: setModesSvc,
  };

  app.get('/v1/projects/:projectId/environments', { preHandler: [requireScopes('environments:read')] }, async (req: any, reply) => {
    const tenantId = extractTenantId(req);
    const rows = await services.listEnvironments(tenantId, req.params.projectId);
    return reply.status(200).send(rows);
  });

  app.post('/v1/projects/:projectId/environments', { preHandler: [requireScopes('environments:write')] }, async (req: any, reply) => {
    const tenantId = extractTenantId(req);
    const body = parseCreateEnvironment(req.body);
    const row = await services.createEnvironment(tenantId, req.params.projectId, body as any);
    return reply.status(201).send(row);
  });

  app.put('/v1/projects/:projectId/environments/:envId', { preHandler: [requireScopes('environments:write')] }, async (req: any, reply) => {
    const tenantId = extractTenantId(req);
    const row = await services.updateEnvironment(tenantId, req.params.projectId, req.params.envId, req.body ?? {});
    return reply.status(200).send(row);
  });

  app.delete('/v1/projects/:projectId/environments/:envId', { preHandler: [requireScopes('environments:write')] }, async (req: any, reply) => {
    const tenantId = extractTenantId(req);
    await services.deleteEnvironment(tenantId, req.params.projectId, req.params.envId);
    return reply.status(204).send();
  });

  app.post('/v1/projects/:projectId/ingestion-modes', { preHandler: [requireScopes('environments:write')] }, async (req: any, reply) => {
    const tenantId = extractTenantId(req);
    const body = parseSetIngestionModes(req.body);
    await services.setIngestionModes(tenantId, req.params.projectId, body as any);
    return reply.status(200).send({ ok: true });
  });
}