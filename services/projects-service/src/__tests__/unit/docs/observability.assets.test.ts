import { describe, it, expect } from 'vitest';
import fs from 'node:fs';
import path from 'node:path';

describe('Observability assets present and valid', () => {
  it('prometheus rules exist and mention alerts', () => {
    const p = path.resolve(__dirname, '../../../..', 'docs/alerts/prometheus.rules.yaml');
    const s = fs.readFileSync(p, 'utf-8');
    expect(s).toContain('HighErrorRate');
    expect(s).toContain('HighLatency');
    expect(s).toContain('groups:');
  });

  it('grafana dashboards exist and are JSON', () => {
    const base = path.resolve(__dirname, '../../../..', 'docs/grafana');
    const files = ['http_latency_dashboard.json', 'error_rate_dashboard.json', 'throughput_dashboard.json'];
    for (const f of files) {
      const s = fs.readFileSync(path.join(base, f), 'utf-8');
      const obj = JSON.parse(s);
      expect(obj.title).toBeTruthy();
      expect(Array.isArray(obj.panels)).toBe(true);
    }
  });
});
