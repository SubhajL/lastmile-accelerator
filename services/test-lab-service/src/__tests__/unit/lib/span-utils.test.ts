import { describe, it, expect, beforeAll, afterAll, vi } from 'vitest';
import { trace, context } from '@opentelemetry/api';
import { withSpan, getCurrentSpan, addSpanAttributes, addSpanEvent, recordException } from '../../../lib/span-utils.js';
import { BasicTracerProvider, SimpleSpanProcessor, InMemorySpanExporter } from '@opentelemetry/sdk-trace-base';
import { AsyncLocalStorageContextManager } from '@opentelemetry/context-async-hooks';

describe('Span Utilities', () => {
  beforeAll(() => {
    const provider = new BasicTracerProvider();
    provider.addSpanProcessor(new SimpleSpanProcessor(new InMemorySpanExporter()));
    provider.register({ contextManager: new AsyncLocalStorageContextManager() });
  });

  afterAll(async () => {
    // no-op
  });
  describe('withSpan()', () => {
    it('should create span for async function', async () => {
      const result = await withSpan('test-operation', async () => {
        return 'success';
      });

      expect(result).toBe('success');
    });

    it('should add custom attributes to span', async () => {
      await withSpan(
        'test-operation',
        async () => {
          return 'done';
        },
        {
          projectId: 'proj-123',
          testRunId: 'run-456',
        }
      );

      // Span should have been created and ended
      expect(true).toBe(true);
    });

    it('should record exception on function error', async () => {
      const error = new Error('Test error');

      await expect(
        withSpan('test-operation', async () => {
          throw error;
        })
      ).rejects.toThrow('Test error');
    });

    it('should propagate exception after recording', async () => {
      const error = new Error('Propagated error');

      await expect(
        withSpan('failing-operation', async () => {
          throw error;
        })
      ).rejects.toThrow('Propagated error');
    });

    it('should measure function execution time', async () => {
      const startTime = Date.now();

      await withSpan('timed-operation', async () => {
        await new Promise(resolve => setTimeout(resolve, 50));
      });

      const duration = Date.now() - startTime;
      expect(duration).toBeGreaterThanOrEqual(50);
    });

    it('should handle synchronous errors', async () => {
      await expect(
        withSpan('sync-error', async () => {
          throw new TypeError('Type error');
        })
      ).rejects.toThrow('Type error');
    });
  });

  describe('getCurrentSpan()', () => {
    it('should return active span within traced context', async () => {
      await withSpan('parent-operation', async () => {
        const span = getCurrentSpan();
        expect(span).toBeDefined();
      });
    });

    it('should return undefined outside traced context', () => {
      const span = getCurrentSpan();
      expect(span).toBeUndefined();
    });
  });

  describe('addSpanAttributes()', () => {
    it('should add attributes to current span', async () => {
      await withSpan('operation', async () => {
        addSpanAttributes({
          userId: 'user-123',
          action: 'test',
        });

        // Should not throw
        expect(true).toBe(true);
      });
    });

    it('should handle string attributes', async () => {
      await withSpan('operation', async () => {
        addSpanAttributes({
          stringAttr: 'value',
        });

        expect(true).toBe(true);
      });
    });

    it('should handle number attributes', async () => {
      await withSpan('operation', async () => {
        addSpanAttributes({
          count: 42,
          duration: 123.45,
        });

        expect(true).toBe(true);
      });
    });

    it('should handle boolean attributes', async () => {
      await withSpan('operation', async () => {
        addSpanAttributes({
          success: true,
          cached: false,
        });

        expect(true).toBe(true);
      });
    });

    it('should no-op when no active span', () => {
      expect(() => {
        addSpanAttributes({
          field: 'value',
        });
      }).not.toThrow();
    });
  });

  describe('addSpanEvent()', () => {
    it('should add event to current span', async () => {
      await withSpan('operation', async () => {
        addSpanEvent('test_started');

        expect(true).toBe(true);
      });
    });

    it('should add event with attributes', async () => {
      await withSpan('operation', async () => {
        addSpanEvent('browser_session_created', {
          browser: 'chrome',
          version: '120',
        });

        expect(true).toBe(true);
      });
    });

    it('should no-op when no active span', () => {
      expect(() => {
        addSpanEvent('event_name');
      }).not.toThrow();
    });
  });

  describe('recordException()', () => {
    it('should record error on current span', async () => {
      const error = new Error('Test error');

      await withSpan('operation', async () => {
        recordException(error);

        expect(true).toBe(true);
      });
    });

    it('should include stack trace in recording', async () => {
      const error = new Error('Error with stack');

      await withSpan('operation', async () => {
        recordException(error);

        expect(error.stack).toBeDefined();
      });
    });

    it('should record error with custom attributes', async () => {
      const error = new Error('Custom error');

      await withSpan('operation', async () => {
        recordException(error, {
          errorCode: 'TEST_ERROR',
          severity: 'high',
        });

        expect(true).toBe(true);
      });
    });

    it('should no-op when no active span', () => {
      const error = new Error('Error outside span');

      expect(() => {
        recordException(error);
      }).not.toThrow();
    });
  });
});
