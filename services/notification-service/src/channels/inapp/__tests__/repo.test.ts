import { describe, it, expect, vi } from 'vitest';
import { createInAppRepo } from '../../inapp/repo.js';

function mockRedis() {
  const store = new Map<string, Record<string,string>>();
  const sets = new Map<string, Set<string>>();
  return {
    hset: vi.fn(async (key: string, field: string, value: string) => {
      const rec = store.get(key) || {};
      rec[field] = value;
      store.set(key, rec);
    }),
    hgetall: vi.fn(async (key: string) => {
      return store.get(key) || {};
    }),
    sadd: vi.fn(async (key: string, member: string) => { if (!sets.has(key)) sets.set(key,new Set()); sets.get(key)!.add(member); }),
    srem: vi.fn(async (key: string, member: string) => { sets.get(key)?.delete(member); }),
    scard: vi.fn(async (key: string) => sets.get(key)?.size ?? 0),
    publish: vi.fn(async (_ch: string, _payload: string) => 1)
  } as any;
}

describe('in-app repo', () => {
  it('save, list and markRead work', async () => {
    const redis = mockRedis();
    const repo = createInAppRepo({ redis, namespace: 'inapp' });

    const id = await repo.save('u1', { title: 't', body: 'b' });
    const list1 = await repo.list('u1', { unreadOnly: true });
    expect(list1.find(n => n.id === id)).toBeTruthy();

    await repo.markRead('u1', id);
    const count = await repo.unreadCount('u1');
    expect(count).toBe(0);
  });
});
