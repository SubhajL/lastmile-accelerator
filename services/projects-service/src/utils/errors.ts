/**
 * Base application error class.
 * All custom errors extend this with statusCode, code, and optional details.
 */
export class AppError extends Error {
  constructor(
    message: string,
    public readonly statusCode: number,
    public readonly code: string,
    public readonly details?: Record<string, unknown>
  ) {
    super(message);
    Object.setPrototypeOf(this, AppError.prototype);
  }
}

/**
 * Validation error (400).
 * Used for invalid request input, schema validation failures.
 */
export class ValidationError extends AppError {
  constructor(message: string, details?: Record<string, unknown>) {
    super(message, 400, 'VALIDATION_ERROR', details);
    Object.setPrototypeOf(this, ValidationError.prototype);
  }
}

/**
 * Authentication error (401).
 * Used for invalid/missing JWT, expired tokens, etc.
 */
export class AuthError extends AppError {
  constructor(message: string, details?: Record<string, unknown>) {
    super(message, 401, 'AUTH_ERROR', details);
    Object.setPrototypeOf(this, AuthError.prototype);
  }
}

/**
 * Not found error (404).
 * Used when resource doesn't exist.
 */
export class NotFoundError extends AppError {
  constructor(message: string, details?: Record<string, unknown>) {
    super(message, 404, 'NOT_FOUND_ERROR', details);
    Object.setPrototypeOf(this, NotFoundError.prototype);
  }
}

/**
 * Forbidden error (403).
 * Used when authenticated user lacks permissions/scopes.
 */
export class ForbiddenError extends AppError {
  constructor(message: string, details?: Record<string, unknown>) {
    super(message, 403, 'FORBIDDEN', details);
    Object.setPrototypeOf(this, ForbiddenError.prototype);
  }
}

/**
 * Conflict error (409).
 * Used when request conflicts with current state (e.g., duplicate email).
 */
export class ConflictError extends AppError {
  constructor(message: string, details?: Record<string, unknown>) {
    super(message, 409, 'CONFLICT_ERROR', details);
    Object.setPrototypeOf(this, ConflictError.prototype);
  }
}

/**
 * Type guard to check if an error is an AppError.
 * Useful for error handling middleware.
 *
 * @param error - Value to check
 * @returns true if error is an AppError with statusCode property
 */
export function isAppError(error: unknown): error is AppError {
  return (
    error instanceof Error &&
    'statusCode' in error &&
    typeof (error as Record<string, unknown>).statusCode === 'number'
  );
}
