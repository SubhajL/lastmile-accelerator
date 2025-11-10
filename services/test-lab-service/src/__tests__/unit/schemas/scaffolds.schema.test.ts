import { describe, it, expect } from 'vitest';
describe('schemas/scaffolds', () => {
  it('validates create payload', async () => {
const { CreateScaffoldSchema } = await import('../../../schemas/scaffolds.js');
    const data = {
      type: 'unit',
      framework: 'vitest',
      language: 'ts',
      config: { setupFiles: ['tests/setup.ts'] },
    };
    const parsed = CreateScaffoldSchema.parse(data);
    expect(parsed.type).toBe('unit');
  });

  it('rejects invalid type', async () => {
const { CreateScaffoldSchema } = await import('../../../schemas/scaffolds.js');
    const data = {
      type: 'weird',
      framework: 'vitest',
      language: 'ts',
      config: {},
    } as any;
    expect(() => CreateScaffoldSchema.parse(data)).toThrow();
  });

  it('validates update payload with partial fields', async () => {
const { UpdateScaffoldSchema } = await import('../../../schemas/scaffolds.js');
    const data = {
      framework: 'jest',
      config: { testMatch: ['**/*.spec.ts'] },
    };
    const parsed = UpdateScaffoldSchema.parse(data);
    expect(parsed.framework).toBe('jest');
  });

  it('validates list query pagination params', async () => {
const { ListScaffoldsQuerySchema } = await import('../../../schemas/scaffolds.js');
    const q = ListScaffoldsQuerySchema.parse({ limit: '25', cursor: 'abc' });
    expect(q.limit).toBe(25);
    expect(q.cursor).toBe('abc');
  });

  it('enforces limit bounds', async () => {
const { ListScaffoldsQuerySchema } = await import('../../../schemas/scaffolds.js');
    expect(() => ListScaffoldsQuerySchema.parse({ limit: '501' })).toThrow();
  });
});
