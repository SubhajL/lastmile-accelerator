import { describe, expect, test, vi } from 'vitest';
import { registerHttpRoutes } from './http-routes.js';
import type { FastifyInstance } from 'fastify';
import type { AppDeps } from './http-routes.js';
import { createPrometheusRegistry } from './metrics/prom.js';

describe('registerHttpRoutes', () => {
  test('should mount healthz and metrics endpoints', () => {
    const mockApp = {
      get: vi.fn(),
      post: vi.fn(),
      decorate: vi.fn(),
	    register: vi.fn(),
	  } as unknown as FastifyInstance;

	  const mockDeps = {
	    queue: { retryNotification: async () => ({ success: true }) },
	    enqueue: async () => 'job-1',
	    config: { service: { name: 'notification-service' } },
	  } as AppDeps;

    const registry = createPrometheusRegistry();

    registerHttpRoutes(mockApp, mockDeps, registry);

    // Should register healthz endpoint
    expect(mockApp.get).toHaveBeenCalledWith('/healthz', expect.any(Function));

    // Should register metrics endpoint
    expect(mockApp.get).toHaveBeenCalledWith('/metrics', expect.any(Function));

    // Should decorate app with metrics
    expect(mockApp.decorate).toHaveBeenCalledWith('metrics', registry);
  });

  test('should register admin routes', () => {
    const mockApp = {
      get: vi.fn(),
      post: vi.fn(),
      decorate: vi.fn(),
      register: vi.fn(),
    } as unknown as FastifyInstance;

	  const mockDeps = {
	    queue: { retryNotification: async () => ({ success: true }) },
	    enqueue: async () => 'job-1',
	    config: { service: { name: 'notification-service' } },
	  } as AppDeps;

    const registry = createPrometheusRegistry();

    registerHttpRoutes(mockApp, mockDeps, registry);

    // Should register admin routes through register plugin
    expect(mockApp.register).toHaveBeenCalledWith(expect.any(Function), { prefix: '/admin' });
  });
});
