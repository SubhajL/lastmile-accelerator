import { describe, it, expect, beforeEach } from 'vitest';
import Fastify from 'fastify';
import request from 'supertest';
import { registerHealthRoutes } from '../../routes/health';
import { metricsMiddleware, resetMetrics } from '../../middleware/metrics';

describe('routes/health.ts', () => {
  let app: any;

  beforeEach(async () => {
    resetMetrics();
    app = Fastify();
    app.addHook('onRequest', metricsMiddleware());
    registerHealthRoutes(app, { serviceName: 'projects-service' });
    await app.ready();
  });

  it('GET /healthz returns ok', async () => {
    const res = await request(app.server).get('/healthz');
    expect(res.status).toBe(200);
    expect(res.body.status).toBe('ok');
    expect(res.body.service).toBe('projects-service');
    expect(typeof res.body.timestamp).toBe('string');
  });

  it('GET /metrics exposes prometheus text with service label', async () => {
    // trigger a request to increment counters
    await request(app.server).get('/healthz');

    const res = await request(app.server).get('/metrics');
    expect(res.status).toBe(200);
    expect(res.headers['content-type']).toContain('text/plain');

    const body = res.text;
    expect(body).toContain('request_total');
    expect(body).toContain('service_name="projects-service"');
  });
});