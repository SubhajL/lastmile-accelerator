import type { FastifyInstance } from 'fastify';
import { requireRoleGuard } from '../auth/jwt.js';

export interface AdminDeps {
  enqueue: (job: {
    tenantId: string;
    userId: string;
    channel: string;
    templateName: string;
    priority: 'critical' | 'high' | 'normal' | 'low';
    payload: Record<string, unknown>;
    maxAttempts: number;
  }) => Promise<string>;
}

const adminSendTestBodySchema = {
  type: 'object',
  required: ['tenantId', 'userId'],
  additionalProperties: false,
  properties: {
    tenantId: { type: 'string', minLength: 1 },
    userId: { type: 'string', minLength: 1 },
    channel: {
      type: 'string',
      enum: ['email', 'sms', 'slack', 'webhook', 'in-app'],
    },
    templateName: { type: 'string', minLength: 1 },
    payload: { type: 'object', additionalProperties: true },
    priority: {
      type: 'string',
      enum: ['critical', 'high', 'normal', 'low'],
    },
    maxAttempts: { type: 'integer', minimum: 1, maximum: 10 },
  },
} as const;

export function adminSendTestRoute(app: FastifyInstance, deps: AdminDeps) {
  app.post(
    '/send-test',
    {
      preValidation: requireRoleGuard('admin'),
      schema: { body: adminSendTestBodySchema },
    },
    async (req, reply) => {
      const body = req.body as Partial<{
        tenantId: string;
        userId: string;
        channel: string;
        templateName: string;
        payload: Record<string, unknown>;
        priority: 'critical' | 'high' | 'normal' | 'low';
        maxAttempts: number;
      }>;

      const tenantId = body?.tenantId;
      const userId = body?.userId;
      const channel = body?.channel ?? 'email';
      const templateName = body?.templateName ?? 'test';
      const payload = body?.payload ?? {};
      const priority = body?.priority ?? 'normal';
      const maxAttempts = body?.maxAttempts ?? 3;

      const jobId = await deps.enqueue({
        tenantId: tenantId!,
        userId: userId!,
        channel,
        templateName,
        payload,
        priority,
        maxAttempts,
      });
      return reply.code(202).send({ ok: true, jobId });
    },
  );
}

export const registerAdminRoutes = adminSendTestRoute;
