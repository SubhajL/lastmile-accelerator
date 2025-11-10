import 'fastify';
import type { UserContext } from './auth.js';

declare module 'fastify' {
  interface FastifyRequest {
    user?: UserContext;
  }
  interface FastifyInstance {
    repos: {
      scaffolds: import('../repo/scaffolds.repo.js').ScaffoldsRepo;
      testRuns?: import('../repo/test-runs.pg.repo.js').PgTestRunsRepo;
      browserTestRuns?: import('../repo/browser-test-runs.pg.repo.js').PgBrowserTestRunsRepo;
      previewEnvs?: import('../repo/preview-envs.pg.repo.js').PgPreviewEnvsRepo;
    };
  }
}
