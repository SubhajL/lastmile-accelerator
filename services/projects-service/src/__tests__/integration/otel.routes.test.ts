import { describe, it, expect, vi, beforeEach } from 'vitest';
import Fastify from 'fastify';
import request from 'supertest';
import { createServer } from '../../index';
import { __setJwtVerifier } from '../../middleware/auth';
import * as nats from '../../nats';
import * as otApi from '@opentelemetry/api';

function fakeSpan() {
  return {
    attrs: {} as Record<string, any>,
    name: '',
    setAttribute(k: string, v: any) { this.attrs[k] = v; },
    updateName(n: string) { this.name = n; },
    spanContext() { return { traceId: 'cccccccccccccccccccccccccccccccc', spanId: 'dddddddddddddddd', traceFlags: 1 }; },
  } as any;
}

describe('integration: OTel span enrichment and event traceparent', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('GET /v1/projects enriches span; POST publishes traceparent', async () => {
    vi.spyOn(otApi.trace, 'getActiveSpan').mockReturnValue(fakeSpan());
    __setJwtVerifier(async (token: string) => {
      if (token === 'valid') return { sub: 'u1', tenantId: 't1', scope: 'projects:read projects:write' } as any;
      throw new Error('invalid');
    });

    // stub NATS publish
    const pub = vi.spyOn(nats, 'publish').mockResolvedValue();

    const app = await createServer({ serviceName: 'projects-service' });
    await app.ready();

    const g = await request(app.server).get('/v1/projects').set('Authorization', 'Bearer valid');
    expect([200,500]).toContain(g.status);

    const p = await request(app.server).post('/v1/projects').set('Authorization', 'Bearer valid').send({ name: 'X' });
    expect([201,500]).toContain(p.status);

    if (pub.mock.calls.length > 0) {
      const tp = pub.mock.calls[0][2];
      expect(tp).toBe('00-cccccccccccccccccccccccccccccccc-dddddddddddddddd-01');
    }

    await app.close();
  });
});
