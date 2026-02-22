import { afterEach, describe, expect, test, vi } from 'vitest';
import Fastify from 'fastify';
import type { FastifyInstance } from 'fastify';
import { adminRetryRoute } from './admin.retry.js';
import { jwtPlugin } from '../auth/jwt.js';
import type { AppDeps } from '../http-routes.js';

function makeApp(secret = 'testsecret', retryNotification = async () => ({ success: true })) {
  const app = Fastify();
  app.register(jwtPlugin, { secret });
  const deps = {
    queue: { retryNotification },
    config: { service: { name: 'notification-service' } },
  } as AppDeps;

  app.register(
    async function (fastify) {
      adminRetryRoute(fastify as any, deps);
    },
    { prefix: '/admin' },
  );
  return { app, deps };
}

describe('POST /admin/retry', () => {
  let appToClose: FastifyInstance | undefined;

  afterEach(async () => {
    await appToClose?.close();
    appToClose = undefined;
  });

  test('should require admin role', async () => {
    const { app } = makeApp();
    appToClose = app;
    await app.ready();

    const res = await app.inject({
      method: 'POST',
      url: '/admin/retry/123',
    });

    expect(res.statusCode).toBe(401);
    expect(res.json()).toEqual({ error: 'unauthorized' });
  });

  test('should accept valid admin JWT and retry notification', async () => {
    const mockRetry = vi.fn().mockResolvedValue({ success: true });
    const { app } = makeApp('test-secret', mockRetry);
    appToClose = app;
    await app.ready();

    const token = app.jwt.sign({ roles: ['admin'] });
    const res = await app.inject({
      method: 'POST',
      url: '/admin/retry/test-notification-123',
      headers: { authorization: `Bearer ${token}` },
    });

    expect(res.statusCode).toBe(200);
    expect(res.json()).toEqual({ status: 'retried', notificationId: 'test-notification-123' });
    expect(mockRetry).toHaveBeenCalledWith('test-notification-123');
  });

  test('should return 400 when notificationId param is empty', async () => {
    const { app } = makeApp();
    appToClose = app;
    await app.ready();

    const token = app.jwt.sign({ roles: ['admin'] });
    const res = await app.inject({
      method: 'POST',
      url: '/admin/retry/',
      headers: { authorization: `Bearer ${token}` },
    });

    expect(res.statusCode).toBe(400);
  });

  test('should return 404 if notification not found', async () => {
    const mockRetry = vi.fn().mockResolvedValue({ success: false, error: 'not_found' });
    const { app } = makeApp('test-secret', mockRetry);
    appToClose = app;
    await app.ready();

    const token = app.jwt.sign({ roles: ['admin'] });
    const res = await app.inject({
      method: 'POST',
      url: '/admin/retry/non-existent',
      headers: { authorization: `Bearer ${token}` },
    });

    expect(res.statusCode).toBe(404);
    expect(res.json()).toEqual({ error: 'Notification not found' });
  });
});
