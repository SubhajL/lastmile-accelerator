export class TimeoutError extends Error {
  constructor(message = 'Operation timed out') {
    super(message);
    this.name = 'TimeoutError';
  }
}

export async function withTimeout<T>(op: (signal: AbortSignal) => Promise<T>, ms: number): Promise<T> {
  const ac = new AbortController();
  let timer: ReturnType<typeof setTimeout> | undefined;
  try {
    const timeout = new Promise<never>((_, reject) => {
      timer = setTimeout(() => {
        ac.abort();
        reject(new TimeoutError());
      }, ms);
    });
    return await Promise.race([op(ac.signal), timeout]);
  } finally {
    if (timer) clearTimeout(timer);
  }
}

export function createRetry<T>(opts: {
  max: number;
  baseMs: number;
  jitterPct: number; // e.g., 0.2 = ±20%
  shouldRetry: (e: unknown) => boolean;
  sleep: (ms: number) => Promise<void>;
}): (fn: (signal: AbortSignal) => Promise<T>) => Promise<T> {
  return async function retrying(fn: (signal: AbortSignal) => Promise<T>): Promise<T> {
    let attempt = 0;
    let lastErr: unknown;
    while (attempt < opts.max) {
      const ac = new AbortController();
      try {
        return await fn(ac.signal);
      } catch (e) {
        lastErr = e;
        attempt += 1;
        if (attempt >= opts.max || !opts.shouldRetry(e)) {
          throw e;
        }
        const base = opts.baseMs * Math.pow(2, attempt - 1);
        const jitter = base * opts.jitterPct;
        const delta = (Math.random() * 2 - 1) * jitter; // ±jitter
        const wait = Math.max(0, Math.floor(base + delta));
        await opts.sleep(wait);
      }
    }
    throw lastErr instanceof Error ? lastErr : new Error(String(lastErr));
  };
}
