type Snapshot = {
  totalRequests: number;
  totalErrors: number;
  durations: { count: number; min: number; max: number; sum: number; avg: number };
};

const state = {
  totalRequests: 0,
  totalErrors: 0,
  durations: [] as number[],
};

import { getMeter } from '../otel';
let histogram: any | null = null;

function record(durationMs: number, statusCode: number, labels: Record<string, string>) {
  state.totalRequests += 1;
  if (statusCode >= 500) state.totalErrors += 1;
  state.durations.push(durationMs);
  try {
    if (!histogram) {
      const meter = getMeter('http');
      histogram = meter.createHistogram('http.server.duration', { description: 'HTTP server request duration (ms)', unit: 'ms' });
    }
    histogram.record(durationMs, labels);
  } catch {
    // OTel not initialized; skip
  }
}

export function resetMetrics() {
  state.totalRequests = 0;
  state.totalErrors = 0;
  state.durations.length = 0;
}

export function getMetricsSnapshot(): Snapshot {
  const count = state.durations.length;
  const min = count ? Math.min(...state.durations) : 0;
  const max = count ? Math.max(...state.durations) : 0;
  const sum = state.durations.reduce((a, b) => a + b, 0);
  const avg = count ? sum / count : 0;
  return {
    totalRequests: state.totalRequests,
    totalErrors: state.totalErrors,
    durations: { count, min, max, sum, avg },
  };
}

export function metricsMiddleware() {
  return async function (request: any, reply: any) {
    const start = Date.now();
    // When the response finishes, record metrics
    if (reply && typeof reply.on === 'function') {
      reply.on('finish', () => {
        const duration = Math.max(0, Date.now() - start);
        const statusCode = typeof reply.statusCode === 'number' ? reply.statusCode : 0;
        const req: any = request as any;
        record(duration, statusCode, {
          method: String(req.method || req.raw?.method || ''),
          route: String((req as any).routerPath || (req as any).routeOptions?.url || req.raw?.url || ''),
          status_code: String(statusCode),
          tenant_id: String((req as any).user?.tenantId || (req as any).user?.tenant_id || ''),
        });
      });
    }
  };
}