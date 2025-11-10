import { trace, context } from '@opentelemetry/api';

export type SpanContextLike = { traceId: string; spanId: string; traceFlags?: number };

export function getActiveSpanContext(): SpanContextLike | null {
  const span = trace.getActiveSpan();
  if (!span) return null;
  const ctx = span.spanContext();
  return { traceId: ctx.traceId, spanId: ctx.spanId, traceFlags: ctx.traceFlags };
}

export function formatTraceparent(ctx: SpanContextLike): string {
  const version = '00';
  const flags = (ctx.traceFlags ?? 1) & 1 ? '01' : '00';
  return `${version}-${ctx.traceId}-${ctx.spanId}-${flags}`;
}

export function getActiveTraceparent(): string | undefined {
  const ctx = getActiveSpanContext();
  if (!ctx) return undefined;
  return formatTraceparent(ctx);
}