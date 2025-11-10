import pino from 'pino';

export interface LoggerConfig {
  serviceName: string;
  env: 'dev' | 'staging' | 'prod';
  level?: string;
}

/**
 * Sensitive field names that should be redacted from logs
 */
const SENSITIVE_FIELDS = [
  'password',
  'token',
  'secret',
  'authorization',
  'apikey',
  'api_key',
  'accesstoken',
  'access_token',
  'refreshtoken',
  'refresh_token',
];

/**
 * Recursively redact sensitive data from an object
 */
export function redactSensitiveData(obj: Record<string, unknown>): Record<string, unknown> {
  const sanitized: Record<string, unknown> = {};
  
  for (const [key, value] of Object.entries(obj)) {
    const lowerKey = key.toLowerCase();
    
    if (SENSITIVE_FIELDS.some(field => lowerKey.includes(field))) {
      sanitized[key] = '[REDACTED]';
    } else if (typeof value === 'object' && value !== null && !Array.isArray(value)) {
      sanitized[key] = redactSensitiveData(value as Record<string, unknown>);
    } else {
      sanitized[key] = value;
    }
  }
  
  return sanitized;
}

/**
 * Create a structured logger with Pino.
 * Automatically configures log level based on environment and adds service context.
 * Redacts sensitive fields from logs.
 *
 * @param config - Logger configuration
 * @returns Configured Pino logger instance with redaction wrapper
 */
export function createLogger(config: LoggerConfig): pino.Logger {
  const { serviceName, env, level } = config;
  
  // Determine log level: debug in dev, info in staging/prod
  const logLevel = level || (env === 'dev' ? 'debug' : 'info');
  
  // Configure transport for pretty printing in dev
  const transport = env === 'dev' 
    ? { target: 'pino-pretty', options: { colorize: true } }
    : undefined;

  const baseLogger = pino({
    level: logLevel,
    base: {
      service: serviceName,
      env,
    },
    transport,
    serializers: {
      err: pino.stdSerializers.err,
      error: pino.stdSerializers.err,
    },
  });

  // Create a proxy to intercept logging calls
  const handler: ProxyHandler<pino.Logger> = {
    get(target, prop) {
      if (typeof prop === 'string' && ['trace', 'debug', 'info', 'warn', 'error', 'fatal'].includes(prop)) {
        const original = target[prop as keyof pino.Logger] as pino.LogFn;
        return function(this: pino.Logger, ...args: unknown[]) {
          if (args.length >= 1 && typeof args[0] === 'object' && args[0] !== null && !Array.isArray(args[0])) {
            args[0] = redactSensitiveData(args[0] as Record<string, unknown>);
          }
          // Cast to any to satisfy TypeScript - we know args are correct at runtime
          return (original as (...a: unknown[]) => void).apply(target, args);
        };
      }
      return target[prop as keyof pino.Logger];
    },
  };

  return new Proxy(baseLogger, handler);
}

/**
 * Create a child logger with additional context.
 * Useful for adding request-specific context (requestId, userId, projectId).
 *
 * @param parent - Parent logger instance
 * @param context - Additional context fields to include in all child logs
 * @returns Child logger with merged context
 */
export function childLogger(
  parent: pino.Logger,
  context: Record<string, unknown>
): pino.Logger {
  return parent.child(context);
}
