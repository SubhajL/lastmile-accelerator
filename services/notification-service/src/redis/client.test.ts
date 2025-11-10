import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { createRedisClient } from './client.js';

class DummyRedis {
  static instances: any[] = [];
  url: string;
  options: any;
  listeners: Record<string, Function[]> = {};
  constructor(url: string, options: any) {
    this.url = url;
    this.options = options;
    DummyRedis.instances.push(this);
  }
  on(evt: string, fn: Function) {
    this.listeners[evt] = this.listeners[evt] || [];
    this.listeners[evt].push(fn);
    return this;
  }
  quit = vi.fn().mockResolvedValue(undefined);
}

describe('redis/client', () => {
  const logger = { error: vi.fn() };
  const cfg = { url: 'redis://localhost:6379', maxRetriesPerRequest: 3 };

  beforeEach(() => {
    DummyRedis.instances = [];
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('constructs redis client with url and options', () => {
    const r = createRedisClient(cfg as any, { RedisCtor: DummyRedis as any, logger });

    expect(DummyRedis.instances.length).toBe(1);
    expect(DummyRedis.instances[0].url).toBe('redis://localhost:6379');
    expect(DummyRedis.instances[0].options.maxRetriesPerRequest).toBe(3);
    expect(r.on).toBeDefined();
  });

  it('attaches error handler', () => {
    const r: any = createRedisClient(cfg as any, { RedisCtor: DummyRedis as any, logger });
    const inst = DummyRedis.instances[0];
    expect(inst.listeners['error']).toBeTruthy();

    const err = new Error('boom');
    inst.listeners['error'][0](err);
    expect(logger.error).toHaveBeenCalled();
  });
});
