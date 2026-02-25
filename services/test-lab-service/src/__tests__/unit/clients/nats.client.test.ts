import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { createNatsConnection } from '../../../clients/nats.js';

vi.mock('nats', () => {
  const te = new TextEncoder();
  const td = new TextDecoder();
  return {
    connect: vi.fn(),
    StringCodec: () => ({
      encode: (s: string) => te.encode(s),
      decode: (u: Uint8Array) => td.decode(u),
    }),
    Subscription: vi.fn(),
  };
});

beforeEach(async () => {
  const { connect } = await import('nats');
  const subs: any[] = [];
  vi.mocked(connect).mockResolvedValue({
    publish: vi.fn(),
    subscribe: vi.fn().mockImplementation((subject: string) => {
      const sub = { subject, [Symbol.asyncIterator]: async function* () {} } as any;
      subs.push(sub);
      return sub;
    }),
    drain: vi.fn().mockResolvedValue(undefined),
  } as any);
});

afterEach(() => {
  vi.clearAllMocks();
});

describe('nats client', () => {
  it('publish encodes json and subscribe registers', async () => {
    const nats = await createNatsConnection('nats://localhost:4222');
    await nats.publish('foo', { a: 1 }, { h: 'v' });
    const sub = (await nats.subscribe('bar', async () => {})) as any;
    expect(sub.subject).toBe('bar');
    await nats.close();
  });
});
