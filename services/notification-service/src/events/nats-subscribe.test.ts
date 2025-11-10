import { describe, it, expect, vi } from 'vitest';
import { createNatsSubscribe } from './nats-subscribe.js';

function makeAsyncIterable(messages: Array<{ data: Uint8Array } | Uint8Array>): AsyncIterable<any> {
  return {
    async *[Symbol.asyncIterator]() {
      for (const m of messages) yield m as any;
    }
  };
}

describe('events/nats-subscribe', () => {
  it('subscribe wraps nats subscription and routes JSON', async () => {
    const nc = {
      subscribe: vi.fn().mockReturnValue(makeAsyncIterable([
        { data: new TextEncoder().encode(JSON.stringify({ type: 'snapshot.ready', data: { x: 1 } })) }
      ]))
    };
    const router = { route: vi.fn().mockResolvedValue({ ok: true }) };

    const subscribe = createNatsSubscribe(nc as any);
    const res = await subscribe({ subject: 'snapshots', router }).once();

    expect(res).toEqual({ processed: 1, failed: 0 });
    expect(router.route).toHaveBeenCalledWith({ type: 'snapshot.ready', data: { x: 1 } });
  });

  it('counts failed messages on invalid JSON', async () => {
    const nc = { subscribe: vi.fn().mockReturnValue(makeAsyncIterable([{ data: new TextEncoder().encode('{oops') }])) };
    const router = { route: vi.fn() };

    const subscribe = createNatsSubscribe(nc as any);
    const res = await subscribe({ subject: 'snapshots', router }).once();

    expect(res).toEqual({ processed: 0, failed: 1 });
    expect(router.route).not.toHaveBeenCalled();
  });
});
