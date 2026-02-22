import type { FastifyInstance } from 'fastify';
import type { Registry } from 'prom-client';
import { registerMetricsEndpoint } from './metrics/prom.js';
import { adminSendTestRoute, type AdminDeps } from './routes/admin.send-test.js';
import { adminRetryRoute } from './routes/admin.retry.js';
import { healthzRoute } from './routes/healthz.js';

export interface AppDeps {
  queue: {
    retryNotification: (notificationId: string) => Promise<{ success: boolean; error?: string }>;
  };
  enqueue: AdminDeps['enqueue'];
  dbHealthCheck?: () => Promise<boolean>;
  config: { service: { name: string } };
}

export function registerHttpRoutes(app: FastifyInstance, deps: AppDeps, registry: Registry): void {
  // Register metrics endpoint and decorate app
  registerMetricsEndpoint(app, registry);

  // Register healthz endpoint
  app.get('/healthz', healthzRoute(deps));

  // Register admin routes with prefix
  app.register(
    async function (fastify) {
      // Register existing admin send-test route
      adminSendTestRoute(fastify, { enqueue: deps.enqueue });

      // Register new admin retry route
      adminRetryRoute(fastify, deps);
    },
    { prefix: '/admin' },
  );
}
