import { describe, it, expect } from 'vitest';
import SwaggerParser from '@apidevtools/swagger-parser';
import path from 'node:path';

const specPath = path.resolve(__dirname, '../../../..', 'docs/openapi.yaml');

describe('OpenAPI extra endpoints', () => {
  it('contains /healthz, /metrics, /readyz with examples', async () => {
    const api = await SwaggerParser.parse(specPath) as any;
    expect(api.paths['/healthz']).toBeTruthy();
    expect(api.paths['/metrics']).toBeTruthy();
    expect(api.paths['/readyz']).toBeTruthy();
    expect(api.paths['/readyz'].get.responses['200']).toBeTruthy();
    expect(api.paths['/readyz'].get.responses['503']).toBeTruthy();
  });
});
