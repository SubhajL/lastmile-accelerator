import Fastify from 'fastify';
import type { FastifyInstance } from 'fastify';

import { registerHealthRoutes } from './routes/health.js';
import { registerFixListRoutes } from './routes/fix-list.js';

export async function createApp(): Promise<FastifyInstance> {
  const app = Fastify({ logger: false });

  registerHealthRoutes(app);
  registerFixListRoutes(app);

  return app;
}
