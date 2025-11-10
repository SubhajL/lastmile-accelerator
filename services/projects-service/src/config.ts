import { z } from 'zod';

const envSchema = z.object({
  SERVICE_PORT: z.string().pipe(z.coerce.number()).default('7002'),
  SERVICE_NAME: z.string(),
  DATABASE_URL: z.string(),
  NATS_URL: z.string(),
  OTEL_EXPORTER_OTLP_ENDPOINT: z.string(),
  JWT_JWKS_URL: z.string(),
  ENV: z.enum(['dev', 'staging', 'prod']).default('dev'),
});

export interface ServerConfig {
  port: number;
  serviceName: string;
  databaseUrl: string;
  natsUrl: string;
  otelEndpoint: string;
  jwtJwksUrl: string;
  env: string;
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
      SERVICE_PORT: process.env.SERVICE_PORT || '7002',
      SERVICE_NAME: process.env.SERVICE_NAME,
      DATABASE_URL: process.env.DATABASE_URL,
      NATS_URL: process.env.NATS_URL,
      OTEL_EXPORTER_OTLP_ENDPOINT: process.env.OTEL_EXPORTER_OTLP_ENDPOINT,
      JWT_JWKS_URL: process.env.JWT_JWKS_URL,
      ENV: process.env.ENV || 'dev',
    });

    return {
      port: raw.SERVICE_PORT as number,
      serviceName: raw.SERVICE_NAME,
      databaseUrl: raw.DATABASE_URL,
      natsUrl: raw.NATS_URL,
      otelEndpoint: raw.OTEL_EXPORTER_OTLP_ENDPOINT,
      jwtJwksUrl: raw.JWT_JWKS_URL,
      env: raw.ENV,
    };
  } catch (error) {
    if (error instanceof z.ZodError) {
      const missingVars = error.issues
        .map((issue) => issue.path.join('.'))
        .join(', ');
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
