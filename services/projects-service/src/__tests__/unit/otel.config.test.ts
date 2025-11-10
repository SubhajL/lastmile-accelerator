import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { computeOtelConfigFromEnv } from '../../otel';

const OLD = { ...process.env } as Record<string, string | undefined>;

describe('otel config from env', () => {
  beforeEach(() => {
    Object.keys(process.env).forEach((k) => {
      if (k.startsWith('OTEL_')) delete (process.env as any)[k];
    });
  });
  afterEach(() => {
    Object.assign(process.env, OLD);
  });

  it('reads exporter and metrics endpoints and sampler', () => {
    process.env.OTEL_EXPORTER_OTLP_ENDPOINT = 'http://otel:4318/v1/traces';
    process.env.OTEL_EXPORTER_OTLP_METRICS_ENDPOINT = 'http://otel:4318/v1/metrics';
    process.env.OTEL_TRACES_SAMPLER = 'traceidratio';
    process.env.OTEL_TRACES_SAMPLER_ARG = '0.25';
    const cfg = computeOtelConfigFromEnv();
    expect(cfg.exporterUrl).toBe('http://otel:4318/v1/traces');
    expect(cfg.metricsUrl).toBe('http://otel:4318/v1/metrics');
    expect(cfg.sampler).toBe('traceidratio');
    expect(cfg.samplerArg).toBe('0.25');
  });

  it('returns undefined values when not set', () => {
    const cfg = computeOtelConfigFromEnv();
    expect(cfg.exporterUrl).toBeUndefined();
    expect(cfg.metricsUrl).toBeUndefined();
  });
});