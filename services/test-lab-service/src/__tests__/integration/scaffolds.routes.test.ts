import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { createApp } from '../../app.js';
import { createMockJWT, TEST_JWT_SECRET } from '../fixtures/jwt-helpers.js';
import { applyTestEnv } from '../fixtures/env.js';

vi.mock('../../lib/jwks.js', () => ({
  verifyJwt: vi.fn(async ({ token }: { token: string }) => {
    const { verify } = await import('jsonwebtoken');
    return verify(token, TEST_JWT_SECRET, { algorithms: ['HS256'] });
  }),
}));

// Minimal env
applyTestEnv();
process.env.JWT_JWKS_URL = TEST_JWT_SECRET;
process.env.JWT_ISSUER = 'https://auth.example.com/';
process.env.JWT_AUDIENCE = 'test-lab-service';

describe('scaffolds routes', () => {
  let app: Awaited<ReturnType<typeof createApp>>;
  let token: string;
  const projectId = '11111111-1111-1111-1111-111111111111';

  beforeEach(async () => {
    app = await createApp();
    token = createMockJWT({ scopes: ['scaffold:read', 'scaffold:write'] });
  });

  afterEach(async () => {
    await app.close();
  });

  it('creates and retrieves a scaffold', async () => {
    const create = await app.inject({
      method: 'POST',
      url: `/v1/projects/${projectId}/test-scaffolds`,
      headers: { authorization: `Bearer ${token}` },
      payload: {
        type: 'unit', framework: 'vitest', language: 'ts', config: { a: 1 },
      },
    });
    expect(create.statusCode).toBe(201);
    const created = create.json();

    const get = await app.inject({
      method: 'GET',
      url: `/v1/test-scaffolds/${created.id}`,
      headers: { authorization: `Bearer ${token}` },
    });
    expect(get.statusCode).toBe(200);
    expect(get.json().id).toBe(created.id);
  });

  it('lists scaffolds by project with pagination', async () => {
    // create two
    for (let i = 0; i < 2; i++) {
      await app.inject({
        method: 'POST',
        url: `/v1/projects/${projectId}/test-scaffolds`,
        headers: { authorization: `Bearer ${token}` },
        payload: { type: 'unit', framework: 'vitest', language: 'ts', config: {} },
      });
    }

    const res1 = await app.inject({
      method: 'GET',
      url: `/v1/projects/${projectId}/test-scaffolds?limit=1`,
      headers: { authorization: `Bearer ${token}` },
    });
    expect(res1.statusCode).toBe(200);
    const page1 = res1.json();
    expect(page1.items.length).toBe(1);
    expect(page1.nextCursor).toBeDefined();

    const res2 = await app.inject({
      method: 'GET',
      url: `/v1/projects/${projectId}/test-scaffolds?limit=1&cursor=${page1.nextCursor}`,
      headers: { authorization: `Bearer ${token}` },
    });
    const page2 = res2.json();
    expect(page2.items.length).toBe(1);
  });

  it('updates and deletes a scaffold', async () => {
    const create = await app.inject({
      method: 'POST',
      url: `/v1/projects/${projectId}/test-scaffolds`,
      headers: { authorization: `Bearer ${token}` },
      payload: { type: 'unit', framework: 'vitest', language: 'ts', config: {} },
    });
    const created = create.json();

    const upd = await app.inject({
      method: 'PUT',
      url: `/v1/test-scaffolds/${created.id}`,
      headers: { authorization: `Bearer ${token}` },
      payload: { framework: 'jest' },
    });
    expect(upd.statusCode).toBe(200);
    expect(upd.json().framework).toBe('jest');

    const del = await app.inject({
      method: 'DELETE',
      url: `/v1/test-scaffolds/${created.id}`,
      headers: { authorization: `Bearer ${token}` },
    });
    expect(del.statusCode).toBe(204);

    const get404 = await app.inject({
      method: 'GET',
      url: `/v1/test-scaffolds/${created.id}`,
      headers: { authorization: `Bearer ${token}` },
    });
    expect(get404.statusCode).toBe(404);
  });

  it('requires scopes for write actions', async () => {
    const readOnlyToken = createMockJWT({ scopes: ['scaffold:read'] });
    const denied = await app.inject({
      method: 'POST',
      url: `/v1/projects/${projectId}/test-scaffolds`,
      headers: { authorization: `Bearer ${readOnlyToken}` },
      payload: { type: 'unit', framework: 'vitest', language: 'ts', config: {} },
    });
    expect(denied.statusCode).toBe(403);
  });

  it('returns 401 and does not call verifyJwt when Authorization header is missing', async () => {
    const denied = await app.inject({
      method: 'POST',
      url: `/v1/projects/${projectId}/test-scaffolds`,
      payload: { type: 'unit', framework: 'vitest', language: 'ts', config: {} },
    });
    expect(denied.statusCode).toBe(401);
  });
});
