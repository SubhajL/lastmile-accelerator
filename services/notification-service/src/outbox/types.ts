export type Brand<T, B extends string> = T & { readonly __brand: B };

export const OUTBOX_TABLE = 'notification_outbox_messages' as const;

export type OutboxStatus = 'pending' | 'sent' | 'failed';

export type OutboxDedupKey = Brand<string, 'OutboxDedupKey'>;

export interface OutboxMessage {
  dedupKey: OutboxDedupKey;
  status: OutboxStatus;
  createdAt: Date;
}

export function isOutboxStatus(x: unknown): x is OutboxStatus {
  return x === 'pending' || x === 'sent' || x === 'failed';
}
