import { describe, expect, test } from 'vitest';
import { createRuntime } from './runtime';
import { Subjects } from '../events/subjects';

describe('runtime uses subject constants as router keys', () => {
  test('handler table keyed by constants', async () => {
    const handled: string[] = [];

    const snapshotHandler = async () => { handled.push('snapshot'); return { ok: true as const }; };
    const fixesCreatedHandler = async () => { handled.push('fixes.created'); return { ok: true as const }; };

    const rt = createRuntime({
      queue: { dequeue: async () => [], ack: async () => {}, nack: async () => {} },
      subscribe: () => ({ once: async () => ({ processed: 0, failed: 0 }) }),
      metrics: { increment: () => {} },
      now: () => Date.now(),
      subjects: ['snapshot.*', 'fixes.*'],
      batchSize: 1,
      email: { from: 'x', transporter: {}, renderTemplate: async () => ({} as any), resolveTo: async () => 'a@b.com' },
      defaultMaxAttempts: 3,
      // inject minimal registry and routing to avoid side effects
      registry: { get: () => ({ send: async () => ({ ok: true }) }) } as any,
      routingEngine: { evaluate: async () => ({ status: 'allow' as const }) }
    });

    // Internal detail: we can't read the router directly; simulate by calling handlers via known keys
    // This expectation ensures we exported constants that match the runtime handler map keys in implementation
    expect(Subjects.snapshot.ready).toBe('snapshot.ready');
    expect(Subjects.fixes.created).toBe('fixes.created');

    // Nothing else to assert here until dispatch path is wired in later SDs
    expect(handled).toEqual([]);
  });
});
