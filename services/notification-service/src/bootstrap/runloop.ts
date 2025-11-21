export interface RunLoopOptions {
  dispatcher: { processNextBatch: (n: number) => Promise<unknown> };
  batchSize: number;
  tickMs: number;
  logger?: { error?: (...args: unknown[]) => void };
}

export interface LoopHandle {
  stop: () => Promise<void>;
}

export function createRunLoop(opts: RunLoopOptions) {
  let timer: ReturnType<typeof setInterval> | undefined;
  let running = false;
  let currentTick: Promise<void> | undefined;

  async function tick() {
    if (running) return;
    running = true;
    try {
      await opts.dispatcher.processNextBatch(opts.batchSize);
    } catch (e) {
      try {
        opts.logger?.error?.('dispatcher tick error', e);
      } catch {
        /* noop */
      }
    } finally {
      running = false;
      currentTick = undefined;
    }
  }

  function start() {
    if (timer) return;
    timer = setInterval(() => {
      currentTick = tick();
    }, opts.tickMs);
  }

  async function stop(): Promise<void> {
    if (timer) clearInterval(timer);
    timer = undefined;
    if (currentTick) await currentTick;
  }

  return { start, stop };
}

export function startDispatcherLoop(
  dispatcher: { processNextBatch: (n: number) => Promise<unknown> },
  batchSize: number,
  intervalMs: number = 100,
): LoopHandle {
  const runLoop = createRunLoop({
    dispatcher,
    batchSize,
    tickMs: intervalMs,
    logger: console,
  });

  runLoop.start();

  return {
    stop: async () => {
      await runLoop.stop();
    },
  };
}
