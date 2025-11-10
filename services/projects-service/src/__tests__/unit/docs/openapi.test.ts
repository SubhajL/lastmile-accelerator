import { describe, it, expect, vi } from 'vitest';
import fs from 'node:fs';
import path from 'node:path';

describe('docs/openapi.yaml basic structure', () => {
  it('contains openapi version and paths', () => {
const p = path.resolve(__dirname, '../../../..', 'docs/openapi.yaml');
    const s = fs.readFileSync(p, 'utf-8');
    expect(s).toContain('openapi: 3.1.0');
    expect(s).toContain('paths:');
    expect(s).toContain('components:');
    expect(s).toContain('securitySchemes:');
  });
});
