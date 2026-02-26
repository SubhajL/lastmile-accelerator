import Fastify from 'fastify';
import type { FastifyInstance } from 'fastify';

import { registerHealthRoutes } from './routes/health.js';
import { registerFixListRoutes } from './routes/fix-list.js';

type CreateAppOptions = {
  logger?: boolean;
};

export async function createApp(opts: CreateAppOptions = {}): Promise<FastifyInstance> {
  const app = Fastify({ logger: opts.logger ?? true });

  registerHealthRoutes(app);
  registerFixListRoutes(app);

  return app;
}
