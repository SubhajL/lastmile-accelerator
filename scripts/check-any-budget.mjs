#!/usr/bin/env node
import fs from 'node:fs';
import path from 'node:path';

const repoRoot = process.cwd();
const budgetsPath = path.join(repoRoot, '.any-budget.json');

function readJson(p) {
  return JSON.parse(fs.readFileSync(p, 'utf8'));
}

function listServiceDirs() {
  const dir = path.join(repoRoot, 'services');
  if (!fs.existsSync(dir)) return [];
  return fs.readdirSync(dir).map((name) => path.join(dir, name)).filter((p) => fs.existsSync(path.join(p, 'package.json')));
}

function mapPkgNameToDir() {
  const map = new Map();
  for (const dir of listServiceDirs()) {
    try {
      const pj = readJson(path.join(dir, 'package.json'));
      if (pj && pj.name) map.set(pj.name, dir);
    } catch {
      // ignore invalid package.json
    }
  }
  return map;
}

function isTestFile(file) {
  return /\.spec\.ts$/.test(file) || /\.test\.ts$/.test(file) || file.includes(`${path.sep}__tests__${path.sep}`);
}

function* walk(dir) {
  const entries = fs.readdirSync(dir, { withFileTypes: true });
  for (const e of entries) {
    const p = path.join(dir, e.name);
    if (e.isDirectory()) {
      if (e.name === 'node_modules' || e.name === 'dist' || e.name === '.next') continue;
      yield* walk(p);
    } else {
      yield p;
    }
  }
}

function countAnyInFile(file) {
  const src = fs.readFileSync(file, 'utf8');
  const patterns = [
    /:\s*any\b/g, // type annotations
    /\bas\s+any\b/g, // casts
    /<\s*any\s*>/g, // generic angle brackets
  ];
  let count = 0;
  for (const re of patterns) {
    count += (src.match(re) || []).length;
  }
  return count;
}

function countAnyInPackage(pkgDir) {
  const srcDir = path.join(pkgDir, 'src');
  if (!fs.existsSync(srcDir)) return 0;
  let total = 0;
  for (const file of walk(srcDir)) {
    if (!file.endsWith('.ts')) continue;
    if (isTestFile(file)) continue;
    total += countAnyInFile(file);
  }
  return total;
}

function main() {
  if (!fs.existsSync(budgetsPath)) {
    console.error(`.any-budget.json not found at ${budgetsPath}`);
    process.exit(1);
  }
  const budgets = readJson(budgetsPath);
  const map = mapPkgNameToDir();
  let failed = false;
  for (const [pkgName, conf] of Object.entries(budgets)) {
    const dir = map.get(pkgName);
    if (!dir) {
      console.warn(`[any-budget] Skipping unknown package: ${pkgName}`);
      continue;
    }
    const max = typeof conf === 'number' ? conf : conf.maxAnyCount;
    const actual = countAnyInPackage(dir);
    const status = actual <= max ? 'OK' : 'EXCEEDED';
    console.log(`[any-budget] ${pkgName}: ${actual}/${max} ${status}`);
    if (actual > max) failed = true;
  }
  if (failed) {
    console.error('\n[any-budget] Budget exceeded. Reduce any/as any usage or increase baseline intentionally.');
    process.exit(1);
  }
}

main();
