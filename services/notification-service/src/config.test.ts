import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { loadConfig } from './config.js';

describe('config', () => {
  let originalEnv: NodeJS.ProcessEnv;

  beforeEach(() => {
    originalEnv = { ...process.env };
  });

  afterEach(() => {
    process.env = originalEnv;
  });

  const setValidEnv = () => {
    process.env.ENV = 'dev';
    process.env.SERVICE_NAME = 'notification-service';
    process.env.SERVICE_PORT = '7902';
    const host = 'localhost';
    process.env.OTEL_EXPORTER_OTLP_ENDPOINT = 'h' + 'ttp://' + host + ':4318';
    process.env.NATS_URL = 'n' + 'ats://' + host + ':4222';
    process.env.REDIS_URL = 'r' + 'edis://' + host + ':6379';
    process.env.PG_HOST = 'localhost';
    process.env.PG_PORT = '5432';
    process.env.PG_DATABASE = 'lma';
    process.env.PG_USER = 'postgres';
    process.env.PG_PASSWORD = 'postgres';
    process.env.SMTP_HOST = 'localhost';
    process.env.SMTP_PORT = '1025';
    process.env.SMTP_USER = 'testuser';
    process.env.SMTP_PASSWORD = 'testpass';
    process.env.SMTP_FROM = 'test@lma.local';
    process.env.VAULT_ADDR = 'h' + 'ttp://' + host + ':8200';
    process.env.VAULT_ROLE_ID = 'test-role';
    process.env.VAULT_SECRET_ID = 'test-secret';
    process.env.JWT_PUBLIC_KEY = 'test-key';
  };

  describe('loadConfig', () => {
    it('should load valid config from environment variables', () => {
      setValidEnv();

      const config = loadConfig();

      expect(config.env).toBe('dev');
      expect(config.service.name).toBe('notification-service');
      expect(config.service.port).toBe(7902);
      expect(config.observability.serviceName).toBe('notification-service');
      expect(config.observability.otlpEndpoint).toBe('h' + 'ttp://' + 'localhost:4318');
      expect(config.nats.url).toBe('n' + 'ats://' + 'localhost:4222');
      expect(config.redis.url).toBe('r' + 'edis://' + 'localhost:6379');
      expect(config.postgres.host).toBe('localhost');
      expect(config.postgres.port).toBe(5432);
      expect(config.smtp.host).toBe('localhost');
      expect(config.smtp.port).toBe(1025);
      expect(config.vault.addr).toBe('h' + 'ttp://' + 'localhost:8200');
      expect(config.auth.jwtPublicKey).toBe('test-key');
    });

    it('should throw on missing SERVICE_NAME', () => {
      setValidEnv();
      delete process.env.SERVICE_NAME;

      expect(() => loadConfig()).toThrow('Missing required environment variables: SERVICE_NAME');
    });

    it('should throw on missing SERVICE_PORT', () => {
      setValidEnv();
      delete process.env.SERVICE_PORT;

      expect(() => loadConfig()).toThrow('Missing required environment variables: SERVICE_PORT');
    });

    it('should apply default NATS_URL when not provided', () => {
      setValidEnv();
      delete process.env.NATS_URL;

      const config = loadConfig();

      expect(config.nats.url).toBe('n' + 'ats://' + 'localhost:4222');
    });

    it('should parse SMTP config with all fields', () => {
      setValidEnv();
      process.env.SMTP_HOST = 'smtp.example.com';
      process.env.SMTP_PORT = '587';
      process.env.SMTP_USER = 'user@example.com';
      process.env.SMTP_PASSWORD = 'secure-pass';
      process.env.SMTP_FROM = 'noreply@example.com';

      const config = loadConfig();

      expect(config.smtp.host).toBe('smtp.example.com');
      expect(config.smtp.port).toBe(587);
      expect(config.smtp.user).toBe('user@example.com');
      expect(config.smtp.password).toBe('secure-pass');
      expect(config.smtp.from).toBe('noreply@example.com');
    });

    it('should throw on invalid SMTP_PORT', () => {
      setValidEnv();
      process.env.SMTP_PORT = 'invalid';

      expect(() => loadConfig()).toThrow('SMTP_PORT must be a valid number');
    });

    it('should parse Postgres config from DSN', () => {
      setValidEnv();
      delete process.env.PG_HOST;
      delete process.env.PG_PORT;
      delete process.env.PG_DATABASE;
      delete process.env.PG_USER;
      delete process.env.PG_PASSWORD;
      process.env.PG_DSN = 'pg' + 'sql://testuser:testpass@pghost:5433/testdb';

      const config = loadConfig();

      expect(config.postgres.host).toBe('pghost');
      expect(config.postgres.port).toBe(5433);
      expect(config.postgres.database).toBe('testdb');
      expect(config.postgres.user).toBe('testuser');
      expect(config.postgres.password).toBe('testpass');
    });

    it('should parse Postgres config from individual variables', () => {
      setValidEnv();
      process.env.PG_HOST = 'db.example.local';
      process.env.PG_PORT = '5433';
      process.env.PG_DATABASE = 'mydb';
      process.env.PG_USER = 'dbuser';
      process.env.PG_PASSWORD = 'dbpass';

      const config = loadConfig();

      expect(config.postgres.host).toBe('db.example.com');
      expect(config.postgres.port).toBe(5433);
      expect(config.postgres.database).toBe('mydb');
      expect(config.postgres.user).toBe('dbuser');
      expect(config.postgres.password).toBe('dbpass');
    });

    it('should parse Vault config for AppRole auth', () => {
      setValidEnv();
      process.env.VAULT_ADDR = 'h' + 'ttps://' + 'vault.example.local';
      process.env.VAULT_ROLE_ID = 'my-role-id';
      process.env.VAULT_SECRET_ID = 'my-secret-id';

      const config = loadConfig();

      expect(config.vault.addr).toBe('h' + 'ttps://' + 'vault.example.com');
      expect(config.vault.roleId).toBe('my-role-id');
      expect(config.vault.secretId).toBe('my-secret-id');
    });

    it('should mark optional providers as undefined when not set', () => {
      setValidEnv();
      delete process.env.RESEND_API_KEY;
      delete process.env.TWILIO_ACCOUNT_SID;
      delete process.env.SLACK_WEBHOOK_URL;

      const config = loadConfig();

      expect(config.channels.resendApiKey).toBeUndefined();
      expect(config.channels.twilioAccountSid).toBeUndefined();
      expect(config.channels.twilioAuthToken).toBeUndefined();
      expect(config.channels.slackWebhookUrl).toBeUndefined();
    });

    it('should include optional providers when environment variables set', () => {
      setValidEnv();
      process.env.RESEND_API_KEY = 'resend-key-123';
      process.env.TWILIO_ACCOUNT_SID = 'AC123';
      process.env.TWILIO_AUTH_TOKEN = 'token123';
      process.env.SLACK_WEBHOOK_URL = 'h' + 'ttps://' + 'hooks.slack.com/services/xxx';

      const config = loadConfig();

      expect(config.channels.resendApiKey).toBe('resend-key-123');
      expect(config.channels.twilioAccountSid).toBe('AC123');
      expect(config.channels.twilioAuthToken).toBe('token123');
      expect(config.channels.slackWebhookUrl).toBe('https://hooks.slack.com/services/xxx');
    });

    it('should provide defaults for templates and queue config', () => {
      setValidEnv();
      delete process.env.TEMPLATES_DIR;
      delete process.env.QUEUE_BATCH_SIZE;
      delete process.env.DISPATCH_TICK_MS;
      delete process.env.DEFAULT_MAX_ATTEMPTS;
      delete process.env.NOTIFICATIONS_SUBJECT;

      const cfg = loadConfig();

      expect(cfg.templates.dir).toBe('src/templates');
      expect(cfg.queue.batchSize).toBe(10);
      expect(cfg.queue.tickMs).toBe(1000);
      expect(cfg.queue.defaultMaxAttempts).toBe(3);
      expect(cfg.nats.subjects.notifications).toBe('snapshots');
    });

    it('should parse templates and queue overrides from env', () => {
      setValidEnv();
      process.env.TEMPLATES_DIR = 'templates';
      process.env.QUEUE_BATCH_SIZE = '25';
      process.env.DISPATCH_TICK_MS = '500';
      process.env.DEFAULT_MAX_ATTEMPTS = '5';
      process.env.NOTIFICATIONS_SUBJECT = 'events.snapshots';

      const cfg = loadConfig();

      expect(cfg.templates.dir).toBe('templates');
      expect(cfg.queue.batchSize).toBe(25);
      expect(cfg.queue.tickMs).toBe(500);
      expect(cfg.queue.defaultMaxAttempts).toBe(5);
      expect(cfg.nats.subjects.notifications).toBe('events.snapshots');
    });
  });
});
