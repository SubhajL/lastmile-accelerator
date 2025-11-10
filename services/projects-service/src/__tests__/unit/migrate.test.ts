import { describe, it, expect, vi, beforeEach } from 'vitest';
import * as db from '../../db';

// Subject under test
import { runMigrations } from '../../db/migrations/migrate';

describe('migrations/migrate.ts', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('applies pending .sql files in order and records them', async () => {
    // Arrange: two files out of order
    const readdirMock = vi.fn().mockResolvedValue(['002_second.sql', '001_first.sql']);
    const readFileMock = vi.fn().mockImplementation(async (path: any) => {
      if (String(path).includes('001_first.sql')) return '/* 001 */\nCREATE TABLE t1(id int);';
      if (String(path).includes('002_second.sql')) return '/* 002 */\nALTER TABLE t1 ADD COLUMN name text;';
      return '';
    });

    const queryCalls: Array<{ sql: string; params?: any[] }> = [];
    const querySpy = vi.spyOn(db, 'query').mockImplementation(async (sql: string, params?: any[]) => {
      queryCalls.push({ sql, params });
      if (/SELECT\s+name\s+FROM\s+migrations/i.test(sql)) {
        return { rows: [] } as any; // no applied migrations yet
      }
      return { rows: [], rowCount: 0 } as any;
    });

    // Act
    await runMigrations({ migrationsDir: 'src/db/migrations', readdirFn: readdirMock as any, readFileFn: readFileMock as any });

    // Assert: created migrations table
    expect(querySpy).toHaveBeenCalled();
    expect(queryCalls.some(c => /CREATE TABLE IF NOT EXISTS\s+migrations/i.test(c.sql))).toBe(true);

    // Applied in numeric order: 001 then 002
    const appliedOrder = queryCalls
      .filter(c => /INSERT\s+INTO\s+migrations/i.test(c.sql))
      .map(c => c.params?.[0]);

    expect(appliedOrder.length).toBe(2);
    expect(appliedOrder[0]).toBe('001_first.sql');
    expect(appliedOrder[1]).toBe('002_second.sql');
  });
});