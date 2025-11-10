/**
 * JWT token payload structure from auth service.
 */
export interface JWTPayload {
  sub: string;           // Subject (user identifier)
  tenant_id: string;     // Tenant identifier for multi-tenancy
  user_id: string;       // User identifier
  scopes: string[];      // OAuth scopes for authorization
  exp: number;           // Expiration timestamp (seconds since epoch)
  iat: number;           // Issued at timestamp
  iss: string;           // Issuer
  aud?: string;          // Audience (optional)
}

/**
 * User context extracted from JWT and attached to request.
 */
export interface UserContext {
  sub: string;
  tenantId: string;
  userId: string;
  scopes: string[];
}

/**
 * Project context for tenant access validation.
 */
export interface ProjectContext {
  projectId: string;
  tenantId: string;
}

/**
 * Base authentication error.
 * Returns 401 Unauthorized.
 */
export class AuthError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'AuthError';
  }
}

/**
 * Authorization error for insufficient permissions.
 * Returns 403 Forbidden.
 */
export class AuthorizationError extends Error {
  public readonly missingScopes?: string[];

  constructor(message: string, missingScopes?: string[]) {
    super(message);
    this.name = 'AuthorizationError';
    this.missingScopes = missingScopes;
  }
}

/**
 * Validation error for invalid request data.
 * Returns 400 Bad Request.
 */
export class ValidationError extends Error {
  public readonly field?: string;

  constructor(message: string, field?: string) {
    super(message);
    this.name = 'ValidationError';
    this.field = field;
  }
}
