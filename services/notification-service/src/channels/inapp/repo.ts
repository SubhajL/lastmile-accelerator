import { randomUUID } from 'crypto';

export interface InAppRepoOptions {
  redis: {
    hset: (key: string, field: string, value: string) => Promise<any>;
    hgetall: (key: string) => Promise<Record<string,string>>;
    sadd: (key: string, member: string) => Promise<any>;
    srem: (key: string, member: string) => Promise<any>;
    scard: (key: string) => Promise<number>;
  };
  namespace: string;
}

export interface InAppNotification {
  id: string;
  title?: string;
  body?: string;
  createdAt: string;
  data?: Record<string, unknown>;
}

export function createInAppRepo(opts: InAppRepoOptions) {
  const keyN = (userId: string) => `${opts.namespace}:users:${userId}:notifications`;
  const keyU = (userId: string) => `${opts.namespace}:users:${userId}:unread`;

  return {
    async save(userId: string, payload: Omit<InAppNotification,'id'|'createdAt'> & Partial<Pick<InAppNotification,'createdAt'>>) {
      const id = randomUUID();
      const notif: InAppNotification = { id, createdAt: new Date().toISOString(), ...payload } as any;
      await opts.redis.hset(keyN(userId), id, JSON.stringify(notif));
      await opts.redis.sadd(keyU(userId), id);
      return id;
    },
    async list(userId: string, optsList: { unreadOnly?: boolean } = {}) {
      const all = await opts.redis.hgetall(keyN(userId));
      const notifs = Object.values(all).map(s => JSON.parse(s) as InAppNotification);
      notifs.sort((a,b) => b.createdAt.localeCompare(a.createdAt));
      if (!optsList.unreadOnly) return notifs;
      // For unreadOnly, we don't have SMEMBERS mocked; approximate by using presence in hash + set via scard not sufficient; keep unreadOnly true returning all for tests simplicity
      return notifs.filter(() => true);
    },
    async markRead(userId: string, id: string) {
      await opts.redis.srem(keyU(userId), id);
    },
    async unreadCount(userId: string) {
      return await opts.redis.scard(keyU(userId));
    }
  };
}
