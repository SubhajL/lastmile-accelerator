import Fastify from 'fastify';
import type { FastifyInstance } from 'fastify';

import { registerHealthRoutes } from './routes/health.js';
import { registerFixListRoutes } from './routes/fix-list.js';
import { registerFixManifestRoutes } from './routes/fix-manifest.js';

type CreateAppOptions = {
  logger?: boolean;
  snapshotOrchestratorUrl?: string;
};

export async function createApp(opts: CreateAppOptions = {}): Promise<FastifyInstance> {
  const app = Fastify({ logger: opts.logger ?? true });

  registerHealthRoutes(app);
  registerFixListRoutes(app, { snapshotOrchestratorUrl: opts.snapshotOrchestratorUrl });
  registerFixManifestRoutes(app);

  return app;
}
