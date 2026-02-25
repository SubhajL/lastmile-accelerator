import type { FastifyRequest, FastifyReply } from 'fastify';
import type { RedisWrapper } from '../clients/redis.js';

export interface RateLimitConfig {
  requestsPerMinute: number;
  windowMs: number;
  exemptRoutes: string[];
  keyPrefix: string;
}

/**
 * Extract tenant identifier from request.
 * Uses JWT sub claim if available, otherwise falls back to IP address.
 */
function getTenantId(request: FastifyRequest): string {
  // Try to get tenant ID from JWT claims
  const user = (request as any).user;
  if (user && user.sub) {
    return user.sub as string;
  }

  // Fall back to IP address from x-forwarded-for header
  const forwarded = request.headers['x-forwarded-for'];
  if (forwarded) {
    // Handle both string and array formats
    const forwardedIp = Array.isArray(forwarded) ? forwarded[0] : forwarded;
    return forwardedIp.split(',')[0].trim();
  }

  return request.ip;
}

/**
 * Create rate limiting middleware using Redis sliding window counter.
 *
 * Algorithm: Sliding window counter
 * - Uses Redis INCR to atomically increment request count
 * - Sets TTL on first request in window
 * - Blocks requests when count exceeds limit
 *
 * Failure mode: Fail-open
 * - If Redis is unavailable, allows requests to maintain availability
 * - Logs errors for monitoring
 *
 * @param redis - Redis client wrapper
 * @param config - Rate limit configuration
 * @returns Fastify preHandler hook
 */
export function createRateLimitMiddleware(redis: RedisWrapper, config: RateLimitConfig) {
  return async function rateLimitHandler(request: FastifyRequest, reply: FastifyReply) {
    // Skip rate limiting for exempt routes (use route URL to ignore query strings)
    const routePath = (request as any).routeOptions?.url || request.url.split('?')[0];
    if (config.exemptRoutes.includes(routePath)) {
      return;
    }

    const tenantId = getTenantId(request);
    const windowSec = Math.ceil(config.windowMs / 1000);
    const rateLimitKey = `${config.keyPrefix}${tenantId}`;

    try {
      // Atomically increment counter
      const requestCount = await redis.incr(rateLimitKey);

      // Set TTL on first request
      if (requestCount === 1) {
        await redis.expire(rateLimitKey, windowSec);
      }

      // Check if limit exceeded
      if (requestCount > config.requestsPerMinute) {
        // Calculate retry-after in seconds
        const retryAfter = windowSec;

        reply
          .code(429)
          .header('Retry-After', String(retryAfter))
          .send({
            error: {
              code: 'RATE_LIMIT_EXCEEDED',
              message: `Too many requests. Limit: ${config.requestsPerMinute} requests per minute. Try again in ${retryAfter} seconds.`,
              traceId: request.id,
            },
          });

        return;
      }

      // Within limit - allow request to continue
    } catch (error) {
      // Redis error - fail open to maintain availability
      console.error(`Rate limit check failed for ${tenantId}:`, error);
      // Allow request despite error
    }
  };
}
