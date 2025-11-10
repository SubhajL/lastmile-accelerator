import { describe, it, expect, vi } from 'vitest';
import { createTemplateLoader } from './loader.js';

function mockStore() {
  return {
    get: vi.fn(async (name: string) => name === 'from-store' ? { subject: 'S', html: '<p>H</p>', text: 'T' } : null)
  } as any;
}

describe('templates/loader', () => {
  it('loads from store with precedence over filesystem', async () => {
    const fs = {
      readFile: vi.fn(async () => { throw new Error('not used'); })
    } as any;
    const loader = createTemplateLoader({ store: mockStore(), fsBaseDir: 'src/templates', fs });

    const tpl = await loader.load('from-store');
    expect(tpl?.subject).toBe('S');
  });

  it('falls back to filesystem when store is missing', async () => {
    const store = { get: vi.fn(async () => null) } as any;
    const loader = createTemplateLoader({ store, fsBaseDir: 'src/templates' });
    const tpl = await loader.load('snapshot-ready');
    expect(tpl?.subject).toMatch(/Snapshot/);
    expect(tpl?.html).toContain('snapshot');
  });
});
