import Fastify from 'fastify';
import { getConfig } from './config.js';
import { initTelemetry } from './lib/telemetry.js';
import type { RedisWrapper } from './clients/redis.js';
import { registerErrorHandler } from './middleware/error-handler.js';
import { registerAuthPlugin } from './middleware/auth.js';
import { createMetricsRoute } from './lib/metrics.js';

export async function createApp() {
  const cfg = getConfig();

  // Initialize telemetry asap
  initTelemetry({
    serviceName: cfg.serviceName,
    serviceVersion: cfg.serviceVersion,
    otlpEndpoint: cfg.otelEndpoint,
    environment: cfg.env as 'dev' | 'staging' | 'prod',
  });

  // Use Fastify default logger for type compatibility; custom logger can be wired later if needed.
  const app = Fastify({ logger: true });

  // Error handler first
  registerErrorHandler(app);

  // Auth plugin (JWKS-based hooks handle verification)
  await registerAuthPlugin(app);

  let redis: RedisWrapper | null = null;

  try {
    // Redis client and cache setup
    const { createRedisClient } = await import('./clients/redis.js');
    redis = await createRedisClient(cfg.redisUrl, { logger: app.log });
    const { createCacheHelper } = await import('./lib/cache.js');
    const cache = createCacheHelper(redis, cfg.redisKeyPrefix);

    // Decorate app with redis and cache for use in routes/services
    app.decorate('redis', redis);
    app.decorate('cache', cache);

    // Register rate limiting middleware (after auth, before routes)
    const { createRateLimitMiddleware } = await import('./middleware/rate-limit.js');
    app.addHook(
      'preHandler',
      createRateLimitMiddleware(redis, {
        requestsPerMinute: cfg.rateLimitRequestsPerMinute,
        windowMs: cfg.rateLimitWindowMs,
        exemptRoutes: ['/healthz', '/metrics'],
        keyPrefix: `${cfg.redisKeyPrefix}rate-limit:`,
      }),
    );

    // Repos: choose backend from config
    if (cfg.repoBackend === 'pg') {
      const { connectDb } = await import('./repo/pg.js');
      const pool = await connectDb(cfg.databaseUrl);
      const { runMigrations } = await import('./repo/db.js');
      await runMigrations(pool as any);
      const { PgScaffoldsRepo } = await import('./repo/scaffolds.pg.repo.js');
      const { PgTestRunsRepo } = await import('./repo/test-runs.pg.repo.js');
      const { PgBrowserTestRunsRepo } = await import('./repo/browser-test-runs.pg.repo.js');
      const { PgPreviewEnvsRepo } = await import('./repo/preview-envs.pg.repo.js');
      app.decorate('repos', {
        scaffolds: new PgScaffoldsRepo(pool as any),
        testRuns: new PgTestRunsRepo(pool as any),
        browserTestRuns: new PgBrowserTestRunsRepo(pool as any),
        previewEnvs: new PgPreviewEnvsRepo(pool as any),
      });
      app.addHook('onClose', async () => {
        const { closeDb } = await import('./repo/pg.js');
        await closeDb(pool as any);
      });
    } else {
      const { InMemoryScaffoldsRepo } = await import('./repo/scaffolds.repo.js');
      app.decorate('repos', { scaffolds: new InMemoryScaffoldsRepo() });
    }

    // Health & metrics
    app.get('/healthz', async () => 'ok');
    createMetricsRoute(app);

    // Routes
    const { registerScaffoldRoutes } = await import('./routes/scaffolds.routes.js');
    registerScaffoldRoutes(app);
    if (cfg.repoBackend === 'pg') {
      const { registerTestRunRoutes } = await import('./routes/test-runs.routes.js');
      registerTestRunRoutes(app);
      const { registerPreviewRoutes } = await import('./routes/previews.routes.js');
      registerPreviewRoutes(app);

      // Event subscribers wiring (feature-flagged)
      if (cfg.enableEventSubscribers) {
        const { createNatsConnection } = await import('./clients/nats.js');
        const nats = await createNatsConnection(cfg.natsUrl);
        const { createS3Client } = await import('./clients/s3.js');
        const s3 = createS3Client(cfg.s3Region);
        const { createK8sClient } = await import('./clients/k8s.js');
        const k8s = await createK8sClient();
        const { RunnerOrchestrator } = await import('./services/runner.js');
        const runner = new RunnerOrchestrator({
          config: {
            namespace: cfg.k8sNamespace,
            serviceAccount: cfg.k8sServiceAccount,
            image: cfg.k8sJobImageDefault,
            ttlSecondsAfterFinished: cfg.k8sTtlSecondsAfterFinished,
            bucketArtifacts: cfg.s3BucketArtifacts,
          },
          batch: k8s.batch as any,
          s3: s3 as any,
          nats,
          testRunsRepo: app.repos.testRuns!,
        });
        const { BrowserGridRunner } = await import('./services/browser-grid.js');
        const gridRunner = new BrowserGridRunner({
          createDriver: async (browser: string, gridUrl: string) => {
            const mod: any = await import('selenium-webdriver');
            const driver = await new mod.Builder().forBrowser(browser).usingServer(gridUrl).build();
            return driver as any;
          },
          gridUrl: cfg.browserGridUrl,
          s3: s3 as any,
          bucketScreenshots: cfg.s3BucketArtifacts,
          nats,
          browserRunsRepo: app.repos.browserTestRuns!,
          testFlow: async (driver: any) => {
            await driver.get('about:blank');
          },
          retry: { retries: 1, backoffMs: 500 },
        });
        const { wireEventSubscribers } = await import('./events/subscribers.js');
        await wireEventSubscribers({ nats, runner, grid: gridRunner });

        app.addHook('onClose', async () => {
          await nats.close();
        });
      }
    }

    // Redis cleanup on app shutdown (catch to guard against double-close)
    app.addHook('onClose', async () => {
      await redis?.close().catch(() => {});
    });

    return app;
  } catch (err) {
    await redis?.close().catch(() => {});
    throw err;
  }
}
