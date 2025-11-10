import type { FastifyInstance, FastifyError, FastifyRequest, FastifyReply } from 'fastify';
import { AuthError, AuthorizationError, ValidationError } from '../types/auth.js';

/**
 * Registers global error handler for the Fastify application.
 * Maps custom errors to appropriate HTTP status codes.
 *
 * @param app - Fastify application instance
 */
export function registerErrorHandler(app: FastifyInstance): void {
  app.setErrorHandler((error: FastifyError | Error, request: FastifyRequest, reply: FastifyReply) => {
    // Log error with request context
    request.log.error({
      err: error,
      url: request.url,
      method: request.method,
    }, 'Request error');

    // Handle authentication errors
    if (error instanceof AuthError) {
      return reply
        .status(401)
        .header('WWW-Authenticate', 'Bearer')
        .send({
          error: error.message,
          statusCode: 401,
        });
    }

    // Handle authorization errors
    if (error instanceof AuthorizationError) {
      const response: {
        error: string;
        statusCode: number;
        missingScopes?: string[];
      } = {
        error: error.message,
        statusCode: 403,
      };

      if (error.missingScopes && error.missingScopes.length > 0) {
        response.missingScopes = error.missingScopes;
      }

      return reply.status(403).send(response);
    }

    // Handle validation errors
    if (error instanceof ValidationError) {
      const response: {
        error: string;
        statusCode: number;
        field?: string;
      } = {
        error: error.message,
        statusCode: 400,
      };

      if (error.field) {
        response.field = error.field;
      }

      return reply.status(400).send(response);
    }

    // Handle Fastify validation errors
    if ('validation' in error) {
      return reply.status(400).send({
        error: 'Validation failed',
        statusCode: 400,
        details: error.validation,
      });
    }

    // Generic error handling - don't leak internal details
    return reply.status(500).send({
      error: 'Internal Server Error',
      statusCode: 500,
    });
  });
}
