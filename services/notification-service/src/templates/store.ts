export interface TemplateBundle { subject?: string; html?: string; text?: string }

export function createTemplateStore(opts: { redis: { get: (key:string)=>Promise<string|null>; set: (key:string, val:string)=>Promise<unknown>; del: (key:string)=>Promise<unknown>; sadd: (key:string, member:string)=>Promise<unknown>; srem: (key:string, member:string)=>Promise<unknown>; smembers: (key:string)=>Promise<string[]> }; namespace: string }) {
  const key = (name: string) => `${opts.namespace}:templates:${name}`;
  const idx = `${opts.namespace}:templates:index`;

  return {
    async get(name: string): Promise<TemplateBundle|null> {
      const raw = await opts.redis.get(key(name));
      return raw ? JSON.parse(raw) as TemplateBundle : null;
    },
    async put(name: string, bundle: TemplateBundle): Promise<void> {
      await opts.redis.set(key(name), JSON.stringify(bundle));
      await opts.redis.sadd(idx, name);
    },
    async delete(name: string): Promise<void> {
      await opts.redis.del(key(name));
      await opts.redis.srem(idx, name);
    },
    async list(): Promise<string[]> {
      return await opts.redis.smembers(idx);
    }
  };
}
