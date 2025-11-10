import { describe, it, expect, vi } from 'vitest';
import { createOtelMetrics } from '../otel.js';

describe('metrics/otel', () => {
  it('creates counters lazily and increments with labels', () => {
    const add = vi.fn();
    const counter = { add };
    const meter = { createCounter: vi.fn().mockReturnValue(counter) } as any;
    const m = createOtelMetrics({ meter, serviceName: 'notification-service' });

    m.increment('notify_sent', { channel: 'email' });
    m.increment('notify_failed', { channel: 'email', reason: 'adapter_error' });

    expect(meter.createCounter).toHaveBeenCalledTimes(2);
    expect(add).toHaveBeenCalled();
  });
});
