import { describe, it, expect } from 'vitest';
import request from 'supertest';
import { createServer } from '../../index';

describe('index bootstrap (smoke)', () => {
  it('createServer returns app serving /healthz', async () => {
    const app = await createServer({ serviceName: 'projects-service', autoInit: false });
    await app.ready();
    const res = await request(app.server).get('/healthz');
    expect(res.status).toBe(200);
    expect(res.body.status).toBe('ok');
  });
});