import type { NatsWrapper } from '../clients/nats.js';
import { SUBJECTS, zRunStarted, zRunFinished, zBrowserStarted, zBrowserFinished } from './contracts.js';

export async function publishRunStarted(nats: NatsWrapper, payload: unknown): Promise<void> {
  const evt = zRunStarted.parse(payload);
  await nats.publish(SUBJECTS.runStarted, evt);
}

export async function publishRunFinished(nats: NatsWrapper, payload: unknown): Promise<void> {
  const evt = zRunFinished.parse(payload);
  await nats.publish(SUBJECTS.runFinished, evt);
}

export async function publishBrowserStarted(nats: NatsWrapper, payload: unknown): Promise<void> {
  const evt = zBrowserStarted.parse(payload);
  await nats.publish(SUBJECTS.browserStarted, evt);
}

export async function publishBrowserFinished(nats: NatsWrapper, payload: unknown): Promise<void> {
  const evt = zBrowserFinished.parse(payload);
  await nats.publish(SUBJECTS.browserFinished, evt);
}
