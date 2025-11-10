type State = 'closed' | 'open' | 'half-open';

export function createCircuitBreaker(opts: { failureThreshold: number; halfOpenAfterMs: number; windowSize: number; now: () => number }) {
  let state: State = 'closed';
  let lastOpenedAt = 0;
  let failures = 0;

  async function execute<T>(fn: () => Promise<T>): Promise<T> {
    const now = opts.now();
    if (state === 'open') {
      if (now - lastOpenedAt >= opts.halfOpenAfterMs) {
        state = 'half-open';
      } else {
        throw new Error('circuit breaker open');
      }
    }

    try {
      const res = await fn();
      // success
      failures = 0;
      state = 'closed';
      return res;
    } catch (e) {
      failures += 1;
      if (failures >= opts.failureThreshold) {
        state = 'open';
        lastOpenedAt = now;
      }
      throw e;
    }
  }

  return { execute };
}
