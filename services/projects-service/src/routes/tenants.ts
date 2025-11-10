import type { FastifyInstance } from 'fastify';
import { getTenant as getSvc, listTenantMembers as listSvc } from '../services/tenantService';
import { extractTenantId as extractTenantIdDefault, requireScopes } from '../middleware/auth';

export function registerTenantRoutes(app: FastifyInstance, opts?: {
  extractTenantId?: (req: any) => string;
  services?: {
    getTenant: typeof getSvc;
    listTenantMembers: typeof listSvc;
  };
}) {
  const extractTenantId = opts?.extractTenantId ?? extractTenantIdDefault;
  const services = opts?.services ?? { getTenant: getSvc, listTenantMembers: listSvc };

  app.get('/v1/tenants/:tenantId', { preHandler: [requireScopes('tenants:read')] }, async (req: any, reply) => {
    const jwtTenant = extractTenantId(req);
    if (req.params.tenantId !== jwtTenant) return reply.status(403).send({ code: 'FORBIDDEN' });
    const row = await services.getTenant(req.params.tenantId);
    return reply.status(200).send(row);
  });

  app.get('/v1/tenants/:tenantId/members', { preHandler: [requireScopes('tenants:read')] }, async (req: any, reply) => {
    const jwtTenant = extractTenantId(req);
    if (req.params.tenantId !== jwtTenant) return reply.status(403).send({ code: 'FORBIDDEN' });
    const rows = await services.listTenantMembers(req.params.tenantId);
    return reply.status(200).send(rows);
  });
}
