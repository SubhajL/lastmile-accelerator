import { describe, it, expect } from 'vitest';
import {
  ValidationError,
  AuthError,
  NotFoundError,
  ConflictError,
  isAppError,
} from '../../utils/errors';

describe('Error Classes', () => {
  describe('ValidationError', () => {
    it('should create error with 400 status and message', () => {
      const error = new ValidationError('Invalid input');
      expect(error.statusCode).toBe(400);
      expect(error.message).toBe('Invalid input');
      expect(error.code).toBe('VALIDATION_ERROR');
    });

    it('should be instanceof Error', () => {
      const error = new ValidationError('test');
      expect(error instanceof Error).toBe(true);
    });

    it('should include details if provided', () => {
      const error = new ValidationError('Invalid', { field: 'email' });
      expect(error.details).toEqual({ field: 'email' });
    });
  });

  describe('AuthError', () => {
    it('should create error with 401 status', () => {
      const error = new AuthError('Unauthorized');
      expect(error.statusCode).toBe(401);
      expect(error.code).toBe('AUTH_ERROR');
    });
  });

  describe('NotFoundError', () => {
    it('should create error with 404 status', () => {
      const error = new NotFoundError('Resource not found');
      expect(error.statusCode).toBe(404);
      expect(error.code).toBe('NOT_FOUND_ERROR');
    });
  });

  describe('ConflictError', () => {
    it('should create error with 409 status', () => {
      const error = new ConflictError('Resource already exists');
      expect(error.statusCode).toBe(409);
      expect(error.code).toBe('CONFLICT_ERROR');
    });
  });

  describe('isAppError()', () => {
    it('should return true for app errors', () => {
      const error = new ValidationError('test');
      expect(isAppError(error)).toBe(true);
    });

    it('should return false for regular errors', () => {
      const error = new Error('regular error');
      expect(isAppError(error)).toBe(false);
    });

    it('should return false for non-errors', () => {
      expect(isAppError({ statusCode: 400 })).toBe(false);
      expect(isAppError('string')).toBe(false);
      expect(isAppError(null)).toBe(false);
    });
  });
});
