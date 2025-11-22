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

export function adminSendTestRoute(app: FastifyInstance, deps: AdminDeps) {
  app.post('/send-test', {
    preHandler: requireRoleGuard('admin')
  }, async (req, reply) => {
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

    if (!tenantId || !userId) {
      return reply.code(400).send({ error: 'tenantId and userId are required' });
    }

    const jobId = await deps.enqueue({ tenantId, userId, channel, templateName, payload, priority, maxAttempts });
    return reply.code(202).send({ ok: true, jobId });
  });
}

export const registerAdminRoutes = adminSendTestRoute;
