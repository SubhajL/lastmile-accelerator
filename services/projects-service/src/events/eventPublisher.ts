import { publish } from '../nats';

type ProjectEventType = 'created' | 'updated' | 'deleted';
type MemberEventType = 'added' | 'role_changed' | 'removed';
type EnvironmentEventType = 'created' | 'updated' | 'deleted';

type BasePayload = { tenantId: string } & Record<string, any>;

import { getActiveTraceparent } from '../observability/trace';

export async function publishProjectEvent(type: ProjectEventType, payload: BasePayload, traceparent?: string) {
  const subject = `projects.${type}`;
  const tp = traceparent || getActiveTraceparent();
  await publish(subject, { ...payload, timestamp: new Date().toISOString() }, tp);
}

export async function publishMemberEvent(type: MemberEventType, payload: BasePayload, traceparent?: string) {
  const subject = `members.${type}`;
  const tp = traceparent || getActiveTraceparent();
  await publish(subject, { ...payload, timestamp: new Date().toISOString() }, tp);
}

export async function publishEnvironmentEvent(type: EnvironmentEventType, payload: BasePayload, traceparent?: string) {
  const subject = `environments.${type}`;
  const tp = traceparent || getActiveTraceparent();
  await publish(subject, { ...payload, timestamp: new Date().toISOString() }, tp);
}
