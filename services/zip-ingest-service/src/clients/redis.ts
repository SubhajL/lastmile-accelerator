import { createClient } from 'redis';
import type { RedisClientType } from 'redis';

export type RedisClientLogger = {
  error: (obj: unknown, msg?: string) => void;
};

export type CreateRedisClientOptions = {
  logger?: RedisClientLogger;
};

export type RedisWrapper = {
  get: (key: string) => Promise<string | null>;
  set: (key: string, value: string, ttlSeconds?: number) => Promise<void>;
  del: (key: string) => Promise<void>;
  close: () => Promise<void>;
};

export async function createRedisClient(
  url: string,
  opts: CreateRedisClientOptions = {},
): Promise<RedisWrapper> {
  const client: RedisClientType = createClient({ url });

  client.on('error', (err) => {
    opts.logger?.error({ err }, 'Redis client error');
  });

  await client.connect();

  async function get(key: string): Promise<string | null> {
    return await client.get(key);
  }

  async function set(key: string, value: string, ttlSeconds?: number): Promise<void> {
    if (ttlSeconds !== undefined && ttlSeconds > 0) {
      await client.setEx(key, ttlSeconds, value);
      return;
    }
    await client.set(key, value);
  }

  async function del(key: string): Promise<void> {
    await client.del(key);
  }

  async function close(): Promise<void> {
    await client.quit();
  }

  return { get, set, del, close };
}

