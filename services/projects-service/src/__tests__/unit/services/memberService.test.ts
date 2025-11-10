import { describe, it, expect, vi, beforeEach } from 'vitest';
import * as db from '../../../db';
import * as events from '../../../events/eventPublisher';
import { AuthError, NotFoundError } from '../../../utils/errors';
import { addMember, getMember, updateMemberRole, removeMember } from '../../../services/memberService';

describe('services/memberService.ts', () => {
  beforeEach(() => { vi.restoreAllMocks(); });

  it('addMember requires owner and publishes member.added', async () => {
    const q = vi.spyOn(db, 'query').mockResolvedValue({ rows: [{ id: 'm1', tenant_id: 't1', email: 'u@x.com', role: 'developer' }] } as any);
    const ev = vi.spyOn(events, 'publishMemberEvent').mockResolvedValue();

    await expect(addMember('t1', { email: 'u@x.com', role: 'developer' }, 'admin')).rejects.toBeInstanceOf(AuthError);

    const res = await addMember('t1', { email: 'u@x.com', role: 'developer' }, 'owner');
    expect(q).toHaveBeenCalled();
    expect(ev).toHaveBeenCalledWith('added', expect.objectContaining({ tenantId: 't1' }), undefined);
    expect(res.email).toBe('u@x.com');
  });

  it('getMember enforces tenant isolation', async () => {
    const q = vi.spyOn(db, 'query');
    q.mockResolvedValueOnce({ rows: [{ id: 'm1', tenant_id: 't1' }] } as any);
    const m = await getMember('t1', 'm1');
    expect(m.id).toBe('m1');

    q.mockResolvedValueOnce({ rows: [] } as any);
    await expect(getMember('t2', 'mX')).rejects.toBeInstanceOf(NotFoundError);
  });

  it('updateMemberRole requires owner and publishes role_changed', async () => {
    const q = vi.spyOn(db, 'query');
    q.mockResolvedValueOnce({ rows: [{ id: 'm1', tenant_id: 't1', role: 'developer' }] } as any); // existing
    q.mockResolvedValueOnce({ rows: [{ id: 'm1', tenant_id: 't1', role: 'admin' }] } as any); // updated
    const ev = vi.spyOn(events, 'publishMemberEvent').mockResolvedValue();

    await expect(updateMemberRole('t1', 'm1', 'admin', 'developer')).rejects.toBeInstanceOf(AuthError);

    const res = await updateMemberRole('t1', 'm1', 'admin', 'owner');
    expect(res.role).toBe('admin');
    expect(ev).toHaveBeenCalledWith('role_changed', expect.objectContaining({ memberId: 'm1', tenantId: 't1', newRole: 'admin' }), undefined);
  });

  it('removeMember prevents removing last owner', async () => {
    vi.spyOn(db, 'transaction').mockImplementation(async (fn: any) => {
      const client = { query: vi.fn().mockResolvedValueOnce({ rows: [{ cnt: '1' }] }) } as any;
      return fn(client);
    });

    await expect(removeMember('t1', 'm1', 'owner')).rejects.toThrow('last owner');

    vi.spyOn(db, 'transaction').mockImplementation(async (fn: any) => {
      const client = { query: vi.fn()
        .mockResolvedValueOnce({ rows: [{ cnt: '2' }] })
        .mockResolvedValueOnce({ rows: [] }) } as any;
      return fn(client);
    });
    const ev = vi.spyOn(events, 'publishMemberEvent').mockResolvedValue();

    await removeMember('t1', 'm1', 'owner');
    expect(ev).toHaveBeenCalledWith('removed', expect.objectContaining({ memberId: 'm1', tenantId: 't1' }), undefined);
  });
});