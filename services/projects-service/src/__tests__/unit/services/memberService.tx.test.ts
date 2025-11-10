import { describe, it, expect, vi } from 'vitest';
import * as db from '../../../db';
import * as events from '../../../events/eventPublisher';
import { removeMember } from '../../../services/memberService';

describe('memberService.removeMember transactional integrity', () => {
  it('wraps owner-count check and deletion in a single transaction', async () => {
    const client = { query: vi.fn()
      .mockResolvedValueOnce({ rows: [{ cnt: '2' }] }) // owners count
      .mockResolvedValueOnce({ rows: [] }) // delete
    } as any;
    const trx = vi.spyOn(db, 'transaction').mockImplementation(async (fn: any) => fn(client));
    const ev = vi.spyOn(events, 'publishMemberEvent').mockResolvedValue();

    await removeMember('t1', 'm1', 'owner');

    expect(trx).toHaveBeenCalledTimes(1);
    expect(client.query).toHaveBeenCalledWith(expect.stringMatching(/COUNT/), ['t1']);
    expect(client.query).toHaveBeenCalledWith(expect.stringMatching(/DELETE FROM members/), ['m1', 't1']);
    expect(ev).toHaveBeenCalledWith('removed', expect.objectContaining({ memberId: 'm1', tenantId: 't1' }), undefined);
  });

  it('prevents deletion when last owner inside transaction', async () => {
    const client = { query: vi.fn().mockResolvedValueOnce({ rows: [{ cnt: '1' }] }) } as any;
    vi.spyOn(db, 'transaction').mockImplementation(async (fn: any) => fn(client));

    await expect(removeMember('t1', 'm1', 'owner')).rejects.toThrow('last owner');
    // no delete performed
    expect(client.query).toHaveBeenCalledTimes(1);
  });
});