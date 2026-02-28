import Fastify from 'fastify';
import type { FastifyInstance } from 'fastify';

import type { ZipUploadSessionStore } from './lib/zip-upload-session-store.js';
import { registerZipIngestRoutes } from './routes/zip-ingest.js';
import { registerZipUploadRoutes } from './routes/zip-uploads.js';

export async function createApp(opts?: {
  snapshotOrchestratorUrl?: string;
  uploadBaseUrl?: string;
  uploadSigningSecret?: string;
  uploadTokenTtlSeconds?: number;
  uploadMaxBytes?: number;
  zipUploadSessionStore?: ZipUploadSessionStore;
  logger?: boolean;
}): Promise<FastifyInstance> {
  const app = Fastify({ logger: opts?.logger ?? true });

  app.addContentTypeParser('application/zip', { parseAs: 'buffer' }, (_req, body, done) =>
    done(null, body),
  );

  app.get('/healthz', async () => 'ok');
  registerZipIngestRoutes(app, {
    snapshotOrchestratorUrl: opts?.snapshotOrchestratorUrl,
  });
  registerZipUploadRoutes(app, {
    snapshotOrchestratorUrl: opts?.snapshotOrchestratorUrl,
    uploadBaseUrl: opts?.uploadBaseUrl,
    uploadSigningSecret: opts?.uploadSigningSecret,
    uploadTokenTtlSeconds: opts?.uploadTokenTtlSeconds,
    uploadMaxBytes: opts?.uploadMaxBytes,
    zipUploadSessionStore: opts?.zipUploadSessionStore,
  });

  return app;
}
