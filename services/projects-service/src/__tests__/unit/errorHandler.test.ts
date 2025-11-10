import { describe, it, expect, vi } from 'vitest';
import { errorHandler } from '../../middleware/errorHandler';
import { ValidationError } from '../../utils/errors';

function createReply() {
  return {
    status: vi.fn().mockReturnThis(),
    send: vi.fn().mockReturnThis(),
  };
}

describe('middleware/errorHandler.ts', () => {
  it('formats ValidationError as 400 with code and message', async () => {
    const err = new ValidationError('Invalid input');
    const req: any = { id: 'req-1' };
    const reply = createReply();

    await errorHandler(err as any, req, reply as any);

    expect(reply.status).toHaveBeenCalledWith(400);
    const body = (reply.send as any).mock.calls[0][0];
    expect(body).toMatchObject({ code: 'VALIDATION_ERROR', message: 'Invalid input' });
  });

  it('formats generic Error as 500 with fallback code', async () => {
    const err = new Error('Kaboom');
    const req: any = { id: 'req-2' };
    const reply = createReply();

    await errorHandler(err as any, req, reply as any);

    expect(reply.status).toHaveBeenCalledWith(500);
    const body = (reply.send as any).mock.calls[0][0];
    expect(body).toMatchObject({ code: 'INTERNAL_SERVER_ERROR', message: 'Kaboom' });
  });
});