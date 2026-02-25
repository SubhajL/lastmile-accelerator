import { createLocalJWKSet, decodeProtectedHeader, jwtVerify, type JSONWebKeySet } from 'jose';
import { z } from 'zod';
import type { JWTPayload } from '../types/auth.js';

type FetchFn = (url: string) => Promise<{ ok: boolean; json: () => Promise<unknown> }>;

const DEFAULT_JWKS_CACHE_TTL_MS = 10 * 60 * 1000;

const jwtPayloadSchema = z.object({
  sub: z.string().min(1),
  tenant_id: z.string().min(1),
  user_id: z.string().min(1),
  scopes: z.array(z.string()).optional().default([]),
  exp: z.number(),
  iat: z.number(),
  iss: z.string().min(1),
  aud: z.union([z.string(), z.array(z.string())]).optional(),
});

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
  const fetchFn: FetchFn = opts.fetchFn ?? defaultFetchFn;
  let cached: { jwks: JSONWebKeySet; expiresAt: number } | null = null;

  function isJsonWebKeySet(body: unknown): body is JSONWebKeySet {
    if (typeof body !== 'object' || body === null) return false;
    const keys = (body as { keys?: unknown }).keys;
    return Array.isArray(keys);
  }

  async function load(): Promise<JSONWebKeySet> {
    const now = Date.now();
    if (cached && now < cached.expiresAt) return cached.jwks;
    const res = await fetchFn(opts.jwksUrl);
    if (!res.ok) throw new Error('JWKS fetch failed');
    const body = await res.json();
    if (!isJsonWebKeySet(body) || body.keys.length === 0) {
      throw new Error('JWKS response invalid');
    }
    cached = { jwks: body, expiresAt: now + opts.cacheTtlMs };
    return cached.jwks;
  }

  return {
    async getResolver() {
      const jwks = await load();
      return createLocalJWKSet(jwks);
    },
  };
}

const defaultFetchers = new Map<string, JwksFetcher>();

function defaultFetchFn(url: string) {
  const fetch = globalThis.fetch;
  if (typeof fetch !== 'function') {
    throw new Error('fetch is not available in this runtime');
  }
  return fetch(url).then((res) => ({
    ok: res.ok,
    json: async () => res.json() as unknown,
  }));
}

function getDefaultFetcher(jwksUrl: string) {
  const existing = defaultFetchers.get(jwksUrl);
  if (existing) return existing;
  const next = createJwksFetcher({ jwksUrl, cacheTtlMs: DEFAULT_JWKS_CACHE_TTL_MS });
  defaultFetchers.set(jwksUrl, next);
  return next;
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

  const fetcher = input.fetcher ?? getDefaultFetcher(input.jwksUrl);
  const resolver = await fetcher.getResolver();

  const { payload } = await jwtVerify(input.token, resolver, {
    algorithms: [expectedAlg],
    issuer: input.issuer,
    audience: input.audience,
    clockTolerance: input.clockSkewSec,
  });

  const parsed = jwtPayloadSchema.parse(payload);
  const aud = Array.isArray(parsed.aud) ? parsed.aud[0] : parsed.aud;
  return { ...parsed, aud };
}
