#!/usr/bin/env node
import { promises as fs } from 'fs';
import path from 'path';

const repoRoot = process.cwd();
const baselinePath = path.join(repoRoot, '.any-budget.json');

function log(msg) { console.log(`[any-budget] ${msg}`); }

const defaultIgnore = [
  'node_modules', 'dist', 'build', '.turbo', '.next', '.git'
];

const testSuffixes = ['.spec.ts', '.test.ts', '.spec.tsx', '.test.tsx'];

async function readJSON(p) {
  try { return JSON.parse(await fs.readFile(p, 'utf8')); } catch { return null; }
}

async function exists(p) {
  try { await fs.access(p); return true; } catch { return false; }
}

async function walk(dir, files = []) {
  const entries = await fs.readdir(dir, { withFileTypes: true });
  for (const e of entries) {
    if (defaultIgnore.includes(e.name)) continue;
    const full = path.join(dir, e.name);
    if (e.isDirectory()) {
      await walk(full, files);
    } else if (e.isFile()) {
      if ((e.name.endsWith('.ts') || e.name.endsWith('.tsx')) &&
          !e.name.endsWith('.d.ts') &&
          !testSuffixes.some(s => e.name.endsWith(s))) {
        files.push(full);
      }
    }
  }
  return files;
}

function countAny(content) {
  const re = /(\bas\s+any\b|:\s*any\b)/g; // 'as any' or ': any'
  let m, count = 0;
  while ((m = re.exec(content)) !== null) count++;
  return count;
}

async function countForTarget(targetPath) {
  if (!(await exists(targetPath))) return { files: 0, any: 0 };
  const files = await walk(targetPath);
  let total = 0;
  for (const f of files) {
    const content = await fs.readFile(f, 'utf8');
    total += countAny(content);
  }
  return { files: files.length, any: total };
}

async function main() {
  const baseline = (await readJSON(baselinePath)) || { defaults: { maxAnyCount: 9999 }, overrides: {} };
  const overrides = baseline.overrides || {};
  const roots = ['services', 'packages', 'frontends'];

  let violations = [];
  for (const [target, cfg] of Object.entries(overrides)) {
    const p = path.join(repoRoot, target);
    const { any } = await countForTarget(p);
    const max = typeof cfg.maxAnyCount === 'number' ? cfg.maxAnyCount : (baseline.defaults?.maxAnyCount ?? 9999);
    log(`${target}: any=${any}, max=${max}`);
    if (any > max) violations.push({ target, any, max });
  }

  if (violations.length) {
    console.error('[any-budget] Violations found:');
    for (const v of violations) console.error(`  ${v.target}: any=${v.any} > max=${v.max}`);
    process.exit(1);
  }
  log('OK');
}

main().catch(e => { console.error(e); process.exit(1); });
