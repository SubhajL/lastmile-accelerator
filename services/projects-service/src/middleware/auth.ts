import { AuthError } from '../utils/errors';
import { createRemoteJWKSet, jwtVerify, type JWTPayload } from 'jose';

// JWT verification config
export interface JwtConfig {
  jwksUrl?: string;
  issuer?: string;
  audience?: string;
  clockToleranceSec?: number;
}

// Pluggable verifier to keep implementation flexible & testable (used in unit tests)
let verifier: ((token: string, config: JwtConfig) => Promise<any>) | null = null;

export function __setJwtVerifier(fn: (token: string, config: JwtConfig) => Promise<any>) {
  verifier = fn;
}

export async function verifyJwt(token: string, config: JwtConfig): Promise<any> {
  if (!verifier) {
    throw new Error('JWT verifier not configured');
  }
  return verifier(token, config);
}

function createJoseVerifier(config: JwtConfig) {
  if (!config.jwksUrl) throw new Error('jwksUrl is required for jose verifier');
  const jwks = createRemoteJWKSet(new URL(config.jwksUrl));
  return async (token: string) => {
    const { payload } = await jwtVerify(token, jwks, {
      issuer: config.issuer,
      audience: config.audience,
      clockTolerance: config.clockToleranceSec ?? 5,
    });
    return payload as JWTPayload & Record<string, any>;
  };
}

export function hasAllScopes(user: any, required: string[]): boolean {
  const scopeStr: string | undefined = user?.scope;
  const scopes = new Set<string>((scopeStr ?? '').split(/\s+/).filter(Boolean));
  return required.every(s => scopes.has(s));
}

export function requireScopes(required: string | string[]) {
  const reqd = Array.isArray(required) ? required : [required];
  return async function (request: any) {
    if (!hasAllScopes(request?.user, reqd)) {
      throw new (await import('../utils/errors')).ForbiddenError('Forbidden: insufficient scope');
    }
  };
}

export function requireRoles(required: string | string[]) {
  const reqd = new Set(Array.isArray(required) ? required : [required]);
  return async function (request: any) {
    const role = request?.user?.role;
    if (!role || !reqd.has(role)) {
      throw new (await import('../utils/errors')).ForbiddenError('Forbidden: insufficient role');
    }
  };
}

export function requireOwner() {
  return requireRoles('owner');
}

export function authMiddleware(config: JwtConfig) {
  // prefer injected verifier for tests; otherwise use jose remote JWKS when configured
  const localVerify: ((token: string) => Promise<any>) = verifier
    ? (t: string) => verifyJwt(t, config)
    : (config.jwksUrl ? createJoseVerifier(config) : (() => { throw new Error('JWT verifier not configured'); }) as any);

  return async function (request: any, _reply: any) {
    const header: string | undefined = request?.headers?.authorization || request?.headers?.Authorization;
    if (!header || !header.startsWith('Bearer ')) {
      throw new AuthError('Unauthorized: missing bearer token');
    }
    const token = header.slice('Bearer '.length).trim();
    try {
      const payload = await localVerify(token);
      request.user = payload;
    } catch (err) {
      throw new AuthError('Unauthorized');
    }
  };
}

export function extractTenantId(request: any): string {
  const tenantId = request?.user?.tenantId || request?.user?.tenant_id;
  if (!tenantId) {
    throw new AuthError('Unauthorized: tenant not found in token');
  }
  return tenantId;
}
