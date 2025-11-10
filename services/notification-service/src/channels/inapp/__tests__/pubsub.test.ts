import { describe, it, expect, vi } from 'vitest';
import { createInAppPublisher } from '../../inapp/pubsub.js';
import { createInAppChannel } from '../../inapp/channel.js';

function mockRedis() {
  return {
    publish: vi.fn().mockResolvedValue(1)
  } as any;
}

describe('in-app publisher and channel', () => {
  it('publisher publishes payload to user channel', async () => {
    const redis = mockRedis();
    const repo = { save: vi.fn().mockResolvedValue('nid') } as any;
    const pub = createInAppPublisher({ redis, channelPrefix: 'inapp', repo, metrics: { increment: vi.fn() } });

    await pub.publish('u1', { title: 't' });

    expect(repo.save).toHaveBeenCalled();
    expect(redis.publish).toHaveBeenCalledWith('inapp:u1', expect.any(String));
  });

  it('channel adapts NotificationJob to in-app payload', async () => {
    const redis = mockRedis();
    const repo = { save: vi.fn().mockResolvedValue('nid') } as any;
    const pub = createInAppPublisher({ redis, channelPrefix: 'inapp', repo, metrics: { increment: vi.fn() } });
    const ch = createInAppChannel({ publisher: pub, resolveToUserId: async (job:any) => job.userId });

    const res = await ch.send({ userId: 'u1', templateName: 'publish-failed', payload: { snapshotId: 's1' } } as any);

    expect(res).toEqual({ ok: true });
    expect(repo.save).toHaveBeenCalled();
  });
});
