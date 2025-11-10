import { describe, it, expect, beforeEach, vi } from 'vitest';
import { initOtel, getTracer, getMeter, closeOtel, __setOtelRuntime } from '../../otel';

describe('otel.ts', () => {
  beforeEach(async () => {
    // Inject mock runtime per test
    const start = vi.fn(async (_service: string, _opts?: any) => {});
    const tracer = { name: 'default-tracer', startSpan: vi.fn() };
    const meter = { name: 'default-meter', createCounter: vi.fn() };
    const getTracerImpl = vi.fn((name?: string) => ({ ...tracer, name: name ?? 'default-tracer' }));
    const getMeterImpl = vi.fn((name?: string) => ({ ...meter, name: name ?? 'default-meter' }));
    const shutdown = vi.fn(async () => {});

    __setOtelRuntime({ start, getTracer: getTracerImpl, getMeter: getMeterImpl, shutdown });

    // Ensure any prior runtime is cleanly closed
    await closeOtel().catch(() => {});
    __setOtelRuntime({ start, getTracer: getTracerImpl, getMeter: getMeterImpl, shutdown });
  });

  it('initializes OTel with service.name and allows tracer/meter access', async () => {
    const startSpy = (global as any).__otelRuntime.start as ReturnType<typeof vi.fn>;
    const getTracerSpy = (global as any).__otelRuntime.getTracer as ReturnType<typeof vi.fn>;
    const getMeterSpy = (global as any).__otelRuntime.getMeter as ReturnType<typeof vi.fn>;

    await initOtel('projects-service', { exporterUrl: 'http://collector:4318' });

    expect(startSpy).toHaveBeenCalledWith('projects-service', { exporterUrl: 'http://collector:4318' });

    const tracer = getTracer('api');
    const meter = getMeter('api');
    expect(tracer.name).toBe('api');
    expect(meter.name).toBe('api');

    // startSpan available on tracer
    expect(typeof tracer.startSpan).toBe('function');
  });

  it('shutdown closes runtime', async () => {
    // start then close
    await initOtel('projects-service');
    const shutdownSpy = (global as any).__otelRuntime.shutdown as ReturnType<typeof vi.fn>;

    await closeOtel();
    expect(shutdownSpy).toHaveBeenCalled();
  });
});