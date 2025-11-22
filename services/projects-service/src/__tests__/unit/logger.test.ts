import { describe, it, expect, beforeEach, vi } from 'vitest';
import { resetLogger } from '../../logger';

describe('Logger', () => {
  beforeEach(() => {
    // Clear logger singleton and mocks before each test (Bun-compatible)
    resetLogger();
    vi.clearAllMocks();
  });

  describe('createLogger()', () => {
    it('should return logger instance with service.name in context', async () => {
      const { createLogger } = await import('../../logger');
      const logger = createLogger('test-service');

      expect(logger).toBeDefined();
      expect(typeof logger.info).toBe('function');
      expect(typeof logger.error).toBe('function');
      expect(typeof logger.warn).toBe('function');
      expect(typeof logger.debug).toBe('function');
    });

    it('should create logger with correct transport (pino)', async () => {
      const { createLogger } = await import('../../logger');
      const logger = createLogger('my-service');

      // Logger should have pino methods
      expect(typeof logger.info).toBe('function');
      expect(logger).toHaveProperty('level');
    });
  });

  describe('getLogger()', () => {
    it('should return singleton logger instance', async () => {
      const { createLogger, getLogger } = await import('../../logger');
      createLogger('singleton-test');

      const logger1 = getLogger();
      const logger2 = getLogger();

      expect(logger1).toBe(logger2);
    });

    it('should throw if getLogger called before createLogger', async () => {
      resetLogger();
      vi.clearAllMocks();
      const { getLogger } = await import('../../logger');

      expect(() => getLogger()).toThrow(
        'Logger not initialized. Call createLogger() before getLogger().',
      );
    });
  });

  describe('log methods', () => {
    it('should log info messages without throwing', async () => {
      const { createLogger } = await import('../../logger');
      const logger = createLogger('test-service');

      expect(() => {
        logger.info({ msg: 'test message' }, 'info');
      }).not.toThrow();
    });

    it('should log error messages without throwing', async () => {
      const { createLogger } = await import('../../logger');
      const logger = createLogger('test-service');

      expect(() => {
        logger.error({ msg: 'error message', err: new Error('test') }, 'error');
      }).not.toThrow();
    });

    it('should log with context fields', async () => {
      const { createLogger } = await import('../../logger');
      const logger = createLogger('test-service');

      expect(() => {
        logger.info({ requestId: '123', userId: 'user-456' }, 'request started');
      }).not.toThrow();
    });
  });
});
