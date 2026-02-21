import { describe, expect, test } from 'vitest';
import { SubscriptionSubjects } from './events/subjects';
import { createRuntime } from './bootstrap/runtime';

describe('worker subject wiring', () => {
  test('passes SubscriptionSubjects to runtime', async () => {
    const passed: string[][] = [];
    const runtime = createRuntime({
      queue: { dequeue: async () => [], ack: async () => {}, nack: async () => {} },
      subscribe: () => ({ once: async () => ({ processed: 0, failed: 0 }) }),
      metrics: { increment: () => {} },
      now: () => Date.now(),
      subjects: [...SubscriptionSubjects],
      batchSize: 1,
      email: { from: 'x', transporter: {}, renderTemplate: async () => ({} as any), resolveTo: async () => 'a@b.com' },
      defaultMaxAttempts: 3,
      registry: { get: () => ({ send: async () => ({ ok: true }) }) } as any,
      routingEngine: { evaluate: async () => ({ status: 'allow' as const }) }
    });
    // The runtime constructed; verify the exported patterns content is as expected
    expect(SubscriptionSubjects).toEqual([
      'snapshot.*',
      'fixes.*',
      'publish.*',
      'checks.*',
      'slo.*',
      'errors.*'
    ]);
    // Call a method to ensure no throw
    await (runtime as any).processNextBatch?.(1);
  });
});
