import type { FastifyInstance } from 'fastify';
import { getMetricsSnapshot } from '../middleware/metrics';

export function registerHealthRoutes(app: FastifyInstance, opts?: { serviceName?: string }) {
  const serviceName = opts?.serviceName ?? process.env.SERVICE_NAME ?? 'projects-service';

  app.get('/healthz', async () => {
    return {
      status: 'ok',
      service: serviceName,
      timestamp: new Date().toISOString(),
    };
  });

  app.get('/metrics', async (request, reply) => {
    const snap = getMetricsSnapshot();
    const lines: string[] = [];

    const labels = `{service_name="${serviceName}"}`;
    lines.push(`# HELP request_total Total number of HTTP requests`);
    lines.push(`# TYPE request_total counter`);
    lines.push(`request_total${labels} ${snap.totalRequests}`);

    lines.push(`# HELP request_errors_total Total number of HTTP 5xx responses`);
    lines.push(`# TYPE request_errors_total counter`);
    lines.push(`request_errors_total${labels} ${snap.totalErrors}`);

    lines.push(`# HELP request_duration_ms Summary of request durations in milliseconds`);
    lines.push(`# TYPE request_duration_ms summary`);
    lines.push(`request_duration_ms_count${labels} ${snap.durations.count}`);
    lines.push(`request_duration_ms_sum${labels} ${snap.durations.sum}`);

    const body = lines.join('\n') + '\n';
    reply.header('Content-Type', 'text/plain; version=0.0.4');
    reply.header('Cache-Control', 'no-cache');
    return reply.send(body);
  });
}