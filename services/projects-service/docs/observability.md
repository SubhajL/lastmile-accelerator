# Observability Guide

This service exposes Prometheus metrics at `/metrics` and emits OpenTelemetry traces.
The dashboards and alert rules in this folder are a ready‑to‑import starter kit.

## How to import Grafana dashboards
1. Open Grafana → Dashboards → Import
2. Upload one of the JSON files from `docs/grafana/`:
   - `http_latency_dashboard.json`
   - `error_rate_dashboard.json`
   - `throughput_dashboard.json`
3. Select your Prometheus data source and save.

### Panels & queries
- Average latency (ms):
  - `increase(request_duration_ms_sum[5m]) / increase(request_duration_ms_count[5m])`
- Error rate (%):
  - `100 * sum(rate(request_errors_total[5m])) / sum(rate(request_total[5m]))`
- Throughput (req/s):
  - `sum(rate(request_total[5m]))`

Note on p95 latency:
- The starter kit uses a summary metric and therefore shows average latency.
- To enable p95/p99, export histograms and replace queries with `histogram_quantile(...)`.

## Prometheus alerts
Import `docs/alerts/prometheus.rules.yaml` into your Prometheus/Alertmanager stack.

Included sample rules:
- HighErrorRate (>5% for 10m):
  - `sum(rate(request_errors_total[5m])) / sum(rate(request_total[5m])) > 0.05`
- HighLatencyAverage (>500ms for 10m):
  - `increase(request_duration_ms_sum[5m]) / increase(request_duration_ms_count[5m]) > 500`

If histograms are enabled, consider adding a true p95 alert:
- `histogram_quantile(0.95, sum(rate(http_server_duration_ms_bucket[5m])) by (le)) > 500`

## Prometheus scrape configuration
Ensure your Prometheus instance scrapes the service `/metrics` endpoint.

Example static scrape config:
```yaml
scrape_configs:
  - job_name: 'projects-service'
    static_configs:
      - targets: ['projects-service:7002']
    metrics_path: /metrics
```

## Metrics reference
- `request_total` — total HTTP requests (counter)
- `request_errors_total` — total 5xx responses (counter)
- `request_duration_ms_sum` — cumulative request duration (ms) (summary)
- `request_duration_ms_count` — number of observed requests (summary)

These are also exported via OpenTelemetry as histograms if you enable a metrics reader with histogram support; the starter kit queries are aligned to the in‑service `/metrics` output by default.
