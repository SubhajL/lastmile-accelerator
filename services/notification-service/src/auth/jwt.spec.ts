import { afterEach, describe, expect, test } from 'vitest';
import Fastify from 'fastify';

import { buildJwtAuth } from './jwt.js';

describe('buildJwtAuth', () => {
  const originalNodeEnv = process.env.NODE_ENV;
  const originalJwtSecret = process.env.JWT_SECRET;

  afterEach(() => {
    process.env.NODE_ENV = originalNodeEnv;
    if (originalJwtSecret === undefined) {
      delete process.env.JWT_SECRET;
    } else {
      process.env.JWT_SECRET = originalJwtSecret;
    }
  });

  test('throws in production when no secret/public key configured', () => {
    process.env.NODE_ENV = 'production';
    delete process.env.JWT_SECRET;

    const app = Fastify();
    expect(() => buildJwtAuth(app, {})).toThrow('JWT secret/public key is required in production');
  });

  test('does not throw outside production when no secret configured', () => {
    process.env.NODE_ENV = 'test';
    delete process.env.JWT_SECRET;

    const app = Fastify();
    expect(() => buildJwtAuth(app, {})).not.toThrow();
  });
});

