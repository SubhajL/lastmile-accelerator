import { describe, it, expect } from 'vitest';
import SwaggerParser from '@apidevtools/swagger-parser';
import path from 'node:path';

const specPath = path.resolve(__dirname, '../../../..', 'docs/openapi.yaml');

describe('OpenAPI spec validation', () => {
  it('docs/openapi.yaml validates successfully', async () => {
    const api = await SwaggerParser.validate(specPath) as any;
    expect(api.openapi || api.swagger).toBe('3.1.0');
  });

  it('documents required scopes on key endpoints', async () => {
    const api = await SwaggerParser.parse(specPath) as any;
    const paths = api.paths || {};
    // spot-check that descriptions mention scope requirements
    expect(paths['/v1/projects'].get.description).toContain('projects:read');
    expect(paths['/v1/projects'].post.description).toContain('projects:write');
    expect(paths['/v1/tenants/{tenantId}'].get.description).toContain('tenants:read');
    expect(paths['/v1/tenants/{tenantId}/members'].post.description).toContain('members:write');
    expect(paths['/v1/projects/{projectId}/environments'].get.description).toContain('environments:read');
    expect(paths['/v1/projects/{projectId}/ingestion-modes'].post.description).toContain('environments:write');
  });
});
