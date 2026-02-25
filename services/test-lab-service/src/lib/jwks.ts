import { createLocalJWKSet, decodeProtectedHeader, jwtVerify, type JWK, type JWTPayload as JoseJWTPayload } from 'jose';
import type { JWTPayload } from '../types/auth.js';

type FetchFn = (url: string) => Promise<{ ok: boolean; json: () => Promise<any> }>;

export interface JwksFetcherOptions {
  jwksUrl: string;
  cacheTtlMs: number;
  fetchFn?: FetchFn;
}

export interface JwksFetcher {
  getResolver: () => Promise<ReturnType<typeof createLocalJWKSet>>;
}

/**
 * Parses JWT header and returns alg and kid.
 */
export function parseHeaderAlg(token: string): { alg: string; kid?: string } {
  const h = decodeProtectedHeader(token);
  if (!h.alg) throw new Error('JWT missing alg header');
  return { alg: String(h.alg), kid: h.kid };
}

/**
 * Create a JWKS fetcher with simple in-memory TTL caching.
 */
export function createJwksFetcher(opts: JwksFetcherOptions): JwksFetcher {
  const fetchFn: FetchFn = opts.fetchFn ?? (globalThis.fetch as any);
  let cached: { jwks: { keys: JWK[] }; expiresAt: number } | null = null;

  async function load(): Promise<{ keys: JWK[] }> {
    const now = Date.now();
    if (cached && now < cached.expiresAt) return cached.jwks;
    const res = await fetchFn(opts.jwksUrl);
    if (!res || !(res as any).ok) throw new Error('JWKS fetch failed');
    const body = await res.json();
    if (!body || !Array.isArray(body.keys) || body.keys.length === 0) {
      throw new Error('JWKS response invalid');
    }
    cached = { jwks: { keys: body.keys }, expiresAt: now + opts.cacheTtlMs };
    return cached.jwks;
  }

  return {
    async getResolver() {
      const jwks = await load();
      return createLocalJWKSet(jwks as any);
    },
  };
}

export interface VerifyJwtInput {
  token: string;
  issuer: string;
  audience: string;
  alg: 'RS256' | 'RS384' | 'RS512' | 'ES256';
  clockSkewSec: number;
  jwksUrl: string;
  fetcher?: JwksFetcher;
}

/**
 * Verifies a JWT against JWKS with issuer/audience/alg checks and clock skew tolerance.
 */
export async function verifyJwt(input: VerifyJwtInput): Promise<JWTPayload> {
  const { alg: expectedAlg } = input;
  const header = parseHeaderAlg(input.token);
  if (header.alg !== expectedAlg) throw new Error(`JWT header algorithm mismatch: expected ${expectedAlg}, got ${header.alg}`);

  const fetcher = input.fetcher ?? createJwksFetcher({ jwksUrl: input.jwksUrl, cacheTtlMs: 600000 });
  const resolver = await fetcher.getResolver();

  const { payload } = await jwtVerify(input.token, resolver, {
    algorithms: [expectedAlg],
    issuer: input.issuer,
    audience: input.audience,
    clockTolerance: input.clockSkewSec,
  });

  return payload as unknown as JWTPayload;
}
