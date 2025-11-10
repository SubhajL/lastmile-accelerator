import { startJsonSubscription } from './nats.js';

export function createNatsSubscribe(nc: { subscribe: (subject: string, opts?: unknown) => AsyncIterable<any> }) {
  return function subscribe({ subject, router }: { subject: string; router: { route: (env: any) => Promise<{ ok: boolean }> } }) {
    // Adapt the nats connection to the shape accepted by startJsonSubscription
    const client = {
      subscribe: (subj: string) => nc.subscribe(subj)
    };
    return startJsonSubscription({ client, subject, router });
  };
}
