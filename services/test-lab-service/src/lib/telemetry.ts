import { NodeSDK } from '@opentelemetry/sdk-node';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';
import { Resource } from '@opentelemetry/resources';
import { SEMRESATTRS_SERVICE_NAME, SEMRESATTRS_SERVICE_VERSION, SEMRESATTRS_DEPLOYMENT_ENVIRONMENT } from '@opentelemetry/semantic-conventions';
import { HttpInstrumentation } from '@opentelemetry/instrumentation-http';
import { FastifyInstrumentation } from '@opentelemetry/instrumentation-fastify';
import { PgInstrumentation } from '@opentelemetry/instrumentation-pg';
import { RedisInstrumentation as Redis4Instrumentation } from '@opentelemetry/instrumentation-redis-4';
import { trace } from '@opentelemetry/api';
import type { Tracer } from '@opentelemetry/api';
import type { IncomingMessage } from 'http';

export interface TelemetryConfig {
  serviceName: string;
  serviceVersion: string;
  otlpEndpoint: string;
  environment: 'dev' | 'staging' | 'prod';
}

/**
 * Factory function to create hook for ignoring specific routes from tracing.
 * Prevents /healthz and /metrics from creating unnecessary spans.
 *
 * @returns Hook function for HttpInstrumentation
 */
function configureIgnoredRoutes() {
  return (req: IncomingMessage): boolean => {
    const url = req.url || '';
    return url.includes('/healthz') || url.includes('/metrics');
  };
}

/**
 * Initialize OpenTelemetry SDK with auto-instrumentation.
 * Configures tracing for HTTP, Fastify, Postgres, and Redis.
 * Sets up resource attributes for service identification.
 *
 * @param config - Telemetry configuration
 * @returns Initialized NodeSDK instance
 */
export function initTelemetry(config: TelemetryConfig): NodeSDK {
  const resource = new Resource({
    [SEMRESATTRS_SERVICE_NAME]: config.serviceName,
    [SEMRESATTRS_SERVICE_VERSION]: config.serviceVersion,
    [SEMRESATTRS_DEPLOYMENT_ENVIRONMENT]: config.environment,
  });

  const traceExporter = new OTLPTraceExporter({
    url: `${config.otlpEndpoint}/v1/traces`,
  });

  const sdk = new NodeSDK({
    resource,
    traceExporter,
    instrumentations: [
      new HttpInstrumentation({
        ignoreIncomingRequestHook: configureIgnoredRoutes(),
      }),
      new FastifyInstrumentation(),
      new PgInstrumentation({
        enhancedDatabaseReporting: true,
      }),
      new Redis4Instrumentation(),
    ],
  });

  sdk.start();

  // Register graceful shutdown handler
  process.on('SIGTERM', () => {
    sdk
      .shutdown()
      .then(() => console.log('Telemetry shutdown complete'))
      .catch((error) => console.error('Error shutting down telemetry', error));
  });

  return sdk;
}

/**
 * Gracefully shutdown the OpenTelemetry SDK.
 * Flushes all pending spans to the collector.
 *
 * @param sdk - NodeSDK instance to shutdown
 */
export async function shutdownTelemetry(sdk: NodeSDK): Promise<void> {
  await sdk.shutdown();
}

/**
 * Get a tracer instance for manual instrumentation.
 * Use this to create custom spans for business logic.
 *
 * @param name - Tracer name (typically the module or component name)
 * @returns Tracer instance
 */
export function getTracer(name: string): Tracer {
  return trace.getTracer(name);
}
