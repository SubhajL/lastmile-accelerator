import { describe, it, expect, beforeEach, vi } from 'vitest';
import { verifyJwt, authMiddleware, extractTenantId, __setJwtVerifier } from '../../middleware/auth';
import { AuthError } from '../../utils/errors';

describe('middleware/auth.ts', () => {
  beforeEach(() => {
    // Default verifier: token === 'valid' -> payload, else throw
    __setJwtVerifier(async (token: string) => {
      if (token === 'valid') return { sub: 'user-1', tenantId: 'tenant-1' };
      throw new Error('invalid');
    });
  });

  it('verifyJwt returns payload on success', async () => {
    const payload = await verifyJwt('valid', {} as any);
    expect(payload).toMatchObject({ sub: 'user-1', tenantId: 'tenant-1' });
  });

  it('authMiddleware attaches user to request for valid bearer token', async () => {
    const req: any = { headers: { authorization: 'Bearer valid' } };
    const reply: any = {};
    const mw = authMiddleware({});

    await mw(req, reply);
    expect(req.user).toMatchObject({ sub: 'user-1', tenantId: 'tenant-1' });
  });

  it('authMiddleware rejects missing/invalid token with AuthError', async () => {
    const req: any = { headers: { authorization: 'Bearer nope' } };
    const mw = authMiddleware({});

    await expect(mw(req, {} as any)).rejects.toBeInstanceOf(AuthError);
  });

  it('extractTenantId returns tenantId or throws AuthError if missing', () => {
    const req: any = { user: { tenantId: 't-1' } };
    expect(extractTenantId(req)).toBe('t-1');

    const req2: any = { user: { } };
    expect(() => extractTenantId(req2)).toThrow(AuthError);
  });
});