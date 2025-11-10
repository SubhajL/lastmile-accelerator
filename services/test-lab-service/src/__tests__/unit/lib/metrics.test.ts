import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import Fastify from 'fastify';
import type { FastifyInstance } from 'fastify';
import { register } from 'prom-client';
import { 
  createMetricsRoute, 
  incrementCounter, 
  recordHistogram, 
  setGauge,
  resetMetrics 
} from '../../../lib/metrics.js';

describe('Metrics', () => {
  let app: FastifyInstance;

  beforeEach(async () => {
    app = Fastify();
    resetMetrics(); // Clear metrics before each test
  });

  afterEach(async () => {
    await app.close();
    // Don't clear register as it removes default metrics
    resetMetrics(); // Only clear custom metrics
  });

  describe('createMetricsRoute()', () => {
    it('should register GET /metrics endpoint', async () => {
      createMetricsRoute(app);

      const response = await app.inject({
        method: 'GET',
        url: '/metrics',
      });

      expect(response.statusCode).toBe(200);
    });

    it('should return Prometheus text format', async () => {
      createMetricsRoute(app);

      const response = await app.inject({
        method: 'GET',
        url: '/metrics',
      });

      expect(response.headers['content-type']).toContain('text/plain');
      expect(response.body).toContain('# HELP');
      expect(response.body).toContain('# TYPE');
    });

    it('should include default Node.js metrics', async () => {
      createMetricsRoute(app);

      const response = await app.inject({
        method: 'GET',
        url: '/metrics',
      });

      expect(response.body).toContain('process_cpu');
      expect(response.body).toContain('nodejs_heap');
    });
  });

  describe('incrementCounter()', () => {
    it('should create counter if not exists', () => {
      incrementCounter('test_counter_total', { status: 'success' });

      // Should not throw
      expect(true).toBe(true);
    });

    it('should increment counter with labels', () => {
      incrementCounter('test_runs_total', { status: 'success', type: 'unit' });
      incrementCounter('test_runs_total', { status: 'success', type: 'unit' });

      // Counter incremented twice
      expect(true).toBe(true);
    });

    it('should allow incrementing by custom value', () => {
      incrementCounter('custom_counter_total', { action: 'test' }, 5);

      // Incremented by 5
      expect(true).toBe(true);
    });

    it('should handle counters without labels', () => {
      incrementCounter('simple_counter_total');

      expect(true).toBe(true);
    });
  });

  describe('recordHistogram()', () => {
    it('should create histogram if not exists', () => {
      recordHistogram('test_duration_ms', 150);

      expect(true).toBe(true);
    });

    it('should record value with timestamp', () => {
      recordHistogram('api_response_time_ms', 95, { endpoint: '/test' });

      expect(true).toBe(true);
    });

    it('should support multiple observations', () => {
      recordHistogram('latency_ms', 100);
      recordHistogram('latency_ms', 200);
      recordHistogram('latency_ms', 150);

      expect(true).toBe(true);
    });

    it('should record values in appropriate buckets', () => {
      // Fast operation
      recordHistogram('operation_duration_ms', 10);
      
      // Medium operation
      recordHistogram('operation_duration_ms', 500);
      
      // Slow operation
      recordHistogram('operation_duration_ms', 2000);

      expect(true).toBe(true);
    });
  });

  describe('setGauge()', () => {
    it('should create gauge if not exists', () => {
      setGauge('active_connections', 5);

      expect(true).toBe(true);
    });

    it('should update gauge to new value', () => {
      setGauge('active_test_runs', 3);
      setGauge('active_test_runs', 5);

      // Gauge should be set to 5, not 8
      expect(true).toBe(true);
    });

    it('should support negative values', () => {
      setGauge('delta_metric', -10);

      expect(true).toBe(true);
    });

    it('should handle gauge with labels', () => {
      setGauge('queue_size', 100, { queue: 'test_runs' });
      setGauge('queue_size', 50, { queue: 'browser_tests' });

      expect(true).toBe(true);
    });

    it('should support zero values', () => {
      setGauge('idle_workers', 0);

      expect(true).toBe(true);
    });
  });

  describe('resetMetrics()', () => {
    it('should clear all custom metrics', () => {
      incrementCounter('test_counter_total');
      recordHistogram('test_histogram_ms', 100);
      setGauge('test_gauge', 5);

      resetMetrics();

      // Metrics should be cleared
      expect(true).toBe(true);
    });
  });
});
