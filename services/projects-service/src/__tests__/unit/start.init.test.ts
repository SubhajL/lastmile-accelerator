import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';

// Mock Fastify to avoid real network listen
vi.mock('fastify', () => {
  return {
    default: () => ({
      addHook: vi.fn(),
      setErrorHandler: vi.fn(),
      register: vi.fn().mockResolvedValue(undefined),
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      delete: vi.fn(),
      listen: vi.fn().mockResolvedValue(undefined),
      ready: vi.fn().mockResolvedValue(undefined),
      server: {},
    }),
  };
});

import * as db from '../../db';
import * as migrate from '../../db/migrations/migrate';
import * as nats from '../../nats';
import * as otel from '../../otel';
import { start } from '../../index';

const OLD_ENV = { ...process.env };

describe('start() initialization wiring', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    process.env = { ...OLD_ENV, NODE_ENV: 'test' };
  });
  afterEach(() => {
    process.env = { ...OLD_ENV };
  });

  it('does not initialize DB/NATS/OTel by default in test env', async () => {
    const dbSpy = vi.spyOn(db, 'initDb').mockResolvedValue({} as any);
    const migSpy = vi.spyOn(migrate, 'runMigrations').mockResolvedValue();
    const natsSpy = vi.spyOn(nats, 'initNats').mockResolvedValue({} as any);
    const otelSpy = vi.spyOn(otel, 'initOtel').mockResolvedValue();

    await start();

    expect(dbSpy).not.toHaveBeenCalled();
    expect(migSpy).not.toHaveBeenCalled();
    expect(natsSpy).not.toHaveBeenCalled();
    expect(otelSpy).not.toHaveBeenCalled();
  });

  it('initializes DB (with migrations), NATS, and OTel when flags enabled', async () => {
    process.env.INIT_DB = 'true';
    process.env.RUN_MIGRATIONS = 'true';
    process.env.INIT_NATS = 'true';
    process.env.INIT_OTEL = 'true';

    process.env.DATABASE_URL = 'postgres://localhost:5432/appdb'; // gitleaks:allow test fixture
    process.env.NATS_URL = 'nats://localhost:4222';
    process.env.OTEL_EXPORTER_OTLP_ENDPOINT = 'http://collector:4318';
    process.env.SERVICE_PORT = '0';
    process.env.SERVICE_HOST = '127.0.0.1';
    process.env.SERVICE_NAME = 'projects-service';

    const dbSpy = vi.spyOn(db, 'initDb').mockResolvedValue({} as any);
    const migSpy = vi.spyOn(migrate, 'runMigrations').mockResolvedValue();
    const natsSpy = vi.spyOn(nats, 'initNats').mockResolvedValue({} as any);
    const otelSpy = vi.spyOn(otel, 'initOtel').mockResolvedValue();

    await start();

    expect(dbSpy).toHaveBeenCalledWith({ connectionString: 'postgres://user:pass@localhost:5432/appdb' });
    expect(migSpy).toHaveBeenCalled();
    expect(natsSpy).toHaveBeenCalledWith('nats://localhost:4222', expect.any(Object));
    expect(otelSpy).toHaveBeenCalledWith('projects-service', { exporterUrl: 'http://collector:4318' });
  });
});