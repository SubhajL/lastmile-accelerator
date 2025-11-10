import { describe, it, expect } from 'vitest';
import {
  parseCreateProject,
  parseUpdateProject,
  parseCreateMember,
  parseCreateEnvironment,
  parseSetIngestionModes,
} from '../../utils/validators';

describe('utils/validators.ts', () => {
  describe('parseCreateProject', () => {
    it('accepts valid name and optional description', () => {
      const dto = parseCreateProject({ name: 'Alpha', description: 'First' });
      expect(dto).toEqual({ name: 'Alpha', description: 'First' });
    });

    it('rejects missing or empty name', () => {
      expect(() => parseCreateProject({} as any)).toThrowError();
      expect(() => parseCreateProject({ name: '' } as any)).toThrowError();
    });
  });

  describe('parseUpdateProject', () => {
    it('accepts at least one field', () => {
      const dto = parseUpdateProject({ name: 'New' });
      expect(dto).toEqual({ name: 'New' });
    });

    it('rejects when no fields provided', () => {
      expect(() => parseUpdateProject({} as any)).toThrowError();
    });
  });

  describe('parseCreateMember', () => {
    it('accepts valid email and role', () => {
      const dto = parseCreateMember({ email: 'u@example.com', role: 'developer' });
      expect(dto).toEqual({ email: 'u@example.com', role: 'developer' });
    });

    it('rejects invalid role', () => {
      expect(() => parseCreateMember({ email: 'u@example.com', role: 'boss' } as any)).toThrowError();
    });
  });

  describe('parseCreateEnvironment', () => {
    it('accepts name and optional config object', () => {
      const dto = parseCreateEnvironment({ name: 'staging', config: { region: 'us' } });
      expect(dto).toEqual({ name: 'staging', config: { region: 'us' } });
    });

    it('rejects empty name', () => {
      expect(() => parseCreateEnvironment({ name: '' } as any)).toThrowError();
    });
  });

  describe('parseSetIngestionModes', () => {
    it('accepts modes from A/B/C with default included', () => {
      const dto = parseSetIngestionModes({ modes: ['A', 'B'], defaultMode: 'A' });
      expect(dto).toEqual({ modes: ['A', 'B'], defaultMode: 'A' });
    });

    it('rejects when default not in modes', () => {
      expect(() => parseSetIngestionModes({ modes: ['A', 'B'], defaultMode: 'C' } as any)).toThrowError();
    });

    it('rejects invalid modes', () => {
      expect(() => parseSetIngestionModes({ modes: ['D'], defaultMode: 'A' } as any)).toThrowError();
    });
  });
});