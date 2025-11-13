export type Environment = 'dev' | 'staging' | 'prod';

export type NotificationChannel = 'email' | 'sms' | 'in-app' | 'webhook' | 'slack';

export interface ServiceConfig {
  name: string;
  port: number;
  env: Environment;
}

export interface ObservabilityConfig {
  serviceName: string;
  serviceVersion: string;
  otlpEndpoint: string;
  environment: Environment;
}

export interface NatsConfig {
  url: string;
  subjects: {
    notifications: string;
  };
}

export interface RedisConfig {
  url: string;
  maxRetriesPerRequest: number;
}

export interface PostgresConfig {
  host: string;
  port: number;
  database: string;
  user: string;
  password: string;
  maxConnections: number;
}

export interface SmtpConfig {
  host: string;
  port: number;
  user: string;
  password: string;
  from: string;
  secure: boolean;
}

export interface VaultConfig {
  addr: string;
  roleId: string;
  secretId: string;
}

export interface AuthConfig {
  jwtPublicKey?: string;
  jwtJwksUrl?: string;
}

export interface ChannelProvidersConfig {
  resendApiKey?: string;
  twilioAccountSid?: string;
  twilioAuthToken?: string;
  twilioFrom?: string;
  slackWebhookUrl?: string;
  webhookUrl?: string;
  webhookSigningSecret?: string;
}

export interface TemplatesConfig {
  dir: string;
}

export interface QueueRuntimeConfig {
  batchSize: number;
  tickMs: number;
  defaultMaxAttempts: number;
}

export interface ChannelRetryConfig {
  max: number;
  baseMs: number;
  jitterPct: number; // 0..1
}

export interface ChannelBreakerConfig {
  failureThreshold: number;
  windowSize: number;
  halfOpenAfterMs: number;
}

export interface ChannelReliabilityConfig {
  timeoutMs: number;
  retry: ChannelRetryConfig;
  breaker: ChannelBreakerConfig;
}

export interface Config {
  env: Environment;
  service: ServiceConfig;
  observability: ObservabilityConfig;
  nats: NatsConfig;
  redis: RedisConfig;
  postgres: PostgresConfig;
  smtp: SmtpConfig;
  vault: VaultConfig;
  auth: AuthConfig;
  channels: ChannelProvidersConfig;
  templates: TemplatesConfig;
  queue: QueueRuntimeConfig;
  reliability: ChannelReliabilityConfig;
}
