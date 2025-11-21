/**
 * Integration tests for service transaction behavior.
 *
 * These tests verify real database transaction commit/rollback behavior using
 * Testcontainers with PostgreSQL. They replace the previous transaction spy
 * unit tests which were brittle due to nested transaction call complexity.
 *
 * **REQUIREMENTS:**
 * - Docker must be running (Testcontainers requires Docker runtime)
 * - Tests will fail with "Could not find a working container runtime strategy" if Docker unavailable
 * - Alternative: Use dev stack PostgreSQL if Testcontainers unavailable (requires manual setup)
 *
 * **Coverage:**
 * - Transaction commit on successful operations
 * - Transaction rollback on business rule violations
 * - Transaction rollback on mid-operation failures
 * - Concurrent operation safety (transaction isolation)
 * - Edge case validation (invalid inputs, duplicates)
 * - Event publishing atomicity
 */
import { afterAll, afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { PostgreSqlContainer, type StartedPostgreSqlContainer } from '@testcontainers/postgresql';
import { randomUUID } from 'node:crypto';
import * as database from '../../db';
import { runMigrations } from '../../db/migrations/migrate';
import * as environmentService from '../../services/environmentService';
import * as memberService from '../../services/memberService';
import * as events from '../../events/eventPublisher';

describe('service transactions against Postgres', () => {
  let pg: StartedPostgreSqlContainer;
  let publishEnvironmentEventSpy: any;
  let publishMemberEventSpy: any;

  beforeAll(async () => {
    pg = await new PostgreSqlContainer('postgres:16-alpine').start();
    await database.initDb({ connectionString: pg.getConnectionUri() });
    await runMigrations();
  }, 120_000);

  afterAll(async () => {
    await database.closeDb();
    if (pg) await pg.stop();
  });

  beforeEach(async () => {
    publishEnvironmentEventSpy = vi.spyOn(events, 'publishEnvironmentEvent').mockResolvedValue();
    publishMemberEventSpy = vi.spyOn(events, 'publishMemberEvent').mockResolvedValue();
    await database.query(
      'TRUNCATE ingestion_modes, environments, members, projects, tenants RESTART IDENTITY CASCADE',
    );
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('removeMember deletes member when multiple owners exist', async () => {
    const tenantId = randomUUID();
    const ownerKeepId = randomUUID();
    const ownerRemoveId = randomUUID();

    await seedTenant(tenantId);
    await seedMember(tenantId, ownerKeepId, 'owner', 'keep@example.com');
    await seedMember(tenantId, ownerRemoveId, 'owner', 'remove@example.com');

    await memberService.removeMember(tenantId, ownerRemoveId, 'owner');

    const remaining = await database.query(
      'SELECT id, role FROM members WHERE tenant_id = $1 ORDER BY id',
      [tenantId],
    );
    expect(remaining.rows).toEqual([{ id: ownerKeepId, role: 'owner' }]);
    expect(publishMemberEventSpy).toHaveBeenCalledTimes(1);
    expect(publishMemberEventSpy).toHaveBeenCalledWith(
      'removed',
      expect.objectContaining({ memberId: ownerRemoveId, tenantId }),
      undefined,
    );
  });

  it('removeMember blocks deleting last owner and keeps row', async () => {
    const tenantId = randomUUID();
    const ownerId = randomUUID();

    await seedTenant(tenantId);
    await seedMember(tenantId, ownerId, 'owner', 'solo-owner@example.com');

    await expect(memberService.removeMember(tenantId, ownerId, 'owner')).rejects.toThrow(
      'Cannot remove last owner',
    );

    const remaining = await database.query('SELECT id FROM members WHERE tenant_id = $1', [
      tenantId,
    ]);
    expect(remaining.rows.map((r) => r.id)).toEqual([ownerId]);
    expect(publishMemberEventSpy).not.toHaveBeenCalled();
  });

  it('deleteEnvironment removes target when multiple environments exist', async () => {
    const tenantId = randomUUID();
    const projectId = randomUUID();
    const devEnvId = randomUUID();
    const stagingEnvId = randomUUID();

    await seedTenant(tenantId);
    await seedProject(tenantId, projectId);
    await seedEnvironment(projectId, devEnvId, 'dev');
    await seedEnvironment(projectId, stagingEnvId, 'staging');

    await environmentService.deleteEnvironment(tenantId, projectId, stagingEnvId);

    const remaining = await database.query(
      'SELECT id, name FROM environments WHERE project_id = $1 ORDER BY name',
      [projectId],
    );
    expect(remaining.rows).toEqual([{ id: devEnvId, name: 'dev' }]);
    expect(publishEnvironmentEventSpy).toHaveBeenCalledTimes(1);
    expect(publishEnvironmentEventSpy).toHaveBeenCalledWith(
      'deleted',
      expect.objectContaining({ environmentId: stagingEnvId, projectId, tenantId }),
      undefined,
    );
  });

  it('deleteEnvironment rejects deleting last remaining environment', async () => {
    const tenantId = randomUUID();
    const projectId = randomUUID();
    const envId = randomUUID();

    await seedTenant(tenantId);
    await seedProject(tenantId, projectId);
    await seedEnvironment(projectId, envId, 'prod');

    await expect(environmentService.deleteEnvironment(tenantId, projectId, envId)).rejects.toThrow(
      'Cannot delete last environment',
    );

    const remaining = await database.query('SELECT id FROM environments WHERE project_id = $1', [
      projectId,
    ]);
    expect(remaining.rows.map((r) => r.id)).toEqual([envId]);
    expect(publishEnvironmentEventSpy).not.toHaveBeenCalled();
  });

  it('setIngestionModes rewrites modes in a single transaction', async () => {
    const tenantId = randomUUID();
    const projectId = randomUUID();

    await seedTenant(tenantId);
    await seedProject(tenantId, projectId);
    await seedIngestionMode(projectId, 'A', false);
    await seedIngestionMode(projectId, 'B', true);

    await environmentService.setIngestionModes(tenantId, projectId, {
      modes: ['B', 'C'],
      defaultMode: 'C',
    });

    const modes = await database.query(
      'SELECT mode, is_default FROM ingestion_modes WHERE project_id = $1 ORDER BY mode',
      [projectId],
    );
    expect(modes.rows).toEqual([
      { mode: 'B', is_default: false },
      { mode: 'C', is_default: true },
    ]);
    expect(publishEnvironmentEventSpy).toHaveBeenCalledTimes(1);
    expect(publishEnvironmentEventSpy).toHaveBeenCalledWith(
      'updated',
      expect.objectContaining({
        ingestionModes: ['B', 'C'],
        defaultMode: 'C',
        projectId,
        tenantId,
      }),
      undefined,
    );
  });

  it('setIngestionModes rolls back when insert fails mid-transaction', async () => {
    const tenantId = randomUUID();
    const projectId = randomUUID();

    await seedTenant(tenantId);
    await seedProject(tenantId, projectId);
    await seedIngestionMode(projectId, 'A', true);
    await seedIngestionMode(projectId, 'B', false);

    const realTransaction = database.transaction;
    vi.spyOn(database, 'transaction').mockImplementation(async (fn: any) =>
      realTransaction(async (client: any) => {
        let insertCount = 0;
        const originalQuery = client.query.bind(client);
        client.query = async (sql: string, params?: any[]) => {
          if (/INSERT INTO ingestion_modes/.test(sql)) {
            insertCount += 1;
            if (insertCount === 2) {
              await originalQuery(sql, params as any);
              throw new Error('insertion failure');
            }
          }
          return originalQuery(sql, params as any);
        };
        return fn(client);
      }),
    );

    await expect(
      environmentService.setIngestionModes(tenantId, projectId, {
        modes: ['A', 'B', 'C'],
        defaultMode: 'B',
      }),
    ).rejects.toThrow('insertion failure');

    const modes = await database.query(
      'SELECT mode, is_default FROM ingestion_modes WHERE project_id = $1 ORDER BY mode',
      [projectId],
    );
    expect(modes.rows).toEqual([
      { mode: 'A', is_default: true },
      { mode: 'B', is_default: false },
    ]);
    expect(publishEnvironmentEventSpy).not.toHaveBeenCalled();
  });

  it('setIngestionModes rejects when defaultMode not in modes list', async () => {
    const tenantId = randomUUID();
    const projectId = randomUUID();

    await seedTenant(tenantId);
    await seedProject(tenantId, projectId);
    await seedIngestionMode(projectId, 'A', true);

    await expect(
      environmentService.setIngestionModes(tenantId, projectId, {
        modes: ['A', 'B'],
        defaultMode: 'C',
      }),
    ).rejects.toThrow();

    const modes = await database.query(
      'SELECT mode, is_default FROM ingestion_modes WHERE project_id = $1 ORDER BY mode',
      [projectId],
    );
    expect(modes.rows).toEqual([{ mode: 'A', is_default: true }]);
    expect(publishEnvironmentEventSpy).not.toHaveBeenCalled();
  });

  it('setIngestionModes handles duplicate modes gracefully', async () => {
    const tenantId = randomUUID();
    const projectId = randomUUID();

    await seedTenant(tenantId);
    await seedProject(tenantId, projectId);
    await seedIngestionMode(projectId, 'A', true);

    await expect(
      environmentService.setIngestionModes(tenantId, projectId, {
        modes: ['A', 'A', 'B'],
        defaultMode: 'A',
      }),
    ).rejects.toThrow();

    const modes = await database.query(
      'SELECT mode FROM ingestion_modes WHERE project_id = $1 ORDER BY mode',
      [projectId],
    );
    expect(modes.rows).toEqual([{ mode: 'A' }]);
    expect(publishEnvironmentEventSpy).not.toHaveBeenCalled();
  });

  it('handles concurrent member additions safely', async () => {
    const tenantId = randomUUID();
    const member1Id = randomUUID();
    const member2Id = randomUUID();

    await seedTenant(tenantId);

    await Promise.all([
      memberService.addMember(
        tenantId,
        {
          email: 'member1@example.com',
          role: 'admin',
        },
        'owner',
      ),
      memberService.addMember(
        tenantId,
        {
          email: 'member2@example.com',
          role: 'developer',
        },
        'owner',
      ),
    ]);

    const members = await database.query(
      'SELECT email, role FROM members WHERE tenant_id = $1 ORDER BY role',
      [tenantId],
    );
    expect(members.rows).toHaveLength(2);
    expect(members.rows).toEqual([
      { email: 'member1@example.com', role: 'admin' },
      { email: 'member2@example.com', role: 'developer' },
    ]);
    expect(publishMemberEventSpy).toHaveBeenCalledTimes(2);
  });

  it('handles concurrent environment creations safely', async () => {
    const tenantId = randomUUID();
    const projectId = randomUUID();

    await seedTenant(tenantId);
    await seedProject(tenantId, projectId);

    await Promise.all([
      environmentService.createEnvironment(tenantId, projectId, {
        name: 'dev',
        config: { region: 'us-east-1' },
      }),
      environmentService.createEnvironment(tenantId, projectId, {
        name: 'staging',
        config: { region: 'us-west-2' },
      }),
    ]);

    const environments = await database.query(
      'SELECT name FROM environments WHERE project_id = $1 ORDER BY name',
      [projectId],
    );
    expect(environments.rows).toHaveLength(2);
    expect(environments.rows).toEqual([{ name: 'dev' }, { name: 'staging' }]);
    expect(publishEnvironmentEventSpy).toHaveBeenCalledTimes(2);
  });
});

async function seedTenant(tenantId: string) {
  await database.query('INSERT INTO tenants (id, name, slug) VALUES ($1, $2, $3)', [
    tenantId,
    `Tenant ${tenantId.slice(0, 8)}`,
    `slug-${tenantId.slice(0, 8)}`,
  ]);
}

async function seedProject(tenantId: string, projectId: string) {
  await database.query('INSERT INTO projects (id, tenant_id, name) VALUES ($1, $2, $3)', [
    projectId,
    tenantId,
    `Project ${projectId.slice(0, 8)}`,
  ]);
}

async function seedEnvironment(projectId: string, environmentId: string, name: string) {
  await database.query(
    'INSERT INTO environments (id, project_id, name, config) VALUES ($1, $2, $3, $4)',
    [environmentId, projectId, name, {}],
  );
}

async function seedMember(tenantId: string, memberId: string, role: string, email: string) {
  await database.query(
    'INSERT INTO members (id, tenant_id, user_id, email, role) VALUES ($1, $2, $3, $4, $5)',
    [memberId, tenantId, randomUUID(), email, role],
  );
}

async function seedIngestionMode(projectId: string, mode: 'A' | 'B' | 'C', isDefault: boolean) {
  await database.query(
    'INSERT INTO ingestion_modes (id, project_id, mode, is_default, created_at) VALUES ($1, $2, $3, $4, NOW())',
    [randomUUID(), projectId, mode, isDefault],
  );
}
