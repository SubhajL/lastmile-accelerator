import { query } from '../db';
import { NotFoundError } from '../utils/errors';
import { publishEnvironmentEvent } from '../events/eventPublisher';

async function assertProjectTenant(tenantId: string, projectId: string) {
  const res = await query(`SELECT id FROM projects WHERE id = $1 AND tenant_id = $2`, [projectId, tenantId]);
  if (res.rows.length === 0) throw new NotFoundError('Project not found');
}

export async function listEnvironments(tenantId: string, projectId: string) {
  await assertProjectTenant(tenantId, projectId);
  const res = await query(`SELECT * FROM environments WHERE project_id = $1 ORDER BY name ASC`, [projectId]);
  return res.rows;
}

export async function getEnvironment(tenantId: string, projectId: string, envId: string) {
  await assertProjectTenant(tenantId, projectId);
  const res = await query(`SELECT * FROM environments WHERE id = $1 AND project_id = $2`, [envId, projectId]);
  if (res.rows.length === 0) throw new NotFoundError('Environment not found');
  return res.rows[0];
}

export async function createEnvironment(tenantId: string, projectId: string, input: { name: string; config?: Record<string, any> }) {
  await assertProjectTenant(tenantId, projectId);
  const res = await query(`INSERT INTO environments (project_id, name, config, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW()) RETURNING *`, [projectId, input.name, input.config ?? {}]);
  const row = res.rows[0];
  await publishEnvironmentEvent('created', { tenantId, projectId, environmentId: row.id, name: input.name }, undefined);
  return row;
}

export async function updateEnvironment(tenantId: string, projectId: string, envId: string, input: { config?: Record<string, any> }) {
  await assertProjectTenant(tenantId, projectId);
  const existing = await query(`SELECT * FROM environments WHERE id = $1 AND project_id = $2`, [envId, projectId]);
  if (existing.rows.length === 0) throw new NotFoundError('Environment not found');
  const res = await query(`UPDATE environments SET config = $1, updated_at = NOW() WHERE id = $2 AND project_id = $3 RETURNING *`, [input.config ?? existing.rows[0].config ?? {}, envId, projectId]);
  const row = res.rows[0];
  await publishEnvironmentEvent('updated', { tenantId, projectId, environmentId: envId }, undefined);
  return row;
}

export async function deleteEnvironment(tenantId: string, projectId: string, envId: string) {
  await assertProjectTenant(tenantId, projectId);
  await (await import('../db')).transaction(async (client: any) => {
    const cnt = await client.query(`SELECT COUNT(1)::text as cnt FROM environments WHERE project_id = $1`, [projectId]);
    const n = parseInt(cnt.rows[0]?.cnt ?? '0', 10);
    if (n <= 1) throw new Error('Cannot delete last environment');
    await client.query(`DELETE FROM environments WHERE id = $1 AND project_id = $2`, [envId, projectId]);
  });
  await publishEnvironmentEvent('deleted', { tenantId, projectId, environmentId: envId }, undefined);
}

export async function setIngestionModes(tenantId: string, projectId: string, input: { modes: Array<'A'|'B'|'C'>; defaultMode: 'A'|'B'|'C' }) {
  await assertProjectTenant(tenantId, projectId);
  await (await import('../db')).transaction(async (client: any) => {
    await client.query(`DELETE FROM ingestion_modes WHERE project_id = $1`, [projectId]);
    for (const mode of input.modes) {
      await client.query(`INSERT INTO ingestion_modes (id, project_id, mode, is_default, created_at) VALUES (gen_random_uuid(), $1, $2, $3, NOW())`, [projectId, mode, mode === input.defaultMode]);
    }
  });
  await publishEnvironmentEvent('updated', { tenantId, projectId, ingestionModes: input.modes, defaultMode: input.defaultMode }, undefined);
}
