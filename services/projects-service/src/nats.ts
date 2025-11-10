import type { NatsConnection } from 'nats';

let nc: NatsConnection | null = null;

export interface NatsInitOptions {
  maxRetries?: number; // default 5
  baseDelayMs?: number; // default 100
}

// indirection for connect + sleep to make tests deterministic
// eslint-disable-next-line @typescript-eslint/ban-types
type Connector = (url: string) => Promise<NatsConnection>;
let connector: Connector | null = null;
let sleeper: (ms: number) => Promise<void> = (ms: number) => new Promise(res => setTimeout(res, ms));

export function __setNatsConnector(fn: Connector) {
  connector = fn;
}

export function __setSleeper(fn: (ms: number) => Promise<void>) {
  sleeper = fn;
}

async function connectWithRetry(url: string, opts: Required<NatsInitOptions>): Promise<NatsConnection> {
  if (!connector) {
    // Lazy import to avoid pulling real lib during tests
    const { connect } = await import('nats');
    // Wrap to always pass options object, as some versions mutate options
    connector = (async (u: string) => (await connect({ servers: u })) as unknown as any) as Connector;
  }

  let attempt = 0;
  let delay = opts.baseDelayMs;
  let lastErr: unknown;

  while (attempt < opts.maxRetries) {
    try {
      return await (connector as Connector)(url);
    } catch (err) {
      lastErr = err;
      attempt += 1;
      if (attempt >= opts.maxRetries) break;
      await sleeper(delay);
      delay *= 2; // exponential backoff
    }
  }

  throw lastErr instanceof Error ? lastErr : new Error('NATS connect failed');
}

export async function initNats(url: string, options: NatsInitOptions = {}): Promise<NatsConnection> {
  if (nc) return nc;
  const opts: Required<NatsInitOptions> = {
    maxRetries: options.maxRetries ?? 5,
    baseDelayMs: options.baseDelayMs ?? 100,
  };
  nc = await connectWithRetry(url, opts);
  return nc;
}

export function getNats(): NatsConnection {
  if (!nc) throw new Error('NATS not initialized. Call initNats() first.');
  return nc;
}

export async function publish(subject: string, payload: Record<string, any>, traceparent?: string): Promise<void> {
  const conn = getNats();
  const data = new TextEncoder().encode(JSON.stringify(payload));
  const options: any = {};
  if (traceparent) {
    options.headers = { ...(options.headers || {}), traceparent };
  }
  // ignore return value; fire-and-forget
  (conn as any).publish(subject, data, options);
}

export async function closeNats(): Promise<void> {
  if (!nc) return;
  const c = nc;
  nc = null;
  if (typeof (c as any).drain === 'function') {
    await (c as any).drain();
  }
  if (typeof (c as any).close === 'function') {
    await (c as any).close();
  }
}