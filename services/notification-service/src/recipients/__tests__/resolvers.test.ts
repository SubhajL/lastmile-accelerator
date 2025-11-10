import { describe, it, expect } from 'vitest';
import { getRecipientEmailFromJob, createDefaultRecipientsResolver } from '../resolvers.js';

describe('recipients/resolvers', () => {
  it('getRecipientEmailFromJob prefers payload.email', async () => {
    const email = await getRecipientEmailFromJob({ payload: { email: 'a@b.com' }, userId: 'u1' } as any);
    expect(email).toBe('a@b.com');
  });

  it('getRecipientEmailFromJob uses payload.emails[0]', async () => {
    const email = await getRecipientEmailFromJob({ payload: { emails: ['x@y.com', 'z@w.com'] }, userId: 'u1' } as any);
    expect(email).toBe('x@y.com');
  });

  it('getRecipientEmailFromJob falls back to userId when email-like', async () => {
    const email = await getRecipientEmailFromJob({ payload: {}, userId: 'user@example.com' } as any);
    expect(email).toBe('user@example.com');
  });

  it('default resolver finds recipients on event', async () => {
    const r = createDefaultRecipientsResolver();
    expect(await r({ recipients: ['a@b.com'] } as any)).toEqual(['a@b.com']);
    expect(await r({ email: 'x@y.com' } as any)).toEqual(['x@y.com']);
    expect(await r({ emails: ['x@y.com', 'z@w.com'] } as any)).toEqual(['x@y.com', 'z@w.com']);
  });
});
