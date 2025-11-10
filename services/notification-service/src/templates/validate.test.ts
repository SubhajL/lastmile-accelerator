import { describe, it, expect } from 'vitest';
import { validateTemplateBundle } from './validate.js';

describe('templates/validate', () => {
  it('accepts subject+html and optional text', () => {
    expect(() => validateTemplateBundle({ subject: 's', html: '<p>x</p>' }, { allowMissingText: true })).not.toThrow();
    expect(() => validateTemplateBundle({ subject: 's', html: '<p>x</p>', text: 't' })).not.toThrow();
  });

  it('rejects missing subject or html', () => {
    expect(() => validateTemplateBundle({ subject: '', html: '<p>x</p>' } as any)).toThrow(/subject/);
    expect(() => validateTemplateBundle({ subject: 's' } as any)).toThrow(/html/);
  });
});
