import fastifyJwt from '@fastify/jwt';
import type { FastifyInstance, FastifyRequest, FastifyReply, preHandlerHookHandler } from 'fastify';
import { AuthError, AuthorizationError } from '../types/auth.js';
import type { JWTPayload, UserContext } from '../types/auth.js';
import { getConfig } from '../config.js';
import { verifyJwt } from '../lib/jwks.js';

/**
 * Registers JWT authentication plugin with Fastify.
 * Configures JWT verification using secret or JWKS URL.
 *
 * @param app - Fastify application instance
 * @param secretOrJwksUrl - JWT secret (for testing) or JWKS URL (for production)
 */
export async function registerAuthPlugin(
  app: FastifyInstance,
  secretOrJwksUrl: string
): Promise<void> {
  await app.register(fastifyJwt, {
    secret: secretOrJwksUrl,
  });

  // Decorate request with user property (check if not already decorated)
  if (!app.hasRequestDecorator('user')) {
    app.decorateRequest('user', null);
  }
}

/**
 * PreHandler hook that authenticates requests via JWT.
 * Validates Bearer token from Authorization header and populates request.user.
 *
 * @throws {AuthError} If token is missing, invalid, or expired
 */
export const authenticateRequest: preHandlerHookHandler = async (
  request: FastifyRequest,
  reply: FastifyReply
) => {
  try {
    // Extract authorization header
    const authHeader = request.headers.authorization;

    if (!authHeader) {
      throw new AuthError('No Authorization header provided');
    }

    if (!authHeader.startsWith('Bearer ')) {
      throw new AuthError('Invalid Authorization header format. Expected: Bearer <token>');
    }

    // Verify JWT
    await request.jwtVerify();

    // Extract claims and populate user context
    const payload = request.user as unknown as JWTPayload;

    request.user = {
      sub: payload.sub,
      tenantId: payload.tenant_id,
      userId: payload.user_id,
      scopes: payload.scopes || [],
    };
  } catch (error) {
    if (error instanceof AuthError) {
      throw error;
    }

    // Handle JWT verification errors
    if (error instanceof Error) {
      if (error.message.includes('expired')) {
        throw new AuthError('Token has expired');
      }
      if (error.message.includes('signature')) {
        throw new AuthError('Invalid token signature');
      }
      throw new AuthError('Invalid token');
    }

    throw new AuthError('Authentication failed');
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
