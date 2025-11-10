import { describe, it, expect } from 'vitest';
import fs from 'node:fs';
import path from 'node:path';

describe('Observability docs content', () => {
  it('observability.md mentions dashboards, alerts, scrape, metrics', () => {
    const p = path.resolve(__dirname, '../../../..', 'docs/observability.md');
    const s = fs.readFileSync(p, 'utf-8');
    expect(s).toContain('How to import Grafana dashboards');
    expect(s).toContain('Prometheus alerts');
    expect(s).toContain('Prometheus scrape configuration');
    expect(s).toContain('Metrics reference');
    expect(s).toContain('throughput_dashboard.json');
    expect(s).toContain('request_total');
    expect(s).toContain('request_errors_total');
  });
});
