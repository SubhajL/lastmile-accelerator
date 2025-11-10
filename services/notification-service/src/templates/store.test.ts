import { describe, it, expect, vi, beforeEach } from 'vitest';
import { createTemplateStore } from './store.js';

function mockRedis() {
  const kv = new Map<string, string>();
  const set = new Map<string, Set<string>>();
  return {
    get: vi.fn(async (k: string) => kv.get(k) ?? null),
    set: vi.fn(async (k: string, v: string) => { kv.set(k, v); }),
    del: vi.fn(async (k: string) => { kv.delete(k); }),
    sadd: vi.fn(async (k: string, v: string) => { if (!set.has(k)) set.set(k, new Set()); set.get(k)!.add(v); }),
    srem: vi.fn(async (k: string, v: string) => { set.get(k)?.delete(v); }),
    smembers: vi.fn(async (k: string) => Array.from(set.get(k) ?? []))
  } as any;
}

describe('templates/store', () => {
  it('put/get/delete/list template bundles', async () => {
    const redis = mockRedis();
    const store = createTemplateStore({ redis, namespace: 'tmpl' });

    await store.put('publish-failed', { subject: 'Sub', html: '<p>{{error}}</p>', text: 'txt' });
    const t1 = await store.get('publish-failed');
    expect(t1?.subject).toBe('Sub');

    const names = await store.list();
    expect(names).toContain('publish-failed');

    await store.delete('publish-failed');
    const t2 = await store.get('publish-failed');
    expect(t2).toBeNull();
  });
});
