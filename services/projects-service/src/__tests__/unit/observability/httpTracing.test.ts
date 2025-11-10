import { describe, it, expect, vi, beforeEach } from 'vitest';
import Fastify from 'fastify';
import * as otApi from '@opentelemetry/api';
import { registerHttpTracing } from '../../../observability/httpTracing';
import { getActiveTraceparent, formatTraceparent } from '../../../observability/trace';

function fakeSpan() {
  return {
    attrs: {} as Record<string, any>,
    name: '',
    setAttribute(k: string, v: any) { this.attrs[k] = v; },
    updateName(n: string) { this.name = n; },
    spanContext() { return { traceId: '0123456789abcdef0123456789abcdef', spanId: '0123456789abcdef', traceFlags: 1 }; },
  };
}

describe('observability/httpTracing', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('enriches active span with route, ids and status', async () => {
    const span = fakeSpan();
    vi.spyOn(otApi.trace, 'getActiveSpan').mockReturnValue(span as any);

    const app = Fastify();
    registerHttpTracing(app, { serviceName: 'projects-service' });
    app.get('/v1/projects', async (_req, reply) => reply.code(200).send({ ok: true }));
    await app.ready();

    await app.inject({ method: 'GET', url: '/v1/projects', headers: { 'x-request-id': 'req-1', authorization: 'Bearer t' } });

    expect(span.name).toBe('GET /v1/projects');
    expect(span.attrs['http.route']).toBe('/v1/projects');
    expect(span.attrs['http.request_id']).toBe('req-1');
    expect(span.attrs['enduser.id']).toBeUndefined();
    expect(span.attrs['tenant.id']).toBeUndefined();
    expect(span.attrs['http.status_code']).toBe(200);
  });
});

describe('observability/trace utils', () => {
  it('formatTraceparent outputs valid W3C header', () => {
    const h = formatTraceparent({ traceId: '0123456789abcdef0123456789abcdef', spanId: '0123456789abcdef', traceFlags: 1 });
    expect(h).toBe('00-0123456789abcdef0123456789abcdef-0123456789abcdef-01');
  });

  it('getActiveTraceparent returns header from active span', () => {
    const span = fakeSpan();
    vi.spyOn(otApi.trace, 'getActiveSpan').mockReturnValue(span as any);
    const tp = getActiveTraceparent();
    expect(tp).toBe('00-0123456789abcdef0123456789abcdef-0123456789abcdef-01');
  });
});
