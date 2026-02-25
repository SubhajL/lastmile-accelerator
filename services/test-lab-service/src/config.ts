import { z } from 'zod';

const envSchema = z.object({
  SERVICE_PORT: z.string().pipe(z.coerce.number().int().positive()).default('7202'),
  SERVICE_NAME: z.string(),
  SERVICE_VERSION: z.string().default('1.0.0'),
  DATABASE_URL: z.string().url(),
  REDIS_URL: z.string().url(),
  NATS_URL: z.string().url(),
  OTEL_EXPORTER_OTLP_ENDPOINT: z.string().url(),
  JWT_JWKS_URL: z.string(),
  JWT_ISSUER: z.string().url(),
  JWT_AUDIENCE: z.string().min(1),
  JWT_ALG: z.enum(['RS256', 'RS384', 'RS512', 'ES256']).default('RS256'),
  JWKS_CACHE_TTL_MS: z.string().pipe(z.coerce.number().int().positive()).default('600000'),
  JWT_CLOCK_SKEW_SEC: z.string().pipe(z.coerce.number().int().min(0)).default('60'),
  S3_BUCKET_PREVIEWS: z.string(),
  BROWSER_GRID_URL: z.string().url(),
  VAULT_ADDR: z.string().url(),
  VAULT_ROLE_ID: z.string().optional(),
  VAULT_SECRET_ID: z.string().optional(),
  ENV: z.enum(['dev', 'staging', 'prod']).default('dev'),
  TEST_TIMEOUT_MS: z.string().pipe(z.coerce.number().int().positive()).default('60000'),
  MAX_PARALLEL_TESTS: z.string().pipe(z.coerce.number().int().positive()).default('3'),
  REPO_BACKEND: z.enum(['memory', 'pg']).default('memory'),
  // New for runner/grid and artifacts
  S3_BUCKET_ARTIFACTS: z.string().default('test-artifacts'),
  S3_REGION: z.string().default('us-east-1'),
  K8S_NAMESPACE: z.string().default('default'),
  K8S_SERVICE_ACCOUNT: z.string().default('default'),
  K8S_JOB_IMAGE_DEFAULT: z.string().default('busybox'),
  K8S_TTL_SECONDS_AFTER_FINISHED: z.string().pipe(z.coerce.number().int().min(0)).default('600'),
  RUNNER_POLL_INTERVAL_MS: z.string().pipe(z.coerce.number().int().positive()).default('3000'),
  ENABLE_EVENT_SUBSCRIBERS: z.string().default('false'),
  // Redis and rate limiting
  RATE_LIMIT_REQUESTS_PER_MINUTE: z
    .string()
    .pipe(z.coerce.number().int().positive())
    .default('100'),
  RATE_LIMIT_WINDOW_MS: z.string().pipe(z.coerce.number().int().positive()).default('60000'),
  CACHE_DEFAULT_TTL_SEC: z.string().pipe(z.coerce.number().int().positive()).default('300'),
  REDIS_KEY_PREFIX: z.string().default('test-lab-service:'),
});

export interface ServerConfig {
  port: number;
  serviceName: string;
  serviceVersion: string;
  databaseUrl: string;
  redisUrl: string;
  natsUrl: string;
  repoBackend: 'memory' | 'pg';
  otelEndpoint: string;
  jwtJwksUrl: string;
  jwtIssuer: string;
  jwtAudience: string;
  jwtAlg: 'RS256' | 'RS384' | 'RS512' | 'ES256';
  jwksCacheTtlMs: number;
  jwtClockSkewSec: number;
  s3BucketPreviews: string;
  browserGridUrl: string;
  vaultAddr: string;
  vaultRoleId?: string;
  vaultSecretId?: string;
  env: 'dev' | 'staging' | 'prod';
  testTimeoutMs: number;
  maxParallelTests: number;
  // new
  s3BucketArtifacts: string;
  s3Region: string;
  k8sNamespace: string;
  k8sServiceAccount: string;
  k8sJobImageDefault: string;
  k8sTtlSecondsAfterFinished: number;
  runnerPollIntervalMs: number;
  enableEventSubscribers: boolean;
  // Redis and rate limiting
  rateLimitRequestsPerMinute: number;
  rateLimitWindowMs: number;
  cacheDefaultTtlSec: number;
  redisKeyPrefix: string;
}

/**
 * Load and validate configuration from environment variables.
 * Uses Zod for type-safe validation with helpful error messages.
 *
 * @throws Error if any required environment variable is missing or invalid
 * @returns Validated configuration object
 */
export function loadConfig(): ServerConfig {
  try {
    const raw = envSchema.parse({
      SERVICE_PORT: process.env.SERVICE_PORT || '7202',
      SERVICE_NAME: process.env.SERVICE_NAME,
      SERVICE_VERSION: process.env.SERVICE_VERSION || '1.0.0',
      DATABASE_URL: process.env.DATABASE_URL,
      REDIS_URL: process.env.REDIS_URL,
      NATS_URL: process.env.NATS_URL,
      OTEL_EXPORTER_OTLP_ENDPOINT: process.env.OTEL_EXPORTER_OTLP_ENDPOINT,
      JWT_JWKS_URL: process.env.JWT_JWKS_URL,
      JWT_ISSUER: process.env.JWT_ISSUER,
      JWT_AUDIENCE: process.env.JWT_AUDIENCE,
      JWT_ALG: process.env.JWT_ALG,
      JWKS_CACHE_TTL_MS: process.env.JWKS_CACHE_TTL_MS || '600000',
      JWT_CLOCK_SKEW_SEC: process.env.JWT_CLOCK_SKEW_SEC || '60',
      S3_BUCKET_PREVIEWS: process.env.S3_BUCKET_PREVIEWS,
      BROWSER_GRID_URL: process.env.BROWSER_GRID_URL,
      VAULT_ADDR: process.env.VAULT_ADDR,
      VAULT_ROLE_ID: process.env.VAULT_ROLE_ID,
      VAULT_SECRET_ID: process.env.VAULT_SECRET_ID,
      ENV: process.env.ENV || 'dev',
      TEST_TIMEOUT_MS: process.env.TEST_TIMEOUT_MS || '60000',
      MAX_PARALLEL_TESTS: process.env.MAX_PARALLEL_TESTS || '3',
      REPO_BACKEND: (process.env.REPO_BACKEND as any) || 'memory',
      S3_BUCKET_ARTIFACTS: process.env.S3_BUCKET_ARTIFACTS || 'test-artifacts',
      S3_REGION: process.env.S3_REGION || 'us-east-1',
      K8S_NAMESPACE: process.env.K8S_NAMESPACE || 'default',
      K8S_SERVICE_ACCOUNT: process.env.K8S_SERVICE_ACCOUNT || 'default',
      K8S_JOB_IMAGE_DEFAULT: process.env.K8S_JOB_IMAGE_DEFAULT || 'busybox',
      K8S_TTL_SECONDS_AFTER_FINISHED: process.env.K8S_TTL_SECONDS_AFTER_FINISHED || '600',
      RUNNER_POLL_INTERVAL_MS: process.env.RUNNER_POLL_INTERVAL_MS || '3000',
      ENABLE_EVENT_SUBSCRIBERS: process.env.ENABLE_EVENT_SUBSCRIBERS || 'false',
      RATE_LIMIT_REQUESTS_PER_MINUTE: process.env.RATE_LIMIT_REQUESTS_PER_MINUTE || '100',
      RATE_LIMIT_WINDOW_MS: process.env.RATE_LIMIT_WINDOW_MS || '60000',
      CACHE_DEFAULT_TTL_SEC: process.env.CACHE_DEFAULT_TTL_SEC || '300',
      REDIS_KEY_PREFIX: process.env.REDIS_KEY_PREFIX || 'test-lab-service:',
    });

    return {
      port: raw.SERVICE_PORT,
      serviceName: raw.SERVICE_NAME,
      serviceVersion: raw.SERVICE_VERSION,
      databaseUrl: raw.DATABASE_URL,
      redisUrl: raw.REDIS_URL,
      natsUrl: raw.NATS_URL,
      otelEndpoint: raw.OTEL_EXPORTER_OTLP_ENDPOINT,
      jwtJwksUrl: raw.JWT_JWKS_URL,
      jwtIssuer: raw.JWT_ISSUER,
      jwtAudience: raw.JWT_AUDIENCE,
      jwtAlg: raw.JWT_ALG,
      jwksCacheTtlMs: raw.JWKS_CACHE_TTL_MS,
      jwtClockSkewSec: raw.JWT_CLOCK_SKEW_SEC,
      s3BucketPreviews: raw.S3_BUCKET_PREVIEWS,
      browserGridUrl: raw.BROWSER_GRID_URL,
      vaultAddr: raw.VAULT_ADDR,
      vaultRoleId: raw.VAULT_ROLE_ID,
      vaultSecretId: raw.VAULT_SECRET_ID,
      env: raw.ENV,
      testTimeoutMs: raw.TEST_TIMEOUT_MS,
      maxParallelTests: raw.MAX_PARALLEL_TESTS,
      repoBackend: raw.REPO_BACKEND,
      s3BucketArtifacts: raw.S3_BUCKET_ARTIFACTS,
      s3Region: raw.S3_REGION,
      k8sNamespace: raw.K8S_NAMESPACE,
      k8sServiceAccount: raw.K8S_SERVICE_ACCOUNT,
      k8sJobImageDefault: raw.K8S_JOB_IMAGE_DEFAULT,
      k8sTtlSecondsAfterFinished: raw.K8S_TTL_SECONDS_AFTER_FINISHED,
      runnerPollIntervalMs: raw.RUNNER_POLL_INTERVAL_MS,
      enableEventSubscribers: String(raw.ENABLE_EVENT_SUBSCRIBERS).toLowerCase() === 'true',
      rateLimitRequestsPerMinute: raw.RATE_LIMIT_REQUESTS_PER_MINUTE,
      rateLimitWindowMs: raw.RATE_LIMIT_WINDOW_MS,
      cacheDefaultTtlSec: raw.CACHE_DEFAULT_TTL_SEC,
      redisKeyPrefix: raw.REDIS_KEY_PREFIX,
    };
  } catch (error) {
    if (error instanceof z.ZodError) {
      const missingVars = error.issues.map((issue) => issue.path.join('.')).join(', ');
      throw new Error(`Configuration error: missing or invalid variables: ${missingVars}`);
    }
    throw error;
  }
}

/**
 * Validate all required environment variables at startup.
 * Calls loadConfig() and fails fast with clear error message.
 *
 * @throws Error if validation fails
 */
export function validateEnv(): void {
  loadConfig();
}

/**
 * Get configuration singleton.
 * For lazy loading after environment is set up.
 */
let config: ServerConfig | null = null;

export function getConfig(): ServerConfig {
  if (!config) {
    config = loadConfig();
  }
  return config;
}

// Test helper to allow reloading config between tests
export function __resetConfigForTests() {
  config = null;
}
