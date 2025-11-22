import type { FastifyRequest, FastifyReply } from 'fastify';
import type { AppDeps } from '../http-routes.js';

export function healthzRoute(deps: AppDeps) {
  return async function (_request: FastifyRequest, reply: FastifyReply) {
    const healthy = (await deps.dbHealthCheck?.()) ?? true;

    if (!healthy) {
      return reply.code(503).send({
        status: 'unhealthy',
        service: deps.config?.service?.name ?? 'notification-service',
        timestamp: new Date().toISOString(),
      });
    }

    return reply.code(200).send({
      status: 'ok',
      timestamp: new Date().toISOString(),
    });
  };
}
