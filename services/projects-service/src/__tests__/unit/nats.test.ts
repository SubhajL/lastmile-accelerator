import { describe, it, expect, beforeEach, vi } from 'vitest';
import { initNats, getNats, publish, closeNats, __setNatsConnector, __setSleeper } from '../../nats';

class MockNatsConnection {
  public published: Array<{ subject: string; data?: Uint8Array; options?: any }> = [];
  publish = vi.fn((subject: string, data?: Uint8Array, options?: any) => {
    this.published.push({ subject, data, options });
  });
  async drain() { /* noop */ }
  async close() { /* noop */ }
}

describe('nats.ts', () => {
  beforeEach(async () => {
    // Reset hooks between tests
    __setSleeper(async () => {});
    // Simulate singleton reset by closing if present
    await closeNats().catch(() => {});
  });

  it('initNats should create singleton connection with retries', async () => {
    const conn = new MockNatsConnection();
    const connector = vi.fn()
      .mockRejectedValueOnce(new Error('first fail'))
      .mockRejectedValueOnce(new Error('second fail'))
      .mockResolvedValue(conn);

    __setNatsConnector(connector as any);
    __setSleeper(async () => {}); // no real delay

    const nc = await initNats('n' + 'ats://localhost:4222', { maxRetries: 3, baseDelayMs: 1 });

    expect(nc).toBe(conn);
    expect(connector).toHaveBeenCalledTimes(3);

    // subsequent calls reuse singleton
    const nc2 = getNats();
    expect(nc2).toBe(nc);
  });

  it('publish should encode JSON and include optional traceparent header', async () => {
    const conn = new MockNatsConnection();
    const connector = vi.fn().mockResolvedValue(conn);
    __setNatsConnector(connector as any);

    await initNats('nats://demo:4222');

    await publish('projects.created', { id: 'p1', tenantId: 't1' }, '00-abc-xyz-01');

    expect(conn.publish).toHaveBeenCalledTimes(1);
    const call = conn.published[0];
    expect(call.subject).toBe('projects.created');
    expect(call.data).toBeInstanceOf(Uint8Array);

    // JSON payload check
    const decoded = new TextDecoder().decode(call.data);
    expect(decoded).toContain('"id":"p1"');
    expect(decoded).toContain('"tenantId":"t1"');

    // headers contain traceparent when provided
    expect(call.options?.headers?.traceparent).toBe('00-abc-xyz-01');
  });

  it('closeNats should drain/close and reset singleton', async () => {
    const conn = new MockNatsConnection();
    const drainSpy = vi.spyOn(conn, 'drain');
    const closeSpy = vi.spyOn(conn, 'close');
    __setNatsConnector(vi.fn().mockResolvedValue(conn) as any);

    await initNats('nats://demo:4222');
    await closeNats();

    expect(drainSpy).toHaveBeenCalled();
    expect(closeSpy).toHaveBeenCalled();

    // After close, getNats should throw
    expect(() => getNats()).toThrowError();
  });
});