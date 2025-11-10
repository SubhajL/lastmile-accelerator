import { query, transaction } from '../db';
import { NotFoundError } from '../utils/errors';
import { publishProjectEvent } from '../events/eventPublisher';
import { randomUUID } from 'crypto';

export async function listProjects(tenantId: string) {
  const sql = `SELECT * FROM projects WHERE tenant_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`;
  const res = await query(sql, [tenantId]);
  return res.rows;
}

export async function getProject(tenantId: string, projectId: string) {
  const sql = `SELECT * FROM projects WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL`;
  const res = await query(sql, [projectId, tenantId]);
  if (res.rows.length === 0) throw new NotFoundError('Project not found');
  return res.rows[0];
}

export async function createProject(tenantId: string, input: { name: string; description?: string }, userId: string) {
  const projectId = randomUUID();
  const now = new Date().toISOString();

  const result = await transaction(async (client) => {
    await client.query(
      `INSERT INTO projects (id, tenant_id, name, description, created_at, updated_at)
       VALUES ($1, $2, $3, $4, NOW(), NOW())`,
      [projectId, tenantId, input.name, input.description ?? null]
    );

    // default environments
    const envs = ['dev', 'staging', 'prod'];
    for (const name of envs) {
      await client.query(
        `INSERT INTO environments (id, project_id, name, created_at, updated_at)
         VALUES ($1, $2, $3, NOW(), NOW())`,
        [randomUUID(), projectId, name]
      );
    }

    // add requester as owner member (if not exists)
    await client.query(
      `INSERT INTO members (id, tenant_id, user_id, email, role, created_at, updated_at)
       VALUES ($1, $2, $3, $4, 'owner', NOW(), NOW())
       ON CONFLICT DO NOTHING`,
      [randomUUID(), tenantId, userId, `${userId}@example.com`]
    );

    return { id: projectId, tenant_id: tenantId, name: input.name, description: input.description ?? null, created_at: now };
  });

  await publishProjectEvent('created', { projectId, tenantId, name: input.name }, undefined);
  return result;
}

export async function updateProject(tenantId: string, projectId: string, input: { name?: string; description?: string }) {
  // ensure project exists and belongs to tenant
  await getProject(tenantId, projectId);

  const fields = [] as string[];
  const params = [] as any[];
  let idx = 1;
  if (input.name) { fields.push(`name = $${idx++}`); params.push(input.name); }
  if (input.description !== undefined) { fields.push(`description = $${idx++}`); params.push(input.description); }
  params.push(projectId, tenantId);

  const sql = `UPDATE projects SET ${fields.join(', ')}, updated_at = NOW() WHERE id = $${idx++} AND tenant_id = $${idx} RETURNING *`;
  const res = await query(sql, params);
  const updated = res.rows[0];

  await publishProjectEvent('updated', { projectId, tenantId, changes: input }, undefined);
  return updated;
}

export async function deleteProject(tenantId: string, projectId: string) {
  // ensure exists
  await getProject(tenantId, projectId);

  await query(`UPDATE projects SET deleted_at = NOW() WHERE id = $1 AND tenant_id = $2`, [projectId, tenantId]);
  await publishProjectEvent('deleted', { projectId, tenantId }, undefined);
}