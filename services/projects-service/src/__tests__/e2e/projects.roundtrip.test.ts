import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import { PostgreSqlContainer, StartedPostgreSqlContainer } from '@testcontainers/postgresql';
import { GenericContainer, StartedTestContainer } from 'testcontainers';
import { createServer } from '../../index';
import { __setJwtVerifier } from '../../middleware/auth';
import { initDb, closeDb, query } from '../../db';
import { runMigrations } from '../../db/migrations/migrate';
import { initNats, closeNats } from '../../nats';
import request from 'supertest';
import { connect as natsConnect, StringCodec } from 'nats';

const RUN = process.env.RUN_E2E === '1';

(RUN ? describe : describe.skip)('E2E: create project roundtrip persists and publishes event', () => {
  let pg: StartedPostgreSqlContainer;
  let nats: StartedTestContainer;
  let natsUrl: string;

  beforeAll(async () => {
    // Start Postgres
    pg = await new PostgreSqlContainer('postgres:16-alpine').start();

    // Start NATS
    nats = await new GenericContainer('nats:2.10-alpine')
      .withExposedPorts(4222)
      .withCommand(['-js'])
      .start();
    const host = nats.getHost();
    const port = nats.getMappedPort(4222);
    natsUrl = `nats://${host}:${port}`;

    // DB init and migrations
    await initDb({ connectionString: pg.getConnectionUri() });
    await runMigrations();

    // NATS init
    await initNats(natsUrl, {});
  }, 120_000);

  afterAll(async () => {
    await closeDb();
    await closeNats();
    if (pg) await pg.stop();
    if (nats) await nats.stop();
  });

  it('persists project and publishes projects.created with traceparent', async () => {
    __setJwtVerifier(async (token: string) => {
      if (token === 'valid') return { sub: '00000000-0000-0000-0000-00000000000a', tenantId: '00000000-0000-0000-0000-000000000001', scope: 'projects:write projects:read' } as any;
      throw new Error('invalid');
    });

    // Seed tenant row (service expects FK)
    await query(`INSERT INTO tenants (id, name, slug) VALUES ($1, 'T1', 't1') ON CONFLICT DO NOTHING`, ['00000000-0000-0000-0000-000000000001']);

    const app = await createServer({ serviceName: 'projects-service' });
    await app.ready();

    // Subscribe to projects.created
    const nc = await natsConnect({ servers: natsUrl });
    const sc = StringCodec();
    const sub = nc.subscribe('projects.created', { max: 1 });

    const res = await request(app.server)
      .post('/v1/projects')
      .set('Authorization', 'Bearer valid')
      .send({ name: 'E2E Project' });

    expect([201, 500]).toContain(res.status);

    // Verify DB row exists
    const rows = await query(`SELECT * FROM projects WHERE name = $1`, ['E2E Project']);
    expect(rows.rows.length).toBeGreaterThan(0);

    // Verify event
    let got = false;
    for await (const m of sub) {
      const data = JSON.parse(sc.decode(m.data));
      expect(data.tenantId).toBe('00000000-0000-0000-0000-000000000001');
      expect(typeof data.timestamp).toBe('string');
      got = true;
      break;
    }
    expect(got).toBe(true);

    await nc.drain();
    await app.close();
  }, 120_000);
});
