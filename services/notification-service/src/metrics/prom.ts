import { Registry, Counter, Histogram, collectDefaultMetrics } from 'prom-client';
import type { FastifyInstance } from 'fastify';

export function createPrometheusRegistry(): Registry {
  const registry = new Registry();

  // Collect default Node.js metrics (cpu, memory, etc.)
  collectDefaultMetrics({ register: registry });

  return registry;
}

export function createNotificationMetrics(registry: Registry) {
  // Counter for notifications enqueued
  const notificationEnqueued = new Counter({
    name: 'notification_enqueued_total',
    help: 'Total number of notifications enqueued',
    labelNames: ['channel', 'type'],
    registers: [registry],
  });

  // Counter for notifications sent
  const notificationSent = new Counter({
    name: 'notification_sent_total',
    help: 'Total number of notifications sent',
    labelNames: ['channel', 'status'],
    registers: [registry],
  });

  // Histogram for notification send duration
  const notificationDuration = new Histogram({
    name: 'notification_send_duration_seconds',
    help: 'Duration of notification send operations in seconds',
    labelNames: ['channel'],
    buckets: [0.1, 0.5, 1, 2, 5, 10], // 100ms to 10s
    registers: [registry],
  });

  return {
    notificationEnqueued,
    notificationSent,
    notificationDuration,
  };
}

export function registerMetricsEndpoint(app: FastifyInstance, registry: Registry): void {
  // Decorate the app with the metrics registry
  app.decorate('metrics', registry);

  // Register the /metrics endpoint
  app.get('/metrics', async (_request, reply) => {
    const metrics = await registry.metrics();
    return reply.header('Content-Type', 'text/plain; version=0.0.4; charset=utf-8').send(metrics);
  });
}
