import Fastify from 'fastify';
import type { FastifyInstance } from 'fastify';
import cors from '@fastify/cors';
import helmet from '@fastify/helmet';
import type { Config } from './types.js';
import type { Pool } from 'pg';
import { healthCheck } from './db/client.js';

export async function createApp(config: Config, dbPool: Pool): Promise<FastifyInstance> {
  const app = Fastify({
    logger: {
      level: config.env === 'prod' ? 'info' : 'debug'
    }
  });

  // Register security plugins
  await app.register(cors, {
    origin: true,
    credentials: true
  });

  await app.register(helmet, {
    contentSecurityPolicy: false // Allow customization later
  });

  // Register health check endpoint
  app.get('/healthz', async (_request, reply) => {
    const isDbHealthy = await healthCheck(dbPool);
    
    if (!isDbHealthy) {
      return reply.code(503).send({
        status: 'unhealthy',
        reason: 'database_unavailable',
        timestamp: new Date().toISOString()
      });
    }

    return reply.code(200).send({
      status: 'ok',
      timestamp: new Date().toISOString()
    });
  });

  // Register metrics endpoint (placeholder for Prometheus metrics)
  app.get('/metrics', async (_request, reply) => {
    // In a real implementation, this would export Prometheus metrics
    // For now, return basic metrics format
    const metrics = [
      '# HELP notification_service_up Service uptime',
      '# TYPE notification_service_up gauge',
      'notification_service_up 1',
      ''
    ].join('\n');

    return reply
      .code(200)
      .header('Content-Type', 'text/plain; version=0.0.4')
      .send(metrics);
  });

  return app;
}
