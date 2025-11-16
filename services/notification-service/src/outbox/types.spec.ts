import { describe, it, expectTypeOf } from 'vitest';
import type { OutboxStatus, OutboxMessage, OutboxDedupKey } from './types.js';

describe('outbox/types', () => {
  it('OutboxStatus matches allowed status literals', () => {
    expectTypeOf<OutboxStatus>().toEqualTypeOf<'pending' | 'sent' | 'failed'>();
  });

  it('OutboxMessage has dedupKey/status/createdAt fields', () => {
    expectTypeOf<OutboxMessage>().toMatchTypeOf<{ dedupKey: OutboxDedupKey; status: OutboxStatus; createdAt: Date }>();
  });

  it('OutboxDedupKey is branded (not plain string)', () => {
    // string should not be assignable to branded key without cast
    expectTypeOf<string>().not.toMatchTypeOf<OutboxDedupKey>();
  });
});
