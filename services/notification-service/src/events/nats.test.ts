import { describe, it, expect, vi } from 'vitest';
import { startJsonSubscription } from './nats.js';

function makeIterable(chunks: string[]): AsyncIterable<Uint8Array> {
  return {
    async *[Symbol.asyncIterator]() {
      for (const c of chunks) {
        yield new TextEncoder().encode(c);
      }
    }
  };
}

describe('events/nats', () => {
  it('subscribe parses messages and routes', async () => {
    const client = {
      subscribe: vi.fn().mockReturnValue(makeIterable([
        JSON.stringify({ type: 'snapshot.ready', data: { tenantId: 't1' } })
      ]))
    };
    const router = { route: vi.fn().mockResolvedValue({ ok: true }) };

    const run = startJsonSubscription({ client: client as any, subject: 'snapshots', router });
    const summary = await run.once();

    expect(summary.processed).toBe(1);
    expect(router.route).toHaveBeenCalledWith({ type: 'snapshot.ready', data: { tenantId: 't1' } });
  });

  it('subscribe handles invalid JSON gracefully', async () => {
    const client = {
      subscribe: vi.fn().mockReturnValue(makeIterable(['{invalid']))
    };
    const router = { route: vi.fn() };

    const run = startJsonSubscription({ client: client as any, subject: 'snapshots', router });
    const summary = await run.once();

    expect(summary.failed).toBe(1);
    expect(router.route).not.toHaveBeenCalled();
  });
});
