import { describe, it, expect, vi } from 'vitest';
import { requestIdMiddleware } from '../../middleware/requestId';

function makeReply() {
  return {
    header: vi.fn().mockReturnThis(),
  };
}

describe('middleware/requestId.ts', () => {
  it('uses incoming x-request-id when present', async () => {
    const req: any = { headers: { 'x-request-id': 'abc-123' } };
    const reply = makeReply();

    const mw = requestIdMiddleware();
    await mw(req, reply as any);

    expect(req.id).toBe('abc-123');
    expect(reply.header).toHaveBeenCalledWith('x-request-id', 'abc-123');
  });

  it('generates a request id when missing and sets response header', async () => {
    const req: any = { headers: {} };
    const reply = makeReply();

    const mw = requestIdMiddleware();
    await mw(req, reply as any);

    expect(typeof req.id).toBe('string');
    expect(req.id.length).toBeGreaterThan(8);
    expect(reply.header).toHaveBeenCalledWith('x-request-id', req.id);
  });
});