import { loadConfig } from './config.js';
import { createHandlebarsRenderer } from './channels/renderer.js';
import { createChannelRegistry } from './channels/registry.js';
import { createRuntime } from './bootstrap/runtime.js';
import { createRunLoop } from './bootstrap/runloop.js';
import { createResendChannel } from './channels/resend.js';
import { createFallbackChannel } from './channels/fallback.js';
import { getRecipientEmailFromJob } from './recipients/resolvers.js';
import { getRecipientPhoneFromJob } from './recipients/phone.js';
import { createOtelMetrics } from './metrics/otel.js';
import { createTwilioSmsChannel } from './channels/sms/twilio.js';
import { createRateLimitedChannel } from './channels/sms/limiter.js';
import { createInAppRepo } from './channels/inapp/repo.js';
import { createInAppPublisher } from './channels/inapp/pubsub.js';
import { createInAppChannel } from './channels/inapp/channel.js';
import { createCircuitBreaker } from './channels/circuitbreaker.js';
import { createWebhookChannel } from './channels/webhook/http.js';
import { createSlackChannel } from './channels/slack/slack.js';
import { createNatsSubscribe } from './events/nats-subscribe.js';
import { createNotificationQueue } from './notifications/queue.js';
import { createDbClient } from './db/client.js';
import { SubscriptionSubjects } from './events/subjects.js';
import { createRedisClient } from './redis/client.js';
import { createTemplateStore } from './templates/store.js';
import { createTemplateLoader } from './templates/loader.js';
import { connect as natsConnect } from 'nats';
import IORedis from 'ioredis';
import { createPreferenceStore } from './prefs/store.js';
import { createRoutingEngine } from './routing/engine.js';

export async function bootstrap() {
  const cfg = loadConfig();

  // Redis + queue
  const redis = createRedisClient(cfg.redis, { RedisCtor: IORedis, logger: console });
  const db = createDbClient(cfg.postgres);
  const queue = createNotificationQueue(redis as any, {
    db: db as any,
    outboxDedupEnabled: Boolean(cfg.features?.outboxEnqueueDedup)
  });

  // NATS
  const nc = await natsConnect({ servers: cfg.nats.url });
  const subscribe = createNatsSubscribe(nc as any);

  // Templates loader (store + filesystem)
  const templateStore = createTemplateStore({ redis: redis as any, namespace: 'tmpl' });
  const templateLoader = createTemplateLoader({ store: templateStore as any, fsBaseDir: cfg.templates.dir });

  // Channels (email)
  const renderTemplate = createHandlebarsRenderer({ templatesDir: cfg.templates.dir, loader: templateLoader as any });
  // Real Nodemailer transporter wiring
  const nodemailer = await import('nodemailer');
  const transporter = nodemailer.createTransport({
    host: cfg.smtp.host,
    port: cfg.smtp.port,
    secure: cfg.smtp.secure || cfg.smtp.port === 465,
    auth: { user: cfg.smtp.user, pass: cfg.smtp.password }
  });

  // Primary email adapter (nodemailer)
  const primaryEmail = (await import('./channels/email.js')).createEmailChannel({
    transporter: transporter as any,
    renderTemplate,
    from: cfg.smtp.from,
    resolveTo: getRecipientEmailFromJob,
    reliability: { timeoutMs: cfg.reliability.timeoutMs, retry: cfg.reliability.retry }
  }) as any;

  // Optional Resend fallback
  const resendApiKey = cfg.channels.resendApiKey;
  const finalEmailAdapter = resendApiKey ?
    createFallbackChannel({
      primary: primaryEmail as any,
      fallback: createResendChannel({
        http: (globalThis as any).fetch,
        apiKey: resendApiKey,
        from: cfg.smtp.from,
        resolveTo: getRecipientEmailFromJob,
        renderTemplate
      }) as any
    }) :
    primaryEmail as any;

  // Metrics (OTel)
  const { metrics: otelMetrics } = await import('@opentelemetry/api');
  const meter = (otelMetrics.getMeterProvider?.().getMeter?.(cfg.observability.serviceName)) || { createCounter: () => ({ add: () => {} }) } as any;
  const metrics = createOtelMetrics({ meter, serviceName: cfg.observability.serviceName });

  const adapters: Record<string, any> = { email: finalEmailAdapter };

  // SMS wiring if Twilio configured
  if (cfg.channels.twilioAccountSid && cfg.channels.twilioAuthToken && cfg.channels.twilioFrom) {
    const smsPrimary = createTwilioSmsChannel({ client: { messages: { create: async () => ({}) } } as any, from: cfg.channels.twilioFrom, resolveTo: getRecipientPhoneFromJob, renderTemplate, metrics, reliability: { timeoutMs: cfg.reliability.timeoutMs, retry: cfg.reliability.retry } });
    // Simple limiter: allow 5 per second burst 10
    let tokens = 10; let last = Date.now();
    const limiter = { tryRemoveTokens: () => { const now = Date.now(); const refill = Math.floor((now - last)/200); if (refill>0){ tokens=Math.min(10, tokens+refill); last=now; } if (tokens>0){ tokens--; return true; } return false; } };
    adapters['sms'] = createRateLimitedChannel({ inner: smsPrimary as any, limiter });
  }

  // In-app channel wiring (always available using same Redis)
  const inappRepo = createInAppRepo({ redis: redis as any, namespace: 'inapp' });
  const inappPub = createInAppPublisher({ redis: redis as any, channelPrefix: 'inapp', repo: inappRepo as any, metrics });
  adapters['in-app'] = createInAppChannel({ publisher: inappPub as any, resolveToUserId: async (job:any) => job.userId });

  // Webhook/Slack wiring
  const breaker = createCircuitBreaker({
    failureThreshold: cfg.reliability.breaker.failureThreshold,
    halfOpenAfterMs: cfg.reliability.breaker.halfOpenAfterMs,
    windowSize: cfg.reliability.breaker.windowSize,
    now: () => Date.now()
  });
  if (cfg.channels.webhookUrl) {
    adapters['webhook'] = createWebhookChannel({ http: (globalThis as any).fetch, url: cfg.channels.webhookUrl, signingSecret: cfg.channels.webhookSigningSecret, metrics, breaker, reliability: { timeoutMs: cfg.reliability.timeoutMs, retry: cfg.reliability.retry } });
  }
  if (cfg.channels.slackWebhookUrl) {
    adapters['slack'] = createSlackChannel({ http: (globalThis as any).fetch, webhookUrl: cfg.channels.slackWebhookUrl, metrics, breaker, reliability: { timeoutMs: cfg.reliability.timeoutMs, retry: cfg.reliability.retry } });
  }

  const registry = createChannelRegistry(adapters);

  // Runtime

  // Preferences & routing engine
  const prefStore = createPreferenceStore({ redis: redis as any, namespace: 'np' });
  const routingEngine = createRoutingEngine({ prefs: prefStore as any, now: () => Date.now() });

  const runtime = createRuntime({
    queue: queue as any,
    subscribe,
    dispatcherFactory: undefined,
    metrics,
    now: () => Date.now(),
    subjects: [...SubscriptionSubjects],
    batchSize: cfg.queue.batchSize,
    email: {
      from: cfg.smtp.from,
      transporter: transporter as any,
      renderTemplate,
      resolveTo: getRecipientEmailFromJob
    },
    defaultMaxAttempts: cfg.queue.defaultMaxAttempts,
    registry,
    routingEngine
  });

  const runloop = createRunLoop({ dispatcher: runtime as any, batchSize: cfg.queue.batchSize, tickMs: cfg.queue.tickMs, logger: console });
  runloop.start();

  async function stop() {
    runloop.stop();
    try { await (nc as any).close?.(); } catch { /* noop */ }
    try { await (redis as any).quit?.(); } catch { /* noop */ }
  }

  return { stop };
}

if (import.meta.url === `file://${process.argv[1]}`) {
  // Fire-and-forget start when invoked directly
  bootstrap();
}
