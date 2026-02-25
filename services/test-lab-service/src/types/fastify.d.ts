import 'fastify';
import type { UserContext } from './auth.js';
import type { RedisWrapper } from '../clients/redis.js';
import type { CacheHelper } from '../lib/cache.js';

declare module 'fastify' {
  interface FastifyRequest {
    user?: UserContext;
  }
  interface FastifyInstance {
    redis: RedisWrapper;
    cache: CacheHelper;
    repos: {
      scaffolds: import('../repo/scaffolds.repo.js').ScaffoldsRepo;
      testRuns?: import('../repo/test-runs.pg.repo.js').PgTestRunsRepo;
      browserTestRuns?: import('../repo/browser-test-runs.pg.repo.js').PgBrowserTestRunsRepo;
      previewEnvs?: import('../repo/preview-envs.pg.repo.js').PgPreviewEnvsRepo;
    };
  }
}
