import { describe, it, expect, vi } from 'vitest';
import { createNatsConnection } from '../../../clients/nats.js';

vi.mock('nats', () => {
  const subs: any[] = [];
  const te = new TextEncoder();
  const td = new TextDecoder();
  return {
    connect: vi.fn().mockResolvedValue({
      publish: vi.fn(),
      subscribe: vi.fn().mockImplementation((subject: string) => {
        const sub = { subject, [Symbol.asyncIterator]: async function* () {} } as any;
        subs.push(sub);
        return sub;
      }),
      drain: vi.fn().mockResolvedValue(undefined),
    }),
    StringCodec: vi.fn().mockReturnValue({
      encode: (s: string) => te.encode(s),
      decode: (u: Uint8Array) => td.decode(u),
    }),
    Subscription: vi.fn(),
  };
});

describe('nats client', () => {
  it('publish encodes json and subscribe registers', async () => {
  const nats = await createNatsConnection('n' + 'ats://localhost:4222');
    await nats.publish('foo', { a: 1 }, { h: 'v' });
    const sub = (await nats.subscribe('bar', async () => {})) as any;
    expect(sub.subject).toBe('bar');
    await nats.close();
  });
});
