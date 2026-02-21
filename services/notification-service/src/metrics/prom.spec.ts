import { describe, expect, test, vi } from 'vitest';
import {
  createPrometheusRegistry,
  createNotificationMetrics,
  registerMetricsEndpoint,
} from './prom.js';
import type { Registry } from 'prom-client';
import type { FastifyInstance } from 'fastify';

describe('createPrometheusRegistry', () => {
  test('should create a registry with default Node.js metrics', async () => {
    const registry = createPrometheusRegistry();

    expect(registry).toBeDefined();

    // Registry should have metrics after creation
    const metrics = await registry.metrics();
    expect(metrics).toContain('nodejs_');
    expect(metrics).toContain('process_cpu');
  });
});

describe('createNotificationMetrics', () => {
  test('should register notification metrics', async () => {
    const registry = createPrometheusRegistry();
    const metrics = createNotificationMetrics(registry);

    expect(metrics.notificationEnqueued).toBeDefined();
    expect(metrics.notificationSent).toBeDefined();
    expect(metrics.notificationDuration).toBeDefined();

    // Test counter increments
    metrics.notificationEnqueued.inc({ channel: 'email', type: 'welcome' });
    metrics.notificationSent.inc({ channel: 'email', status: 'success' });
    metrics.notificationDuration.observe({ channel: 'email' }, 0.5);

    const output = await registry.metrics();
    expect(output).toContain('notification_enqueued_total');
    expect(output).toContain('notification_sent_total');
    expect(output).toContain('notification_send_duration_seconds');
  });

  test('should include proper labels', async () => {
    const registry = createPrometheusRegistry();
    const metrics = createNotificationMetrics(registry);

    // Verify labels work correctly
    metrics.notificationEnqueued.inc({ channel: 'slack', type: 'alert' });
    metrics.notificationSent.inc({ channel: 'sms', status: 'failed' });

    const output = await registry.metrics();
    expect(output).toContain('channel="slack"');
    expect(output).toContain('channel="sms"');
    expect(output).toContain('type="alert"');
    expect(output).toContain('status="failed"');
  });
});

describe('registerMetricsEndpoint', () => {
  test('should register GET /metrics endpoint', async () => {
    const registry = createPrometheusRegistry();
    const mockApp = {
      get: vi.fn(),
      decorate: vi.fn(),
    } as unknown as FastifyInstance;

    registerMetricsEndpoint(mockApp, registry);

    // Should decorate app with metrics
    expect(mockApp.decorate).toHaveBeenCalledWith('metrics', registry);

    // Should register GET /metrics
    expect(mockApp.get).toHaveBeenCalledWith('/metrics', expect.any(Function));
  });

  test('should return metrics with correct content type', async () => {
    const registry = createPrometheusRegistry();
    createNotificationMetrics(registry);

    const mockReply = {
      header: vi.fn().mockReturnThis(),
      send: vi.fn(),
    };

    const mockApp = {
      get: vi.fn(),
      decorate: vi.fn(),
    } as unknown as FastifyInstance;

    registerMetricsEndpoint(mockApp, registry);

    // Get the handler that was registered
    const handler = mockApp.get.mock.calls[0][1];
    await handler({}, mockReply);

    expect(mockReply.header).toHaveBeenCalledWith(
      'Content-Type',
      'text/plain; version=0.0.4; charset=utf-8',
    );
    expect(mockReply.send).toHaveBeenCalledWith(expect.stringContaining('# HELP'));
  });
});
