import type { FastifyInstance, FastifyRequest, FastifyReply } from 'fastify';
import type { AppDeps } from '../http-routes.js';
import { requireRoleGuard } from '../auth/jwt.js';

interface RetryParams {
  notificationId: string;
}

export function adminRetryRoute(app: FastifyInstance, deps: AppDeps): void {
  app.post<{ Params: RetryParams }>(
    '/retry/:notificationId',
    {
      preHandler: requireRoleGuard('admin'),
      schema: {
        params: {
          type: 'object',
          properties: {
            notificationId: { type: 'string' },
          },
          required: ['notificationId'],
        },
      },
    },
    async (request: FastifyRequest<{ Params: RetryParams }>, reply: FastifyReply) => {
      const { notificationId } = request.params;

      if (!notificationId) {
        return reply.code(400).send({ error: 'notificationId is required' });
      }

      const result = await deps.queue.retryNotification(notificationId);

      if (result.success) {
        return reply.send({ status: 'retried', notificationId });
      }

      if (result.error === 'not_found') {
        return reply.code(404).send({ error: 'Notification not found' });
      }

      return reply.code(500).send({ error: 'Failed to retry notification' });
    },
  );
}
