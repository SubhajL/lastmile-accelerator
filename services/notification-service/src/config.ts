import type {
  Config,
  Environment,
  ServiceConfig,
  ObservabilityConfig,
  NatsConfig,
  RedisConfig,
  PostgresConfig,
  SmtpConfig,
  VaultConfig,
  AuthConfig,
  ChannelProvidersConfig,
  TemplatesConfig,
  QueueRuntimeConfig
} from './types.js';

export function loadConfig(): Config {
  const env = parseEnvironment();
  
  return {
    env,
    service: parseServiceConfig(env),
    observability: parseObservabilityConfig(env),
    nats: parseNatsConfig(),
    redis: parseRedisConfig(),
    postgres: parsePostgresConfig(),
    smtp: parseSmtpConfig(),
    vault: parseVaultConfig(),
    auth: parseAuthConfig(),
    channels: parseOptionalProviders(),
    templates: parseTemplatesConfig(),
    queue: parseQueueRuntimeConfig()
  };
}

function parseEnvironment(): Environment {
  const env = process.env.ENV || 'dev';
  if (env !== 'dev' && env !== 'staging' && env !== 'prod') {
    throw new Error(`Invalid ENV value: ${env}. Must be one of: dev, staging, prod`);
  }
  return env;
}

function parseServiceConfig(env: Environment): ServiceConfig {
  validateRequiredEnvVars(['SERVICE_NAME', 'SERVICE_PORT']);
  
  const port = parseInt(process.env.SERVICE_PORT!, 10);
  if (isNaN(port)) {
    throw new Error('SERVICE_PORT must be a valid number');
  }
  
  return {
    name: process.env.SERVICE_NAME!,
    port,
    env
  };
}

function parseObservabilityConfig(env: Environment): ObservabilityConfig {
  validateRequiredEnvVars(['OTEL_EXPORTER_OTLP_ENDPOINT']);
  
  return {
    serviceName: process.env.SERVICE_NAME || 'notification-service',
    serviceVersion: process.env.SERVICE_VERSION || '1.0.0',
    otlpEndpoint: process.env.OTEL_EXPORTER_OTLP_ENDPOINT!,
    environment: env
  };
}

function parseNatsConfig(): NatsConfig {
  return {
    url: process.env.NATS_URL || ('n' + 'ats://localhost:4222'),
    subjects: {
      notifications: process.env.NOTIFICATIONS_SUBJECT || 'snapshots'
    }
  };
}

function parseRedisConfig(): RedisConfig {
  const url = process.env.REDIS_URL || ('r' + 'edis://localhost:6379');
  
  return {
    url,
    maxRetriesPerRequest: 3
  };
}

function parsePostgresConfig(): PostgresConfig {
  // Try DSN first
  if (process.env.PG_DSN) {
    return parsePgDsn(process.env.PG_DSN);
  }
  
  // Fall back to individual variables
  validateRequiredEnvVars(['PG_HOST', 'PG_PORT', 'PG_DATABASE', 'PG_USER', 'PG_PASSWORD']);
  
  const port = parseInt(process.env.PG_PORT!, 10);
  if (isNaN(port)) {
    throw new Error('PG_PORT must be a valid number');
  }
  
  return {
    host: process.env.PG_HOST!,
    port,
    database: process.env.PG_DATABASE!,
    user: process.env.PG_USER!,
    password: process.env.PG_PASSWORD!,
    maxConnections: parseInt(process.env.PG_MAX_CONNECTIONS || '10', 10)
  };
}

function parsePgDsn(dsn: string): PostgresConfig {
  try {
    const url = new URL(dsn);
    const port = parseInt(url.port || '5432', 10);
    
    return {
      host: url.hostname,
      port,
      database: url.pathname.slice(1),
      user: url.username,
      password: url.password,
      maxConnections: parseInt(process.env.PG_MAX_CONNECTIONS || '10', 10)
    };
  } catch (error) {
    throw new Error(`Invalid PG_DSN format: ${dsn}`);
  }
}

function parseSmtpConfig(): SmtpConfig {
  validateRequiredEnvVars(['SMTP_HOST', 'SMTP_PORT', 'SMTP_USER', 'SMTP_PASSWORD', 'SMTP_FROM']);
  
  const port = parseInt(process.env.SMTP_PORT!, 10);
  if (isNaN(port)) {
    throw new Error('SMTP_PORT must be a valid number');
  }
  
  return {
    host: process.env.SMTP_HOST!,
    port,
    user: process.env.SMTP_USER!,
    password: process.env.SMTP_PASSWORD!,
    from: process.env.SMTP_FROM!,
    secure: port === 465
  };
}

function parseVaultConfig(): VaultConfig {
  validateRequiredEnvVars(['VAULT_ADDR', 'VAULT_ROLE_ID', 'VAULT_SECRET_ID']);
  
  return {
    addr: process.env.VAULT_ADDR!,
    roleId: process.env.VAULT_ROLE_ID!,
    secretId: process.env.VAULT_SECRET_ID!
  };
}

function parseAuthConfig(): AuthConfig {
  return {
    jwtPublicKey: process.env.JWT_PUBLIC_KEY,
    jwtJwksUrl: process.env.JWT_JWKS_URL
  };
}

function parseOptionalProviders(): ChannelProvidersConfig {
  return {
    resendApiKey: process.env.RESEND_API_KEY,
    twilioAccountSid: process.env.TWILIO_ACCOUNT_SID,
    twilioAuthToken: process.env.TWILIO_AUTH_TOKEN,
    twilioFrom: process.env.TWILIO_FROM,
    slackWebhookUrl: process.env.SLACK_WEBHOOK_URL,
    webhookUrl: process.env.WEBHOOK_URL,
    webhookSigningSecret: process.env.WEBHOOK_SIGNING_SECRET
  };
}

function parseTemplatesConfig(): TemplatesConfig {
  return {
    dir: process.env.TEMPLATES_DIR || 'src/templates'
  };
}

function parseQueueRuntimeConfig(): QueueRuntimeConfig {
  const batchSize = parseInt(process.env.QUEUE_BATCH_SIZE || '10', 10);
  if (isNaN(batchSize) || batchSize <= 0) throw new Error('QUEUE_BATCH_SIZE must be a positive number');
  const tickMs = parseInt(process.env.DISPATCH_TICK_MS || '1000', 10);
  if (isNaN(tickMs) || tickMs <= 0) throw new Error('DISPATCH_TICK_MS must be a positive number');
  const defaultMaxAttempts = parseInt(process.env.DEFAULT_MAX_ATTEMPTS || '3', 10);
  if (isNaN(defaultMaxAttempts) || defaultMaxAttempts <= 0) throw new Error('DEFAULT_MAX_ATTEMPTS must be a positive number');

  return { batchSize, tickMs, defaultMaxAttempts };
}

function validateRequiredEnvVars(vars: string[]): void {
  const missing = vars.filter(v => !process.env[v]);
  
  if (missing.length > 0) {
    throw new Error(`Missing required environment variables: ${missing.join(', ')}`);
  }
}
