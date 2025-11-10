import type { FastifyInstance } from 'fastify';
import { getLogger } from '../logger';
import { trace } from '@opentelemetry/api';

function extractTrace() {
  const span: any = trace.getActiveSpan();
  if (!span) return {} as Record<string, unknown>;
  const ctx = span.spanContext();
  return { trace_id: ctx.traceId, span_id: ctx.spanId };
}

export function registerRequestLogging(app: FastifyInstance) {
  app.addHook('onRequest', async (req: any, reply: any) => {
    const base = getLogger();
    const correlation = {
      request_id: req?.id || req?.headers?.['x-request-id'],
      ...extractTrace(),
      user_id: req?.user?.sub,
      tenant_id: req?.user?.tenantId || req?.user?.tenant_id,
      method: req?.method || req?.raw?.method,
      route: req?.routerPath || req?.routeOptions?.url || req?.raw?.url,
    } as Record<string, unknown>;
    req.ctxLogger = base.child(correlation);

    if (reply && typeof reply.on === 'function') {
      reply.on('finish', () => {
        const log = req.ctxLogger || getLogger();
        const line = {
          status_code: reply?.statusCode,
          route: req?.routerPath || req?.routeOptions?.url || req?.raw?.url,
          method: req?.method || req?.raw?.method,
        };
        (log.info || log.debug || (()=>{})).call(log, { ...line }, 'request completed');
      });
    }
  });
}
