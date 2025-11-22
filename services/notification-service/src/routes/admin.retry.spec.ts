import { describe, expect, test, vi, beforeAll } from 'vitest';
import Fastify from 'fastify';
import { adminRetryRoute } from './admin.retry.js';
import { jwtPlugin } from '../auth/jwt.js';
import type { AppDeps } from '../types.js';

function makeApp(secret = 'testsecret', retryNotification = async () => ({ success: true })) {
  const app = Fastify();
  app.register(jwtPlugin, { secret });
  const deps = {
    queue: { retryNotification },
    config: { service: { name: 'notification-service' } },
  } as AppDeps;
  adminRetryRoute(app as any, deps);
  return { app, deps };
}

describe('POST /admin/retry', () => {
  test('should require admin role', async () => {
    const { app } = makeApp();
    await app.ready();

    const res = await app.inject({
      method: 'POST',
      url: '/retry/123',
    });

    expect(res.statusCode).toBe(401);
    expect(res.json()).toEqual({ error: 'unauthorized' });
  });

  test('should accept valid admin JWT and retry notification', async () => {
    const mockRetry = vi.fn().mockResolvedValue({ success: true });
    const { app } = makeApp('test-secret', mockRetry);
    await app.ready();

    const token = app.jwt.sign({ roles: ['admin'] });
    const res = await app.inject({
      method: 'POST',
      url: '/retry/test-notification-123',
      headers: { authorization: `Bearer ${token}` },
    });

    expect(res.statusCode).toBe(200);
    expect(res.json()).toEqual({ status: 'retried', notificationId: 'test-notification-123' });
    expect(mockRetry).toHaveBeenCalledWith('test-notification-123');
  });

  test('should return 400 if notificationId is empty', async () => {
    const { app } = makeApp();
    await app.ready();

    const token = app.jwt.sign({ roles: ['admin'] });
    const res = await app.inject({
      method: 'POST',
      url: '/retry/',
      headers: { authorization: `Bearer ${token}` },
    });

    expect(res.statusCode).toBe(400); // Empty notificationId is treated as empty string
  });

  test('should return 404 if notification not found', async () => {
    const mockRetry = vi.fn().mockResolvedValue({ success: false, error: 'not_found' });
    const { app } = makeApp('test-secret', mockRetry);
    await app.ready();

    const token = app.jwt.sign({ roles: ['admin'] });
    const res = await app.inject({
      method: 'POST',
      url: '/retry/non-existent',
      headers: { authorization: `Bearer ${token}` },
    });

    expect(res.statusCode).toBe(404);
    expect(res.json()).toEqual({ error: 'Notification not found' });
  });
});
