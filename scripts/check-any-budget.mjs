#!/usr/bin/env node
import { readFile, readdir, stat } from 'node:fs/promises';
import path from 'node:path';

const argv = process.argv.slice(2);
const configFlagIdx = argv.indexOf('--config');
const configPath = configFlagIdx !== -1 ? argv[configFlagIdx + 1] : '.any-budget.json';

function shouldSkip(filePath) {
  const p = filePath.replace(/\\/g, '/');
  if (/node_modules\//.test(p)) return true;
  if (/^(dist|build|coverage|.git)\//.test(p)) return true;
  if (/__tests__\//.test(p)) return true;
  if (/\.(test|spec)\.[cm]?tsx?$/.test(p)) return true;
  if (/\.d\.[cm]?ts$/.test(p)) return true;
  return false;
}

async function walk(dir) {
  const out = [];
  const entries = await readdir(dir, { withFileTypes: true });
  for (const e of entries) {
    const full = path.join(dir, e.name);
    if (e.isDirectory()) {
      if (shouldSkip(full + '/')) continue;
      out.push(...await walk(full));
    } else if (e.isFile()) {
      if (shouldSkip(full)) continue;
      if (/\.[cm]?tsx?$/.test(e.name)) out.push(full);
    }
  }
  return out;
}

function countAny(text) {
  const re1 = /:\s*any\b/g; // type annotations
  const re2 = /\bas\s+any\b/g; // casts
  const c1 = text.match(re1)?.length || 0;
  const c2 = text.match(re2)?.length || 0;
  return c1 + c2;
}

async function main() {
  let cfgRaw;
  try {
    cfgRaw = await readFile(configPath, 'utf8');
  } catch (e) {
    console.error(`any-budget: config not found at ${configPath}`);
    process.exit(1);
  }
  const cfg = JSON.parse(cfgRaw);
  let failed = false;
  for (const [name, opts] of Object.entries(cfg)) {
    const srcDir = opts.srcDir;
    const max = Number(opts.maxAnyCount ?? 0);
    if (!srcDir || Number.isNaN(max)) {
      console.error(`any-budget: invalid config for ${name}`);
      failed = true;
      continue;
    }
    const files = (await walk(srcDir)).sort();
    let total = 0;
    for (const f of files) {
      const text = await readFile(f, 'utf8');
      total += countAny(text);
    }
    const status = total <= max ? 'OK' : 'FAIL';
    console.log(`${name}: any-uses=${total} budget=${max} => ${status}`);
    if (total > max) failed = true;
  }
  if (failed) {
    console.error('any-budget: exceeded budget. Raise budget temporarily or reduce any/as any usage.');
    process.exit(1);
  }
}

main().catch((e) => {
  console.error('any-budget: error', e);
  process.exit(1);
});
