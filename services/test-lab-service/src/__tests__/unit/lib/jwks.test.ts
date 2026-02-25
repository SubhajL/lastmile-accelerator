import { describe, it, expect, beforeEach, vi } from 'vitest';
import { generateKeyPair, exportJWK, SignJWT } from 'jose';

const iss = 'https://auth.example.com/';
const aud = 'test-lab-service';

async function makeKeyPair() {
  const { publicKey, privateKey } = await generateKeyPair('RS256');
  const jwk = await exportJWK(publicKey);
  (jwk as any).kid = 'kid1';
  (jwk as any).use = 'sig';
  (jwk as any).alg = 'RS256';
  return { publicKey, privateKey, jwk } as const;
}

async function makeToken(privateKey: any, claims?: Partial<Record<string, any>>) {
  const now = Math.floor(Date.now() / 1000);
  const jwt = await new SignJWT({
    sub: 'user-1',
    tenant_id: 'tenant-1',
    user_id: 'user-1',
    scopes: ['scaffold:read'],
    ...claims,
  })
    .setProtectedHeader({ alg: 'RS256', kid: 'kid1', typ: 'JWT' })
    .setIssuer(claims?.iss ?? iss)
    .setAudience(claims?.aud ?? aud)
    .setIssuedAt(now)
    .setExpirationTime((claims?.exp as number) ?? now + 300)
    .sign(privateKey);
  return jwt;
}

describe('jwks verifier util', () => {
  let jwks: any;
  let token: string;
  let fetchFn: any;
  let privateKey: any;

  beforeEach(async () => {
    const kp = await makeKeyPair();
    privateKey = kp.privateKey;
    jwks = { keys: [kp.jwk] };
    token = await makeToken(privateKey);

    fetchFn = vi.fn(async () => ({ ok: true, json: async () => jwks }));
  });

  it('verifyJwt accepts valid token (signature/claims)', async () => {
    const { verifyJwt, createJwksFetcher } = await import('../../../lib/jwks.js');
    const fetcher = createJwksFetcher({ jwksUrl: 'https://jwks.local', cacheTtlMs: 600000, fetchFn });
    const res = await verifyJwt({
      token,
      issuer: iss,
      audience: aud,
      alg: 'RS256',
      clockSkewSec: 60,
      jwksUrl: 'https://jwks.local',
      fetcher,
    });
    expect(res.sub).toBe('user-1');
    expect(fetchFn).toHaveBeenCalledTimes(1);
  });

  it('verifyJwt rejects wrong issuer', async () => {
    const { verifyJwt, createJwksFetcher } = await import('../../../lib/jwks.js');
    const fetcher = createJwksFetcher({ jwksUrl: 'https://jwks.local', cacheTtlMs: 600000, fetchFn });
    await expect(
      verifyJwt({ token, issuer: 'https://wrong/', audience: aud, alg: 'RS256', clockSkewSec: 0, jwksUrl: 'https://jwks.local', fetcher })
    ).rejects.toThrow();
  });

  it('verifyJwt rejects wrong audience', async () => {
    const { verifyJwt, createJwksFetcher } = await import('../../../lib/jwks.js');
    const fetcher = createJwksFetcher({ jwksUrl: 'https://jwks.local', cacheTtlMs: 600000, fetchFn });
    await expect(
      verifyJwt({ token, issuer: iss, audience: 'another', alg: 'RS256', clockSkewSec: 0, jwksUrl: 'https://jwks.local', fetcher })
    ).rejects.toThrow();
  });

  it('verifyJwt rejects expired token respecting clock skew', async () => {
    const { verifyJwt, createJwksFetcher } = await import('../../../lib/jwks.js');
    const expired = await makeToken(privateKey, { exp: Math.floor(Date.now() / 1000) - 10 });
    const fetcher = createJwksFetcher({ jwksUrl: 'https://jwks.local', cacheTtlMs: 600000, fetchFn });
    await expect(
      verifyJwt({ token: expired, issuer: iss, audience: aud, alg: 'RS256', clockSkewSec: 0, jwksUrl: 'https://jwks.local', fetcher })
    ).rejects.toThrow();
  });

  it('verifyJwt rejects wrong alg (header mismatch)', async () => {
    const { verifyJwt, createJwksFetcher } = await import('../../../lib/jwks.js');
    const fetcher = createJwksFetcher({ jwksUrl: 'https://jwks.local', cacheTtlMs: 600000, fetchFn });
    await expect(
      verifyJwt({ token, issuer: iss, audience: aud, alg: 'RS512', clockSkewSec: 0, jwksUrl: 'https://jwks.local', fetcher })
    ).rejects.toThrow(/algorithm/);
  });

  it('createJwksFetcher caches keys and refetches after TTL', async () => {
    vi.useFakeTimers();
    const { verifyJwt, createJwksFetcher } = await import('../../../lib/jwks.js');
    const fetcher = createJwksFetcher({ jwksUrl: 'https://jwks.local', cacheTtlMs: 10, fetchFn });
    await verifyJwt({ token, issuer: iss, audience: aud, alg: 'RS256', clockSkewSec: 0, jwksUrl: 'https://jwks.local', fetcher });
    expect(fetchFn).toHaveBeenCalledTimes(1);

    // Advance time to exceed TTL
    const now = Date.now();
    vi.setSystemTime(new Date(now + 20));
    await verifyJwt({ token, issuer: iss, audience: aud, alg: 'RS256', clockSkewSec: 0, jwksUrl: 'https://jwks.local', fetcher });
    expect(fetchFn).toHaveBeenCalledTimes(2);
    vi.useRealTimers();
  });
});
