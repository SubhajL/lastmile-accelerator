import { createDefaultChannelRegistry } from '../channels/registry.js';
import {
  createDispatcher,
  type ChannelRegistry as DispatcherChannelRegistry,
} from '../notifications/dispatcher.js';
import { createEventRouter } from '../consumers/router.js';
import { Subjects } from '../events/subjects.js';
import { createSnapshotConsumer } from '../consumers/snapshot-consumer.js';
import { createFixesConsumer } from '../consumers/fixes-consumer.js';
import { createPublishConsumer } from '../consumers/publish-consumer.js';
import { createChecksConsumer } from '../consumers/checks-consumer.js';
import { createSloConsumer } from '../consumers/slo-consumer.js';
import { createErrorsConsumer } from '../consumers/errors-consumer.js';
import type { TemplateRenderer } from '../channels/email.js';

export interface RuntimeOptions {
  queue: {
    dequeue: (n: number) => Promise<any[]>;
    ack: (id: string) => Promise<void>;
    nack: (id: string, error: string, retry: boolean) => Promise<void>;
    enqueue?: (job: any) => Promise<string>;
  };
  subscribe: (args: { subject: string; router: { route: (env: any) => Promise<any> } }) => {
    once: () => Promise<{ processed: number; failed: number }>;
  };
  dispatcherFactory?: typeof createDispatcher;
  metrics: { increment: (name: string, labels?: Record<string, string | number>) => void };
  now: () => number;
  subjects: string[];
  batchSize: number;
  email: {
    from: string;
    transporter: any;
    renderTemplate: TemplateRenderer;
    resolveTo: (job: any) => Promise<string>;
  };
  defaultMaxAttempts: number;
  registry?: DispatcherChannelRegistry;
  routingEngine?: {
    evaluate: (
      job: any,
    ) => Promise<
      | { status: 'allow' }
      | { status: 'block'; reason: string }
      | { status: 'defer'; reason: string }
    >;
  };
}

function toEventHandler<T>(fn: (evt: T) => Promise<unknown>) {
  return async (data: unknown) => fn(data as T) as any;
}

export function createRuntime(opts: RuntimeOptions) {
  const registry = opts.registry ?? createDefaultChannelRegistry({ email: opts.email });
  const dispatcher = (opts.dispatcherFactory ?? createDispatcher)({
    queue: opts.queue as any,
    registry,
    metrics: opts.metrics,
    now: opts.now,
    routing: opts.routingEngine as any,
  });

  const snapshotConsumer = createSnapshotConsumer({
    queue: opts.queue as any,
    defaultMaxAttempts: opts.defaultMaxAttempts,
    resolveRecipients: async (evt: any) => [evt.userId ?? 'user-unknown'],
  });
  const fixesConsumer = createFixesConsumer({
    queue: opts.queue as any,
    defaultMaxAttempts: opts.defaultMaxAttempts,
    resolveRecipients: async () => ['u-fixes'],
  });
  const publishConsumer = createPublishConsumer({
    queue: opts.queue as any,
    defaultMaxAttempts: opts.defaultMaxAttempts,
    resolveRecipients: async () => ['u-publish'],
  });
  const checksConsumer = createChecksConsumer({
    queue: opts.queue as any,
    defaultMaxAttempts: opts.defaultMaxAttempts,
    resolveRecipients: async () => ['u-checks'],
  });
  const sloConsumer = createSloConsumer({
    queue: opts.queue as any,
    defaultMaxAttempts: opts.defaultMaxAttempts,
    resolveRecipients: async () => ['u-slo'],
  });
  const errorsConsumer = createErrorsConsumer({
    queue: opts.queue as any,
    defaultMaxAttempts: opts.defaultMaxAttempts,
    resolveRecipients: async () => ['u-errors'],
  });

  const router = createEventRouter({
    handlers: {
      // event envelope type keys — centralized constants
      [Subjects.snapshot.ready]: toEventHandler(snapshotConsumer.handle),
      [Subjects.fixes.created]: toEventHandler(fixesConsumer.handleCreated),
      [Subjects.fixes.applied]: toEventHandler(fixesConsumer.handleApplied),
      // publish events route to a single consumer; the event subtype is in payload.status
      [Subjects.publish.started]: toEventHandler(publishConsumer.handle),
      [Subjects.publish.healthy]: toEventHandler(publishConsumer.handle),
      [Subjects.publish.rolledback]: toEventHandler(publishConsumer.handle),
      [Subjects.publish.failed]: toEventHandler(publishConsumer.handle),
      [Subjects.checks.failed]: toEventHandler(checksConsumer.handle),
      [Subjects.slo.budget_exhausted]: toEventHandler(sloConsumer.handle),
      [Subjects.errors.critical]: toEventHandler(errorsConsumer.handle),
    },
  });

  const subscriptions = opts.subjects.map((subject) => ({
    subject,
    sub: opts.subscribe({ subject, router }),
  }));

  async function startOnce() {
    const results = await Promise.all(subscriptions.map((s) => s.sub.once()));
    // Increment per subject summary
    for (let i = 0; i < results.length; i++) {
      const { processed, failed } = results[i] as any;
      const subject = subscriptions[i]!.subject;
      if (processed) opts.metrics.increment('nats_processed', { subject });
      if (failed) opts.metrics.increment('nats_failed', { subject });
    }
    return dispatcher.processNextBatch(opts.batchSize);
  }

  return { startOnce, processNextBatch: dispatcher.processNextBatch, router };
}

export interface SubscriptionHandle {
  stop: () => Promise<void>;
}

export async function startNatsSubscriptions(
  subjects: readonly string[],
  subscribe: RuntimeOptions['subscribe'],
  router: ReturnType<typeof createEventRouter>,
): Promise<SubscriptionHandle> {
  const subscriptionPromises: Promise<unknown>[] = [];

  for (const subject of subjects) {
    const sub = subscribe({ subject, router });
    const promise = sub.once().catch(() => undefined);
    subscriptionPromises.push(promise);
  }

  return {
    stop: async () => {
      // Subscriptions will naturally close when NATS connection closes
      // Wait a moment for any in-flight message processing
      await Promise.race([
        Promise.allSettled(subscriptionPromises),
        new Promise((resolve) => setTimeout(resolve, 1000)),
      ]);
    },
  };
}
