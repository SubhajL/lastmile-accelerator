import { query } from '../db';
import { NotFoundError } from '../utils/errors';

export async function getTenant(tenantId: string) {
  const res = await query(`SELECT * FROM tenants WHERE id = $1 AND deleted_at IS NULL`, [tenantId]);
  if (res.rows.length === 0) throw new NotFoundError('Tenant not found');
  return res.rows[0];
}

export async function listTenantMembers(tenantId: string) {
  const res = await query(`SELECT * FROM members WHERE tenant_id = $1 ORDER BY email ASC`, [tenantId]);
  return res.rows;
}