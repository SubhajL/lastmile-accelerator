export function createOtelMetrics({ meter, serviceName }: { meter: unknown; serviceName: string }) {
  const counters = new Map<string, { add: (n: number, attrs?: Record<string, unknown>) => void }>();

  function counter(name: string) {
    if (!counters.has(name)) {
      const m = meter as { createCounter: (n: string, opts?: { description?: string }) => { add: (n: number, attrs?: Record<string, unknown>) => void } };
      counters.set(name, m.createCounter(name, { description: `${serviceName}:${name}` }));
    }
    return counters.get(name)!;
  }

  return {
    increment(name: string, labels: Record<string, string | number> = {}) {
      counter(name).add(1, labels);
    }
  };
}
