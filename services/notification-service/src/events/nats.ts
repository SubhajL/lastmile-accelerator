export interface JsonRouter {
  route: (envelope: { type: string; data: unknown }) => Promise<{ ok: boolean }>;
}

export interface NatsLikeClient {
  // For tests, subscribe returns an async iterable of messages (Uint8Array payloads)
  subscribe: (subject: string, opts?: unknown) => AsyncIterable<Uint8Array> | AsyncIterable<{ data: Uint8Array }>;
}

export interface StartJsonSubscriptionOptions {
  client: NatsLikeClient;
  subject: string;
  router: JsonRouter;
  decoder?: TextDecoder;
}

export function startJsonSubscription(opts: StartJsonSubscriptionOptions) {
  const decoder = opts.decoder ?? new TextDecoder();

  async function once(): Promise<{ processed: number; failed: number }> {
    const sub = opts.client.subscribe(opts.subject);
    let processed = 0;
    let failed = 0;

    for await (const raw of sub as any) {
      const bytes: Uint8Array = (raw && 'data' in raw) ? (raw as any).data : (raw as Uint8Array);
      const text = decoder.decode(bytes);
      try {
        const envelope = JSON.parse(text);
        await opts.router.route(envelope);
        processed += 1;
      } catch {
        failed += 1;
      }
    }

    return { processed, failed };
  }

  return { once };
}
