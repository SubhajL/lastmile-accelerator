import Fastify from 'fastify';
import type { FastifyInstance } from 'fastify';
import cors from '@fastify/cors';
import helmet from '@fastify/helmet';
import type { Config } from './types.js';
import type { Pool } from 'pg';
import { healthCheck } from './db/client.js';
import { createPrometheusRegistry, createNotificationMetrics } from './metrics/prom.js';
import { registerHttpRoutes, type AppDeps } from './http-routes.js';
import { buildJwtAuth } from './auth/jwt.js';

export async function createApp(config: Config, dbPool: Pool): Promise<FastifyInstance> {
  const app = Fastify({
    logger: {
      level: config.env === 'prod' ? 'info' : 'debug',
    },
  });

  // Register security plugins
  await app.register(cors, {
    origin: true,
    credentials: true,
  });

  await app.register(helmet, {
    contentSecurityPolicy: false, // Allow customization later
  });

  // Build JWT auth
  buildJwtAuth(app, config.auth);

  // Create Prometheus registry and metrics
  const registry = createPrometheusRegistry();
  createNotificationMetrics(registry); // Metrics will be used later

  // Create app dependencies
  const appDeps: AppDeps = {
    queue: {
      retryNotification: async () => {
        // TODO: Implement actual retry logic
        return { success: true };
      },
    },
    enqueue: async () => {
      // TODO: Implement actual enqueue logic
      return 'job-id-123';
    },
    dbHealthCheck: () => healthCheck(dbPool),
    config: config,
  };

  // Register all HTTP routes
  registerHttpRoutes(app, appDeps, registry);

  return app;
}
