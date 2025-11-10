import { query } from '../db';
import { AuthError, NotFoundError } from '../utils/errors';
import { publishMemberEvent } from '../events/eventPublisher';

export async function addMember(tenantId: string, input: { email: string; role: string }, requesterRole: string) {
  if (requesterRole !== 'owner') throw new AuthError('Forbidden');
  const res = await query(
    `INSERT INTO members (tenant_id, user_id, email, role, created_at, updated_at)
     VALUES ($1, gen_random_uuid(), $2, $3, NOW(), NOW()) RETURNING *`,
    [tenantId, input.email, input.role]
  );
  const row = res.rows[0];
  await publishMemberEvent('added', { memberId: row.id, tenantId, email: input.email, role: input.role }, undefined);
  return row;
}

export async function getMember(tenantId: string, memberId: string) {
  const res = await query(`SELECT * FROM members WHERE id = $1 AND tenant_id = $2`, [memberId, tenantId]);
  if (res.rows.length === 0) throw new NotFoundError('Member not found');
  return res.rows[0];
}

export async function updateMemberRole(tenantId: string, memberId: string, newRole: string, requesterRole: string) {
  if (requesterRole !== 'owner') throw new AuthError('Forbidden');
  await getMember(tenantId, memberId);
  const res = await query(`UPDATE members SET role = $1, updated_at = NOW() WHERE id = $2 AND tenant_id = $3 RETURNING *`, [newRole, memberId, tenantId]);
  const row = res.rows[0];
  await publishMemberEvent('role_changed', { memberId, tenantId, newRole }, undefined);
  return row;
}

export async function removeMember(tenantId: string, memberId: string, requesterRole: string) {
  if (requesterRole !== 'owner') throw new AuthError('Forbidden');
  await (await import('../db')).transaction(async (client: any) => {
    const owners = await client.query(`SELECT COUNT(1)::text as cnt FROM members WHERE tenant_id = $1 AND role = 'owner'`, [tenantId]);
    const cnt = parseInt(owners.rows[0]?.cnt ?? '0', 10);
    if (cnt <= 1) throw new Error('Cannot remove last owner');
    await client.query(`DELETE FROM members WHERE id = $1 AND tenant_id = $2`, [memberId, tenantId]);
  });
  await publishMemberEvent('removed', { memberId, tenantId }, undefined);
}
