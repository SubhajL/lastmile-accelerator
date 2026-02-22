import type IORedis from 'ioredis';
import type { RedisConfig } from '../types.js';

type RedisCtorOptions = { maxRetriesPerRequest?: number };
export interface CreateRedisClientOptions {
  RedisCtor?: new (url: string, options?: RedisCtorOptions) => IORedis;
  logger?: { error?: (...args: unknown[]) => void };
}

export function createRedisClient(config: RedisConfig, opts: CreateRedisClientOptions) {
  if (!opts?.RedisCtor) throw new Error('RedisCtor required');
  const RedisClass = opts.RedisCtor as unknown as new (url: string, options?: RedisCtorOptions) => IORedis;
  const client = new RedisClass(config.url, {
    maxRetriesPerRequest: config.maxRetriesPerRequest
  });

  client.on('error', (err: unknown) => {
    try { opts.logger?.error?.('redis error', err); } catch { /* noop */ }
  });

  return client as unknown as IORedis;
}
