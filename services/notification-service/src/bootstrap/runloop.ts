export interface RunLoopOptions {
  dispatcher: { processNextBatch: (n: number) => Promise<unknown> };
  batchSize: number;
  tickMs: number;
  logger?: { error?: (...args: unknown[]) => void };
}

export function createRunLoop(opts: RunLoopOptions) {
  let timer: ReturnType<typeof setInterval> | undefined;
  let running = false;

  async function tick() {
    if (running) return;
    running = true;
    try {
      await opts.dispatcher.processNextBatch(opts.batchSize);
    } catch (e) {
      try {
        opts.logger?.error?.('dispatcher tick error', e);
      } catch { /* noop */ }
    } finally {
      running = false;
    }
  }

  function start() {
    if (timer) return;
    timer = setInterval(() => { void tick(); }, opts.tickMs);
  }

  function stop() {
    if (timer) clearInterval(timer);
    timer = undefined;
  }

  return { start, stop };
}
