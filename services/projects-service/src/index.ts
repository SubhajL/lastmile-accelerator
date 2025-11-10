import Fastify from 'fastify';
import { requestIdMiddleware } from './middleware/requestId';
import { metricsMiddleware } from './middleware/metrics';
import { errorHandler } from './middleware/errorHandler';
import { registerHealthRoutes } from './routes/health';
import { registerHttpTracing } from './observability/httpTracing';
import { registerRequestLogging } from './middleware/logging';
import { registerReadyRoute } from './routes/ready';
import { registerProjectRoutes } from './routes/projects';
import { registerTenantRoutes } from './routes/tenants';
import { registerMemberRoutes } from './routes/members';
import { registerEnvironmentRoutes } from './routes/environments';
import { createLogger } from './logger';

export interface ServerOptions {
  serviceName?: string;
  autoInit?: boolean; // reserved for wiring external deps later
  auth?: import('./middleware/auth').JwtConfig;
}

export async function createServer(opts: ServerOptions = {}) {
  const serviceName = opts.serviceName ?? process.env.SERVICE_NAME ?? 'projects-service';

  // initialize logger early
  try { createLogger(serviceName); } catch { /* no-op if already created */ }

  const app = Fastify({ logger: false });
  // core middleware
  if (opts.autoInit !== false) {
    app.addHook('onRequest', requestIdMiddleware());
    registerRequestLogging(app);
    app.addHook('onRequest', metricsMiddleware());
    app.setErrorHandler(errorHandler as any);
  }

  // http tracing enrichment (safe no-op without active span)
  registerHttpTracing(app, { serviceName });

  // auth for /v1/*
  registerAuth(app, opts.auth ?? {});

  // routes
  registerHealthRoutes(app, { serviceName });
  registerProjectRoutes(app);
  registerTenantRoutes(app);
  registerMemberRoutes(app);
  registerEnvironmentRoutes(app);
  registerReadyRoute(app);

  return app;
}

import { authMiddleware, type JwtConfig } from './middleware/auth';
export function registerAuth(app: any, cfg: JwtConfig) {
  const mw = authMiddleware(cfg);
  app.addHook('preHandler', async (req: any, reply: any) => {
    const url = req.raw?.url || req.url || '';
    if (url.startsWith('/v1/')) {
      await mw(req, reply);
    }
  });
}

function envTrue(name: string, defaultVal: boolean): boolean {
  const v = process.env[name];
  if (v === undefined) return defaultVal;
  return v.toLowerCase() === 'true';
}

export async function start() {
  const port = Number(process.env.SERVICE_PORT || 7002);
  const host = process.env.SERVICE_HOST || '0.0.0.0';
  const serviceName = process.env.SERVICE_NAME || 'projects-service';

  // Determine defaults: enable in non-test env by default
  const defaultEnable = process.env.NODE_ENV === 'test' ? false : true;

  // Conditional initializations
  const shouldInitDb = envTrue('INIT_DB', defaultEnable);
  const shouldRunMigrations = envTrue('RUN_MIGRATIONS', defaultEnable) && shouldInitDb;
  const shouldInitNats = envTrue('INIT_NATS', defaultEnable);
  const exporterUrl = process.env.OTEL_EXPORTER_OTLP_ENDPOINT;
  const metricsExporterUrl = process.env.OTEL_EXPORTER_OTLP_METRICS_ENDPOINT;
  const shouldInitOtel = envTrue('INIT_OTEL', defaultEnable) && !!exporterUrl;

  if (shouldInitDb) {
    const { initDb } = await import('./db');
    const conn = process.env.DATABASE_URL;
    if (!conn) throw new Error('DATABASE_URL is required when INIT_DB=true');
    await initDb({ connectionString: conn });

    if (shouldRunMigrations) {
      const { runMigrations } = await import('./db/migrations/migrate');
      await runMigrations();
    }
  }

  if (shouldInitNats) {
    const { initNats } = await import('./nats');
    const natsUrl = process.env.NATS_URL;
    if (!natsUrl) throw new Error('NATS_URL is required when INIT_NATS=true');
    await initNats(natsUrl, {} as any);
  }

  if (shouldInitOtel) {
    const { initOtel } = await import('./otel');
    await initOtel(serviceName, { exporterUrl, metricsUrl: metricsExporterUrl });
  }

  const app = await createServer({
    serviceName,
    auth: {
      jwksUrl: process.env.AUTH_JWKS_URL,
      issuer: process.env.AUTH_ISSUER,
      audience: process.env.AUTH_AUDIENCE,
      clockToleranceSec: process.env.AUTH_CLOCK_TOLERANCE_SEC ? Number(process.env.AUTH_CLOCK_TOLERANCE_SEC) : undefined,
    },
  });
  await app.listen({ port, host });
  // eslint-disable-next-line no-console
  console.log(`service listening on :${port}`);
  return app;
}

// Auto-start unless in test
if (process.env.NODE_ENV !== 'test') {
  start().catch(err => {
    // eslint-disable-next-line no-console
    console.error(err);
    process.exit(1);
  });
}
