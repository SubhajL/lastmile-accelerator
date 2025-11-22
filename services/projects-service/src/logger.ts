import pino, { Logger } from 'pino';

/**
 * Singleton logger instance
 */
let loggerInstance: Logger | null = null;

/**
 * Create logger with service name and structured logging.
 * Configures pino for JSON output with request ID tracking.
 *
 * @param serviceName - Name of the service for tagging all logs
 * @returns Configured pino logger instance
 */
export function createLogger(serviceName: string): Logger {
  const isDev = process.env.NODE_ENV !== 'production';

  const logger = pino({
    level: process.env.LOG_LEVEL || (isDev ? 'debug' : 'info'),
    base: {
      service: {
        name: serviceName,
      },
    },
    transport: isDev
      ? {
          target: 'pino-pretty',
          options: {
            colorize: true,
            translateTime: 'SYS:standard',
            ignore: 'pid,hostname',
          },
        }
      : undefined,
  });

  loggerInstance = logger;
  return logger;
}

/**
 * Get singleton logger instance.
 * Must call createLogger() first.
 *
 * @throws Error if logger not initialized
 * @returns Configured logger instance
 */
export function getLogger(): Logger {
  if (!loggerInstance) {
    throw new Error('Logger not initialized. Call createLogger() before getLogger().');
  }
  return loggerInstance;
}

/**
 * Child logger for request context.
 * Inherits service.name from parent and adds custom fields.
 *
 * @param fields - Additional context fields (requestId, userId, etc.)
 * @returns Child logger with context fields
 */
export function getContextLogger(fields: Record<string, unknown>): Logger {
  return getLogger().child(fields);
}

/**
 * Reset logger singleton state.
 * FOR TESTING ONLY - clears the singleton instance to allow fresh logger creation.
 *
 * **IMPORTANT**: Only call from test code. Guarded by NODE_ENV to prevent production misuse.
 *
 * @internal
 * @throws Error if called in non-test environment
 */
export function resetLogger(): void {
  if (process.env.NODE_ENV !== 'test' && process.env.NODE_ENV !== undefined) {
    throw new Error('resetLogger() can only be called in test environment');
  }
  // Note: Pino loggers with transports are auto-closed on process exit.
  // Manual cleanup not needed for test isolation use case.
  loggerInstance = null;
}
