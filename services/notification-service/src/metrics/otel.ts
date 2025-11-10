export function createOtelMetrics({ meter, serviceName }: { meter: { createCounter: (name: string, opts?: any) => { add: (n: number, attrs?: Record<string, any>) => void } }, serviceName: string }) {
  const counters = new Map<string, { add: (n: number, attrs?: Record<string, any>) => void }>();

  function counter(name: string) {
    if (!counters.has(name)) {
      counters.set(name, meter.createCounter(name, { description: `${serviceName}:${name}` }));
    }
    return counters.get(name)!;
  }

  return {
    increment(name: string, labels: Record<string, string | number> = {}) {
      counter(name).add(1, labels);
    }
  };
}
