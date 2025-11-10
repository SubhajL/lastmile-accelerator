import { readdir, readFile } from 'fs/promises';
import path from 'path';
import { query } from '../../db';

export interface MigrateOptions {
  migrationsDir?: string;
  readdirFn?: typeof readdir;
  readFileFn?: typeof readFile;
}

export async function runMigrations(opts: MigrateOptions = {}): Promise<void> {
  const migrationsDir = opts.migrationsDir ?? 'src/db/migrations';
  const readdirImpl = opts.readdirFn ?? readdir;
  const readFileImpl = opts.readFileFn ?? readFile;

  // 1) Ensure migrations table exists
  await query(
    `CREATE TABLE IF NOT EXISTS migrations (
      name TEXT PRIMARY KEY,
      applied_at TIMESTAMPTZ DEFAULT NOW()
    )`
  );

  // 2) Determine already applied migrations
  const applied = await query('SELECT name FROM migrations');
  const appliedSet = new Set((applied as any).rows.map((r: any) => r.name));

  // 3) Load .sql files and sort
  const files = (await readdirImpl(migrationsDir))
    .filter((f: any) => String(f).endsWith('.sql'))
    .sort();

  // 4) Apply pending
  for (const file of files) {
    if (appliedSet.has(file)) continue;
    const filePath = path.join(migrationsDir, file);
    const sql = await readFileImpl(filePath, 'utf8');

    if (sql.trim().length === 0) continue;

    await query(sql);
    await query('INSERT INTO migrations(name) VALUES ($1)', [file]);
  }
}