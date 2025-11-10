export interface EventEnvelope {
  type: string;
  data: unknown;
}

export type HandlerResult =
  | { ok: true }
  | { ok: false; error: string };

export type EventHandler = (data: any) => Promise<HandlerResult>;

export interface EventRouterOptions {
  handlers: Record<string, EventHandler>;
}

export function createEventRouter(opts: EventRouterOptions) {
  return {
    async route(envelope: EventEnvelope): Promise<HandlerResult> {
      const handler = opts.handlers[envelope.type];
      if (!handler) {
        return { ok: false, error: `No handler for type ${envelope.type}` };
      }
      try {
        return await handler(envelope.data);
      } catch (err) {
        const msg = err instanceof Error ? err.message : String(err);
        return { ok: false, error: msg };
      }
    }
  };
}
