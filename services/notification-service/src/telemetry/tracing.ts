import { SpanStatusCode, trace, context, type Attributes } from '@opentelemetry/api';

export function getTracer(serviceName: string) {
  return trace.getTracer(serviceName);
}

export async function withSpan<T>(name: string, attrs: Attributes, fn: () => Promise<T>): Promise<T> {
  const tracer = getTracer('notification-service');
  const options = attrs && Object.keys(attrs as Record<string, unknown>).length ? { attributes: attrs } : undefined;
  const span = tracer.startSpan(name, options as unknown as { attributes?: Attributes });
  try {
    if (attrs && Object.keys(attrs as Record<string, unknown>).length) span.setAttributes(attrs);
    const ctx = trace.setSpan(context.active(), span);
    const res = await context.with(ctx, fn);
    span.setStatus({ code: SpanStatusCode.OK });
    return res;
  } catch (err) {
    if (err instanceof Error) span.recordException(err);
    span.setStatus({ code: SpanStatusCode.ERROR, message: err instanceof Error ? err.message : String(err) });
    throw err;
  } finally {
    span.end();
  }
}

export function recordRetry(span: { addEvent: (name: string, attrs?: Record<string, unknown>) => void }, attempt: number, reason?: string) {
  span.addEvent('retry', { attempt, ...(reason ? { reason } : {}) });
}
