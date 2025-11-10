import type { FastifyInstance } from 'fastify';
import { trace } from '@opentelemetry/api';

function setIf(span: any, key: string, val: any) {
  if (!span) return;
  if (val === undefined || val === null || val === '') return;
  if (typeof span.setAttribute === 'function') span.setAttribute(key, val);
}

export function registerHttpTracing(app: FastifyInstance, opts?: { serviceName?: string }) {
  // Add lightweight hooks that enrich auto-instrumented http spans
  app.addHook('onRequest', async (req) => {
    const span: any = trace.getActiveSpan();
    if (!span) return;
    const route = (req as any).routerPath || (req as any).routeOptions?.url || (req as any).raw?.url;
    const method = (req as any).method || (req as any).raw?.method;
    if (typeof (span as any).updateName === 'function' && route && method) {
      (span as any).updateName(`${method} ${route}`);
    }
    setIf(span, 'http.route', route);
    setIf(span, 'http.request_id', (req as any).id || (req as any).headers?.['x-request-id']);
    setIf(span, 'enduser.id', (req as any).user?.sub);
    setIf(span, 'tenant.id', (req as any).user?.tenantId || (req as any).user?.tenant_id);
  });

  app.addHook('onResponse', async (req, reply) => {
    const span: any = trace.getActiveSpan();
    if (!span) return;
    setIf(span, 'http.status_code', reply.statusCode);
  });
}