import type { FastifyInstance, FastifyRequest, FastifyReply, preHandlerHookHandler } from 'fastify';
import { AuthError, AuthorizationError } from '../types/auth.js';
import type { JWTPayload, UserContext } from '../types/auth.js';
import { getConfig } from '../config.js';
import { verifyJwt } from '../lib/jwks.js';

/**
 * Register lightweight auth decorations for Fastify.
 * Adds request.user if missing; verification is handled by requireAuth/optionalAuth via JWKS.
 *
 * @param app - Fastify application instance
 */
export async function registerAuthPlugin(app: FastifyInstance): Promise<void> {
  if (!app.hasRequestDecorator('user')) {
    app.decorateRequest('user', null);
  }
}

// Legacy fastify-jwt path is intentionally removed. Use requireAuth/optionalAuth.

export const requireAuth: preHandlerHookHandler = async (
  request: FastifyRequest,
  _reply: FastifyReply
) => {
  const token = extractBearerToken(request.headers.authorization);
  if (!token) {
    throw new AuthError('Missing or invalid Authorization header');
  }

  const cfg = getConfig();
  try {
    const claims = await verifyJwt({
      token,
      issuer: cfg.jwtIssuer,
      audience: cfg.jwtAudience,
      alg: cfg.jwtAlg,
      clockSkewSec: cfg.jwtClockSkewSec,
      jwksUrl: cfg.jwtJwksUrl,
    });
    request.user = toUserContext(claims);
  } catch {
    throw new AuthError('Invalid token');
  }
};

/**
 * Factory function that creates a preHandler hook to check for required OAuth scopes.
 * User must have ALL specified scopes to access the route.
 *
 * @param scopes - Required OAuth scopes
 * @returns PreHandler hook that validates scopes
 * @throws {AuthorizationError} If user lacks any required scope
 */
export function requireScopes(...scopes: string[]): preHandlerHookHandler {
  return async (request: FastifyRequest, reply: FastifyReply) => {
    const user = request.user as UserContext | undefined;

    if (!user) {
      throw new AuthError('User not authenticated');
    }

    const userScopes = user.scopes || [];
    const missingScopes = scopes.filter(scope => !userScopes.includes(scope));

    if (missingScopes.length > 0) {
      throw new AuthorizationError(
        `Missing required scopes: ${missingScopes.join(', ')}`,
        missingScopes
      );
    }
  };
}

/**
 * Extract bearer token from Authorization header.
 */
export function extractBearerToken(header?: string): string | null {
  if (!header) return null;
  if (!header.startsWith('Bearer ')) return null;
  const token = header.slice('Bearer '.length).trim();
  return token.length > 0 ? token : null;
}

/** Map JWT claims to UserContext. */
export function toUserContext(payload: JWTPayload): UserContext {
  return {
    sub: payload.sub,
    tenantId: payload.tenant_id,
    userId: payload.user_id,
    scopes: Array.isArray(payload.scopes) ? payload.scopes : [],
  };
}

/** @deprecated use requireAuth instead; will be removed in PR 4 (route-wiring) */
export const authenticateRequest = requireAuth;

/**
 * optionalAuth: does not fail requests. If a valid Bearer token is present,
 * verifies it and decorates request.user; otherwise proceeds.
 */
export const optionalAuth: preHandlerHookHandler = async (
  request: FastifyRequest,
  _reply: FastifyReply
) => {
  const header = request.headers.authorization;
  const token = extractBearerToken(header);
  if (!token) return; // proceed unauthenticated

  try {
    const cfg = getConfig();
    const claims = await verifyJwt({
      token,
      issuer: cfg.jwtIssuer,
      audience: cfg.jwtAudience,
      alg: cfg.jwtAlg,
      clockSkewSec: cfg.jwtClockSkewSec,
      jwksUrl: cfg.jwtJwksUrl,
    });
    request.user = toUserContext(claims);
  } catch {
    // swallow all errors: proceed unauthenticated
  }
};
