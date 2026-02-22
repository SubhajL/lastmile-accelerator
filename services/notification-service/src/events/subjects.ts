// Typed subject registry for NATS event names and subscription patterns

export type Brand<T, B extends string> = T & { readonly __brand: B };

export type Subject = Brand<string, 'NatsSubject'>;
export type SubjectPattern = Brand<string, 'NatsSubjectPattern'>;

export const Subjects = Object.freeze({
  snapshot: Object.freeze({
    ready: 'snapshot.ready' as Subject
  }),
  fixes: Object.freeze({
    created: 'fixes.created' as Subject,
    applied: 'fixes.applied' as Subject
  }),
  publish: Object.freeze({
    started: 'publish.started' as Subject,
    healthy: 'publish.healthy' as Subject,
    rolledback: 'publish.rolledback' as Subject,
    failed: 'publish.failed' as Subject
  }),
  checks: Object.freeze({
    failed: 'checks.failed' as Subject
  }),
  slo: Object.freeze({
    budget_exhausted: 'slo.budget_exhausted' as Subject
  }),
  errors: Object.freeze({
    critical: 'errors.critical' as Subject
  })
} as const);

export type EventType =
  | typeof Subjects.snapshot.ready
  | typeof Subjects.fixes.created
  | typeof Subjects.fixes.applied
  | typeof Subjects.publish.started
  | typeof Subjects.publish.healthy
  | typeof Subjects.publish.rolledback
  | typeof Subjects.publish.failed
  | typeof Subjects.checks.failed
  | typeof Subjects.slo.budget_exhausted
  | typeof Subjects.errors.critical;

const EVENT_SET: ReadonlySet<string> = new Set<string>([
  Subjects.snapshot.ready,
  Subjects.fixes.created,
  Subjects.fixes.applied,
  Subjects.publish.started,
  Subjects.publish.healthy,
  Subjects.publish.rolledback,
  Subjects.publish.failed,
  Subjects.checks.failed,
  Subjects.slo.budget_exhausted,
  Subjects.errors.critical
]);

export function isEventType(x: string): x is EventType {
  return EVENT_SET.has(x);
}

export const SubscriptionSubjects: readonly SubjectPattern[] = [
  'snapshot.*' as SubjectPattern,
  'fixes.*' as SubjectPattern,
  'publish.*' as SubjectPattern,
  'checks.*' as SubjectPattern,
  'slo.*' as SubjectPattern,
  'errors.*' as SubjectPattern
] as const;
