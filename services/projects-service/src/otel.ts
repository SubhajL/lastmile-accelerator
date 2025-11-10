// OpenTelemetry bootstrap with injectable runtime to keep tests fast and dependency-light.
// The real implementation can lazy-load @opentelemetry/* packages.

let runtime: {
  start: (serviceName: string, options?: { exporterUrl?: string; metricsUrl?: string }) => Promise<void>;
  getTracer: (name?: string) => any;
  getMeter: (name?: string) => any;
  shutdown: () => Promise<void>;
  getActiveSpanContext?: () => { traceId: string; spanId: string; traceFlags?: number } | null;
} | null = null;

export function __setOtelRuntime(rt: {
  start: (serviceName: string, options?: { exporterUrl?: string; metricsUrl?: string }) => Promise<void>;
  getTracer: (name?: string) => any;
  getMeter: (name?: string) => any;
  shutdown: () => Promise<void>;
  getActiveSpanContext?: () => { traceId: string; spanId: string; traceFlags?: number } | null;
}) {
  runtime = rt;
  // expose for tests' convenience to access spies
  (global as any).__otelRuntime = rt;
}

async function getDefaultRuntime() {
  // Lazy import the heavy SDKs only if a test didn't inject a runtime.
  const dynImport = new Function('m', 'return import(m)') as (m: string) => Promise<any>;
  const [sdkNode, traceExp, metricsExp, resources, semconv, api, autoInst] = await Promise.all([
    dynImport('@opentelemetry/sdk-node'),
    dynImport('@opentelemetry/exporter-trace-otlp-http'),
    dynImport('@opentelemetry/exporter-metrics-otlp-http'),
    dynImport('@opentelemetry/resources'),
    dynImport('@opentelemetry/semantic-conventions'),
    dynImport('@opentelemetry/api'),
    dynImport('@opentelemetry/auto-instrumentations-node'),
  ]);
  const { NodeSDK } = sdkNode;
  const { OTLPTraceExporter } = traceExp;
  const { OTLPMetricExporter } = metricsExp;
  const { Resource } = resources;
  const { SemanticResourceAttributes } = semconv;
  const { getNodeAutoInstrumentations } = autoInst;

  let sdk: any;

  return {
    async start(serviceName: string, options?: { exporterUrl?: string; metricsUrl?: string }) {
      const traceExporter = new OTLPTraceExporter({ url: options?.exporterUrl });
      const metricExporter = new OTLPMetricExporter({ url: options?.metricsUrl || options?.exporterUrl });
      sdk = new NodeSDK({
        resource: new Resource({ [SemanticResourceAttributes.SERVICE_NAME]: serviceName }),
        traceExporter,
        metricReader: new (await dynImport('@opentelemetry/sdk-metrics')).PeriodicExportingMetricReader({ exporter: metricExporter }),
        instrumentations: [getNodeAutoInstrumentations()],
      });
      await sdk.start();
    },
    getTracer(name?: string) {
      return api.trace.getTracer(name ?? 'default');
    },
    getMeter(name?: string) {
      return api.metrics.getMeter(name ?? 'default');
    },
    async shutdown() {
      if (sdk) await sdk.shutdown();
    },
    getActiveSpanContext() {
      const span = api.trace.getActiveSpan();
      if (!span) return null;
      const ctx = span.spanContext();
      return { traceId: ctx.traceId, spanId: ctx.spanId, traceFlags: ctx.traceFlags };
    },
  };
}

export async function initOtel(serviceName: string, options?: { exporterUrl?: string; metricsUrl?: string }): Promise<void> {
  if (!runtime) {
    runtime = await getDefaultRuntime();
  }
  await runtime.start(serviceName, options);
}

export function computeOtelConfigFromEnv() {
  const exporterUrl = process.env.OTEL_EXPORTER_OTLP_ENDPOINT;
  const metricsUrl = process.env.OTEL_EXPORTER_OTLP_METRICS_ENDPOINT;
  const sampler = (process.env.OTEL_TRACES_SAMPLER || '').toLowerCase();
  const samplerArg = process.env.OTEL_TRACES_SAMPLER_ARG;
  return { exporterUrl, metricsUrl, sampler, samplerArg };
}

export function getTracer(name?: string): any {
  if (!runtime) throw new Error('OTel not initialized. Call initOtel() first.');
  return runtime.getTracer(name);
}

export function getMeter(name?: string): any {
  if (!runtime) throw new Error('OTel not initialized. Call initOtel() first.');
  return runtime.getMeter(name);
}

export async function closeOtel(): Promise<void> {
  if (!runtime) return;
  await runtime.shutdown();
  runtime = null;
}