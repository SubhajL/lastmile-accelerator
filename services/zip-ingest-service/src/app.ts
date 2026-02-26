import Fastify from 'fastify';
import type { FastifyInstance } from 'fastify';

import { registerHealthRoutes } from './routes/health.js';
import { registerZipIngestRoutes } from './routes/zip-ingest.js';

export async function createApp(opts?: {
  snapshotOrchestratorUrl?: string;
  logger?: boolean;
}): Promise<FastifyInstance> {
  const app = Fastify({ logger: opts?.logger ?? true });

  registerHealthRoutes(app);
  registerZipIngestRoutes(app, {
    snapshotOrchestratorUrl: opts?.snapshotOrchestratorUrl,
  });

  return app;
}
