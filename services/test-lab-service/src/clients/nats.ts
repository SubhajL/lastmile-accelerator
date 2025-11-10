import { connect, NatsConnection, StringCodec, Subscription } from 'nats';

export interface NatsWrapper {
  nc: NatsConnection;
  publish: (subject: string, payload: unknown, headers?: Record<string, string>) => Promise<void>;
  subscribe: (subject: string, handler: (data: any) => Promise<void>) => Promise<Subscription>;
  close: () => Promise<void>;
}

export async function createNatsConnection(url: string): Promise<NatsWrapper> {
  const nc = await connect({ servers: url });
  const sc = StringCodec();

  async function publish(subject: string, payload: unknown, headers?: Record<string, string>) {
    const msg = sc.encode(JSON.stringify(payload));
    const h = headers ? Object.entries(headers).reduce((acc, [k, v]) => { acc[k] = v; return acc; }, {} as any) : undefined;
    nc.publish(subject, msg, { headers: h as any });
  }

  async function subscribe(subject: string, handler: (data: any) => Promise<void>) {
    const sub = nc.subscribe(subject);
    (async () => {
      for await (const m of sub) {
        try {
          const data = JSON.parse(sc.decode(m.data));
          await handler(data);
        } catch (e) {
          // swallow errors to avoid breaking sub loop
        }
      }
    })();
    return sub;
  }

  async function close() { await nc.drain(); }

  return { nc, publish, subscribe, close };
}
