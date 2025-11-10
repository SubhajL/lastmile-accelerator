import { register, Counter, Histogram, Gauge, collectDefaultMetrics } from 'prom-client';
import type { FastifyInstance } from 'fastify';

// Store metric instances for reuse
const metrics = new Map<string, Counter | Histogram | Gauge>();

// Collect default Node.js metrics
collectDefaultMetrics({ register });

/**
 * Registers GET /metrics endpoint on Fastify app.
 * Exports metrics in Prometheus text format.
 *
 * @param app - Fastify application instance
 */
export function createMetricsRoute(app: FastifyInstance): void {
  app.get('/metrics', async (_request, reply) => {
    const metrics = await register.metrics();
    reply.header('Content-Type', register.contentType);
    return metrics;
  });
}

/**
 * Helper to increment custom business metrics.
 * Creates counter on first call, reuses on subsequent calls.
 *
 * @param name - Counter name (should end with _total)
 * @param labels - Optional labels for dimensions
 * @param value - Amount to increment (default: 1)
 */
export function incrementCounter(
  name: string,
  labels?: Record<string, string>,
  value: number = 1
): void {
  let counter = metrics.get(name) as Counter;

  if (!counter) {
    const labelNames = labels ? Object.keys(labels) : [];
    counter = new Counter({
      name,
      help: `Counter for ${name}`,
      labelNames,
      registers: [register],
    });
    metrics.set(name, counter);
  }

  if (labels) {
    counter.inc(labels, value);
  } else {
    counter.inc(value);
  }
}

/**
 * Records histogram metric for latencies.
 * Creates histogram on first call with standard buckets.
 *
 * @param name - Histogram name
 * @param value - Value to record
 * @param labels - Optional labels for dimensions
 */
export function recordHistogram(
  name: string,
  value: number,
  labels?: Record<string, string>
): void {
  let histogram = metrics.get(name) as Histogram;

  if (!histogram) {
    const labelNames = labels ? Object.keys(labels) : [];
    histogram = new Histogram({
      name,
      help: `Histogram for ${name}`,
      labelNames,
      buckets: [10, 50, 100, 200, 500, 1000, 2000, 5000, 10000], // Standard latency buckets in ms
      registers: [register],
    });
    metrics.set(name, histogram);
  }

  if (labels) {
    histogram.observe(labels, value);
  } else {
    histogram.observe(value);
  }
}

/**
 * Sets gauge metric for current state values.
 * Creates gauge on first call, updates on subsequent calls.
 *
 * @param name - Gauge name
 * @param value - Value to set
 * @param labels - Optional labels for dimensions
 */
export function setGauge(
  name: string,
  value: number,
  labels?: Record<string, string>
): void {
  let gauge = metrics.get(name) as Gauge;

  if (!gauge) {
    const labelNames = labels ? Object.keys(labels) : [];
    gauge = new Gauge({
      name,
      help: `Gauge for ${name}`,
      labelNames,
      registers: [register],
    });
    metrics.set(name, gauge);
  }

  if (labels) {
    gauge.set(labels, value);
  } else {
    gauge.set(value);
  }
}

/**
 * Clears all custom metrics from the registry.
 * Used for testing to reset state between tests.
 */
export function resetMetrics(): void {
  // Unregister each custom metric from the registry
  for (const [name, metric] of metrics.entries()) {
    try {
      register.removeSingleMetric(name);
    } catch (e) {
      // Metric might not exist in registry, ignore
    }
  }
  metrics.clear();
}
