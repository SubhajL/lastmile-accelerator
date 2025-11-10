import { describe, it, expect, vi, beforeEach } from 'vitest';
import * as db from '../../../db';
import { getTenant, listTenantMembers } from '../../../services/tenantService';
import { NotFoundError } from '../../../utils/errors';

describe('services/tenantService.ts', () => {
beforeEach(() => { vi.restoreAllMocks(); });

  it('getTenant returns non-deleted tenant', async () => {
    vi.spyOn(db, 'query').mockResolvedValueOnce({ rows: [{ id: 't1', name: 'Acme', deleted_at: null }] } as any);
    const t = await getTenant('t1');
    expect(t.id).toBe('t1');
  });

  it('getTenant throws for deleted tenant', async () => {
    vi.spyOn(db, 'query').mockResolvedValueOnce({ rows: [] } as any);
    await expect(getTenant('t1')).rejects.toBeInstanceOf(NotFoundError);
  });

  it('listTenantMembers returns all members ordered by email', async () => {
    vi.spyOn(db, 'query').mockResolvedValueOnce({ rows: [
      { id: 'm1', email: 'a@x.com' },
      { id: 'm2', email: 'b@x.com' },
    ] } as any);
    const rows = await listTenantMembers('t1');
    expect(rows.map(r => r.email)).toEqual(['a@x.com', 'b@x.com']);
  });
});