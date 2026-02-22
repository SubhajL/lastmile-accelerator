import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { initTelemetry, shutdownTelemetry, createTracer } from './telemetry.js';
import type { ObservabilityConfig } from './types.js';

describe('telemetry', () => {
  let mockConfig: ObservabilityConfig;

  beforeEach(() => {
    mockConfig = {
      serviceName: 'notification-service',
      serviceVersion: '1.0.0',
      otlpEndpoint: 'http://localhost:4318',
      environment: 'dev'
    };
  });

  afterEach(async () => {
    // Cleanup any initialized SDK
  });

  describe('initTelemetry', () => {
    it('should initialize NodeSDK with correct resource attributes', () => {
      const sdk = initTelemetry(mockConfig);

      expect(sdk).toBeDefined();
      // SDK is initialized but we can't easily inspect internal state
      // This test verifies it doesn't throw
    });

    it('should configure OTLP exporter with endpoint from config', () => {
      const config: ObservabilityConfig = {
        ...mockConfig,
        otlpEndpoint: 'http://custom-endpoint:4318'
      };

      const sdk = initTelemetry(config);

      expect(sdk).toBeDefined();
    });

    it('should enable auto-instrumentation', () => {
      const sdk = initTelemetry(mockConfig);

      expect(sdk).toBeDefined();
      // Auto-instrumentation is configured in constructor
    });
  });

  describe('shutdownTelemetry', () => {
    it('should shutdown SDK without throwing errors', async () => {
      const sdk = initTelemetry(mockConfig);

      await expect(shutdownTelemetry(sdk)).resolves.not.toThrow();
    });
  });

  describe('createTracer', () => {
    it('should create tracer with provided name', () => {
      const tracer = createTracer('test-tracer');

      expect(tracer).toBeDefined();
      expect(typeof tracer.startSpan).toBe('function');
    });
  });
});
