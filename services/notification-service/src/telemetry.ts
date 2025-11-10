import { NodeSDK } from '@opentelemetry/sdk-node';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';
import { Resource } from '@opentelemetry/resources';
import { SemanticResourceAttributes } from '@opentelemetry/semantic-conventions';
import { HttpInstrumentation } from '@opentelemetry/instrumentation-http';
import { FastifyInstrumentation } from '@opentelemetry/instrumentation-fastify';
import { trace } from '@opentelemetry/api';
import type { Tracer } from '@opentelemetry/api';
import type { ObservabilityConfig } from './types.js';

export function initTelemetry(config: ObservabilityConfig): NodeSDK {
  const resource = new Resource({
    [SemanticResourceAttributes.SERVICE_NAME]: config.serviceName,
    [SemanticResourceAttributes.SERVICE_VERSION]: config.serviceVersion,
    [SemanticResourceAttributes.DEPLOYMENT_ENVIRONMENT]: config.environment
  });

  const traceExporter = new OTLPTraceExporter({
    url: `${config.otlpEndpoint}/v1/traces`
  });

  const sdk = new NodeSDK({
    resource,
    traceExporter,
    instrumentations: [
      new HttpInstrumentation({
        ignoreIncomingRequestHook: (req) => {
          // Ignore health checks and metrics
          const url = req.url || '';
          return url.includes('/healthz') || url.includes('/metrics');
        }
      }),
      new FastifyInstrumentation()
    ]
  });

  sdk.start();

  process.on('SIGTERM', () => {
    sdk.shutdown()
      .then(() => console.log('Telemetry terminated'))
      .catch((error) => console.error('Error terminating telemetry', error));
  });

  return sdk;
}

export async function shutdownTelemetry(sdk: NodeSDK): Promise<void> {
  await sdk.shutdown();
}

export function createTracer(name: string): Tracer {
  return trace.getTracer(name);
}
