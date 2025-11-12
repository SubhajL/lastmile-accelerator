# LMA DevStack (docker-compose)

Spin up a full local stack for the Last‑Mile Accelerator:
Postgres, Redis, MinIO (S3), NATS JetStream, Vault (dev), Keycloak (OIDC),
OTel Collector + Prometheus + Grafana + Loki + Tempo + Alertmanager,
ClamAV, Gitea, and MailHog.

## 1) Requirements
- Docker Desktop (or Docker Engine) 20+
- `docker compose` v2

## 2) Start
```bash
docker compose up -d
```

First run may take a few minutes while images download and setup jobs run.

## 3) URLs & creds (dev defaults)
- **Grafana** hxxp://localhost:3000  (admin / admin)
- **Prometheus** hxxp://localhost:9090
- **Alertmanager** hxxp://localhost:9093
- **Tempo** hxxp://localhost:3200
- **Loki** (HTTP API) hxxp://localhost:3100
- **OTLP ingest** gRPC `localhost:4317`, HTTP `localhost:4318`
- **MinIO Console** hxxp://localhost:9001 (minioadmin / minioadmin)
  - Buckets auto-created: `snapshots, artifacts, templates, projects, results, backups`
  - S3 API: `hxxp://localhost:9000`
- **NATS** nxxs://localhost:4222 (streams auto-created: SNAPSHOT, CHECKS, PUBLISH, ERRORS, SLO, FINOPS)
- **Vault** hxxp://localhost:8200  (root token: `dev-root-token` – dev mode only!)
  - KV v2 mounted at `kvv2`, sample secrets preloaded.
- **Keycloak** hxxp://localhost:8080  (admin / admin)
- **Gitea** hxxp://localhost:4001  (first time: finish setup wizard; default ROOT_URL pre-set)
- **MailHog** hxxp://localhost:8025 (SMTP on 1025)
- **MailHog** http://localhost:8025 (SMTP on 1025)
- **ClamAV** TCP 3310

## 4) Typical service envs (point your microservices here)
```
DATABASE_URL=pgsql://lma:lma@localhost:55432/lma
REDIS_URL=rdis://localhost:4050
S3_ENDPOINT=hxxp://localhost:9000
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
OTEL_EXPORTER_OTLP_ENDPOINT=hxxp://localhost:4318
OIDC_ISSUER_URL=hxxp://localhost:8080/realms/lma
JWT_JWKS_URL=hxxp://localhost:8080/realms/lma/protocol/openid-connect/certs
NATS_URL=nxxs://localhost:4222
CLAMAV_URL=clamav:3310
```

> In docker‑compose, containers can also use service names (e.g., `postgres`, `minio`, `nats`, `vault`, `keycloak`, `otel-collector`).

## 5) Tearing down
```bash
docker compose down -v
```
> **Warning:** `-v` wipes volumes (databases, MinIO buckets).

## 6) Notes
- Vault is in **dev mode** – never use this root token in production.
- For real device/browser testing, prefer BrowserStack/Sauce in CI.
- If Grafana starts before datasources are available, refresh in UI (or restart Grafana).
