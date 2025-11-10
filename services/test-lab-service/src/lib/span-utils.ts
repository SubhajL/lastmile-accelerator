import { trace, context, Span, SpanStatusCode } from '@opentelemetry/api';
import type { Attributes } from '@opentelemetry/api';

/**
 * Wraps an async function in a traced span with automatic error recording.
 * Captures function execution time and success/failure status.
 *
 * @param name - Name of the span
 * @param fn - Async function to execute within the span
 * @param attributes - Optional attributes to add to the span
 * @returns Promise resolving to the function result
 * @throws Re-throws any error from the function after recording it
 */
export async function withSpan<T>(
  name: string,
  fn: () => Promise<T>,
  attributes?: Record<string, string | number | boolean>
): Promise<T> {
  const tracer = trace.getTracer('test-lab-service');
  const span = tracer.startSpan(name);

  // Add custom attributes if provided
  if (attributes) {
    span.setAttributes(attributes as Attributes);
  }

  // Execute function within active span context
  const activeContext = trace.setSpan(context.active(), span);
  
  try {
    const result = await context.with(activeContext, fn);
    span.setStatus({ code: SpanStatusCode.OK });
    return result;
  } catch (error) {
    span.setStatus({
      code: SpanStatusCode.ERROR,
      message: error instanceof Error ? error.message : 'Unknown error',
    });
    span.recordException(error instanceof Error ? error : new Error(String(error)));
    throw error;
  } finally {
    span.end();
  }
}

/**
 * Returns the currently active span from context.
 * Useful for adding attributes to existing spans from business logic.
 *
 * @returns Current active span or undefined if no span is active
 */
export function getCurrentSpan(): Span | undefined {
  return trace.getActiveSpan();
}

/**
 * Adds attributes to the current active span.
 * Validates attribute types (string, number, boolean only).
 *
 * @param attributes - Key-value pairs to add as span attributes
 */
export function addSpanAttributes(
  attributes: Record<string, string | number | boolean>
): void {
  const span = getCurrentSpan();
  if (!span) {
    return; // No-op if no active span
  }

  span.setAttributes(attributes as Attributes);
}

/**
 * Adds a timestamped event to the current span.
 * Useful for marking important moments within a span's lifetime.
 *
 * @param name - Event name (e.g., 'test_started', 'browser_session_created')
 * @param attributes - Optional event attributes
 */
export function addSpanEvent(
  name: string,
  attributes?: Record<string, unknown>
): void {
  const span = getCurrentSpan();
  if (!span) {
    return; // No-op if no active span
  }

  span.addEvent(name, attributes as Attributes);
}

/**
 * Records an exception on the current span with stack trace.
 * Marks span status as ERROR.
 *
 * @param error - Error instance to record
 * @param attributes - Optional additional attributes (e.g., error code, severity)
 */
export function recordException(
  error: Error,
  attributes?: Record<string, string>
): void {
  const span = getCurrentSpan();
  if (!span) {
    return; // No-op if no active span
  }

  span.recordException(error);
  span.setStatus({
    code: SpanStatusCode.ERROR,
    message: error.message,
  });

  if (attributes) {
    span.setAttributes(attributes as Attributes);
  }
}
