export interface InAppPublisherOptions {
  redis: { publish: (channel: string, payload: string) => Promise<any> };
  channelPrefix: string;
  repo: { save: (userId: string, payload: any) => Promise<string> };
  metrics: { increment: (name: string, labels?: Record<string, string | number>) => void };
}

export function createInAppPublisher(opts: InAppPublisherOptions) {
  return {
    async publish(userId: string, payload: any) {
      const id = await opts.repo.save(userId, payload);
      await opts.redis.publish(`${opts.channelPrefix}:${userId}`, JSON.stringify({ id, ...payload }));
      opts.metrics.increment('notify_sent', { channel: 'in-app' });
      return id;
    }
  };
}
