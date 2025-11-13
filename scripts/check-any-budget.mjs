#!/usr/bin/env node
import fs from 'node:fs';
import path from 'node:path';

const args = process.argv;
const cfgIdx = args.indexOf('--config');
const configPath = cfgIdx !== -1 ? args[cfgIdx + 1] : '.any-budget.json';

if (!fs.existsSync(configPath)) {
  console.error(`[any-budget] config not found: ${configPath}`);
  process.exit(1);
}

const cfg = JSON.parse(fs.readFileSync(configPath, 'utf8'));

const isTsSrc = (p) => p.endsWith('.ts') && !p.endsWith('.test.ts') && !p.endsWith('.spec.ts');
const walk = (dir) => {
  const out = [];
  for (const e of fs.readdirSync(dir, { withFileTypes: true })) {
    const p = path.join(dir, e.name);
    if (e.isDirectory()) out.push(...walk(p));
    else if (isTsSrc(p)) out.push(p);
  }
  return out;
};

const ANY_RE = /(:|\bas)\s+any\b/g;

let failures = 0;
for (const [pkg, { srcDir, maxAnyCount }] of Object.entries(cfg)) {
  const files = walk(srcDir);
  let count = 0;
  for (const f of files) {
    const txt = fs.readFileSync(f, 'utf8');
    const matches = txt.match(ANY_RE);
    if (matches) count += matches.length;
  }
  if (count > maxAnyCount) {
    console.error(`[any-budget] ${pkg}: found ${count} > budget ${maxAnyCount}`);
    failures++;
  } else {
    console.log(`[any-budget] ${pkg}: ${count}/${maxAnyCount} \u2714`);
  }
}

if (failures) process.exit(1);
