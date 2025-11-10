import type { FastifyInstance } from 'fastify';
import { listProjects as listSvc, getProject as getSvc, createProject as createSvc, updateProject as updateSvc, deleteProject as deleteSvc } from '../services/projectService';
import { parseCreateProject, parseUpdateProject } from '../utils/validators';
import { extractTenantId as extractTenantIdDefault, requireScopes } from '../middleware/auth';

export function registerProjectRoutes(app: FastifyInstance, opts?: {
  extractTenantId?: (req: any) => string;
  services?: {
    listProjects: typeof listSvc;
    getProject: typeof getSvc;
    createProject: typeof createSvc;
    updateProject: typeof updateSvc;
    deleteProject: typeof deleteSvc;
  };
}) {
  const extractTenantId = opts?.extractTenantId ?? extractTenantIdDefault;
  const services = opts?.services ?? {
    listProjects: listSvc,
    getProject: getSvc,
    createProject: createSvc,
    updateProject: updateSvc,
    deleteProject: deleteSvc,
  };

  app.get('/v1/projects', { preHandler: [requireScopes('projects:read')] }, async (req, reply) => {
    const tenantId = extractTenantId(req);
    const rows = await services.listProjects(tenantId);
    return reply.status(200).send(rows);
  });

  app.post('/v1/projects', { preHandler: [requireScopes('projects:write')] }, async (req: any, reply) => {
    const tenantId = extractTenantId(req);
    const body = parseCreateProject(req.body);
    const userId = req.user?.sub ?? 'unknown';
    const row = await services.createProject(tenantId, body, userId);
    return reply.status(201).send(row);
  });

  app.get('/v1/projects/:projectId', { preHandler: [requireScopes('projects:read')] }, async (req: any, reply) => {
    const tenantId = extractTenantId(req);
    const row = await services.getProject(tenantId, req.params.projectId);
    return reply.status(200).send(row);
  });

  app.put('/v1/projects/:projectId', { preHandler: [requireScopes('projects:write')] }, async (req: any, reply) => {
    const tenantId = extractTenantId(req);
    const body = parseUpdateProject(req.body);
    const row = await services.updateProject(tenantId, req.params.projectId, body);
    return reply.status(200).send(row);
  });

  app.delete('/v1/projects/:projectId', { preHandler: [requireScopes('projects:write')] }, async (req: any, reply) => {
    const tenantId = extractTenantId(req);
    await services.deleteProject(tenantId, req.params.projectId);
    return reply.status(204).send();
  });
}