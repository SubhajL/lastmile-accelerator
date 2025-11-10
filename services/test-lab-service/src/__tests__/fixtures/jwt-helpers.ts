import { sign } from 'jsonwebtoken';
import type { JWTPayload } from '../../types/auth.js';

// Test secret key for signing JWTs in tests
export const TEST_JWT_SECRET = 'test-secret-key-for-testing-only';

/**
 * Creates a valid mock JWT for testing.
 * 
 * @param customClaims - Optional custom claims to override defaults
 * @returns Signed JWT string
 */
export function createMockJWT(customClaims: Partial<JWTPayload> = {}): string {
  const now = Math.floor(Date.now() / 1000);
  
  const defaultClaims: JWTPayload = {
    sub: 'user-123',
    tenant_id: 'tenant-abc',
    user_id: 'user-123',
    scopes: ['test:read', 'test:write'],
    exp: now + 3600, // Expires in 1 hour
    iat: now,
    iss: 'https://auth.test-lab.com',
  };

  const claims = { ...defaultClaims, ...customClaims };

  return sign(claims, TEST_JWT_SECRET, { algorithm: 'HS256' });
}

/**
 * Creates an expired JWT for testing token expiration.
 * 
 * @param customClaims - Optional custom claims
 * @returns Signed JWT with past expiration
 */
export function createExpiredJWT(customClaims: Partial<JWTPayload> = {}): string {
  const now = Math.floor(Date.now() / 1000);
  
  const claims: JWTPayload = {
    sub: 'user-123',
    tenant_id: 'tenant-abc',
    user_id: 'user-123',
    scopes: ['test:read'],
    exp: now - 3600, // Expired 1 hour ago
    iat: now - 7200,
    iss: 'https://auth.test-lab.com',
    ...customClaims,
  };

  return sign(claims, TEST_JWT_SECRET, { algorithm: 'HS256' });
}

/**
 * Creates a JWT with invalid signature for testing signature verification.
 * 
 * @param customClaims - Optional custom claims
 * @returns JWT signed with wrong secret
 */
export function createInvalidSignatureJWT(customClaims: Partial<JWTPayload> = {}): string {
  const now = Math.floor(Date.now() / 1000);
  
  const claims: JWTPayload = {
    sub: 'user-123',
    tenant_id: 'tenant-abc',
    user_id: 'user-123',
    scopes: ['test:read'],
    exp: now + 3600,
    iat: now,
    iss: 'https://auth.test-lab.com',
    ...customClaims,
  };

  // Sign with different secret to create invalid signature
  return sign(claims, 'wrong-secret-key', { algorithm: 'HS256' });
}
