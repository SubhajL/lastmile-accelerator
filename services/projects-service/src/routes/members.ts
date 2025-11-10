import type { FastifyInstance } from 'fastify';
import { addMember as addSvc, updateMemberRole as updateRoleSvc, removeMember as removeSvc } from '../services/memberService';
import { parseCreateMember } from '../utils/validators';
import { requireScopes, requireOwner } from '../middleware/auth';

export function registerMemberRoutes(app: FastifyInstance, opts?: {
  services?: {
    addMember: typeof addSvc;
    updateMemberRole: typeof updateRoleSvc;
    removeMember: typeof removeSvc;
  };
}) {
  const services = opts?.services ?? { addMember: addSvc, updateMemberRole: updateRoleSvc, removeMember: removeSvc };

  app.post('/v1/tenants/:tenantId/members', { preHandler: [requireScopes('members:write'), requireOwner()] }, async (req: any, reply) => {
    const body = parseCreateMember(req.body);
    const row = await services.addMember(req.params.tenantId, body as any, req.user?.role ?? 'owner');
    return reply.status(201).send(row);
  });

  app.put('/v1/members/:memberId', { preHandler: [requireScopes('members:write'), requireOwner()] }, async (req: any, reply) => {
    const role = req.body?.role as string;
    const row = await services.updateMemberRole(req.query?.tenantId ?? 't1', req.params.memberId, role, req.user?.role ?? 'owner');
    return reply.status(200).send(row);
  });

  app.delete('/v1/members/:memberId', { preHandler: [requireScopes('members:write'), requireOwner()] }, async (req: any, reply) => {
    await services.removeMember(req.query?.tenantId ?? 't1', req.params.memberId, req.user?.role ?? 'owner');
    return reply.status(204).send();
  });
}