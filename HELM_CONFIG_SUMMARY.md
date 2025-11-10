# Helm Configuration Summary

This document summarizes the environment-specific Helm configuration generated for all LMA services.

## Files Generated

For each service under `services/*/helm/`:
- ✅ `values.dev.yaml` - Development environment configuration
- ✅ `values.staging.yaml` - Staging environment configuration
- ✅ `values.prod.yaml` - Production environment configuration
- ✅ `templates/externalsecret.yaml` - ExternalSecrets Operator integration
- ✅ `templates/deployment.yaml` - Patched with env vars and secrets mount

## Service Configuration Matrix

### Common Environment Variables (All Services)

| Variable | Description |
|----------|-------------|
| `ENV` | Environment name (dev/staging/prod) |
| `SERVICE_PORT` | Service REST port |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OpenTelemetry collector endpoint |

### Per-Service Configuration

| Service | REST Port | Non-Secret Env Vars | Secret Keys | Notes |
|---------|-----------|---------------------|-------------|-------|
| **publisher-service** | 7201 | ROLLBACK_STRATEGY, CANARY_HEALTH_URL | STRIPE_API_KEY, GITHUB_APP_PRIVATE_KEY, KUBE_CONFIG | - |
| **fix-engine-service** | 7151 | PATCH_MAX_FILES, PATCH_MAX_LOC, PATCH_TTL_DAYS, DIFF_ENGINE | (none) | No secrets required |
| **scm-app-service** | 7051 | SCM_ALLOWED_HOSTS, SNAPSHOT_BUCKET | GITHUB_APP_ID, GITHUB_APP_PRIVATE_KEY, GITHUB_WEBHOOK_SECRET | GitHub integration |
| **zip-ingest-service** | 7052 | MAX_UPLOAD_MB, CLAMAV_URL, SNAPSHOT_BUCKET | (none) | No secrets required |
| **agent-ingest-service** | 7053 | AGENT_PUBLIC_KEY_JWKS, REPORT_MAX_MB, RESULTS_BUCKET | (none) | No secrets required |
| **snapshot-orchestrator-service** | 7054 | SNAPSHOT_TTL_HOURS, PIPELINE_TOPIC, S3_SNAPSHOT_PREFIX | (none) | No secrets required |
| **projects-service** | 7002 | JWT_PUBLIC_KEY, S3_BUCKET_PROJECTS | DATABASE_URL | PostgreSQL connection |
| **spec-memory-service** | 7101 | EMBEDDING_PROVIDER | DATABASE_URL, EMBEDDING_API_KEY | DB + AI service |
| **ai-debugger-service** | 7102 | REDIS_URL, S3_BUCKET_REPROS | LLM_API_KEY | AI/ML service |
| **scaffold-secure-service** | 7103 | (none) | (none) | Minimal config |
| **secrets-env-service** | 7104 | VAULT_ADDR | VAULT_ROLE_ID, VAULT_SECRET_ID | Vault integration |
| **db-guardian-service** | 7105 | BACKUP_S3_BUCKET | PG_SUPER_DSN | Database admin |
| **dep-governance-service** | 7106 | SYFT_OPTS, GRYPE_DB_DIR, LICENSE_ALLOWLIST | (none) | Security scanning |
| **test-lab-service** | 7202 | PLAYWRIGHT_WORKERS, GRID_URL, PREVIEW_TTL_MIN | (none) | Testing service |
| **authz-matrix-service** | 7203 | SCHEMA_REPO_URL, IDOR_RULESETS_S3 | (none) | Authorization |
| **rate-limit-service** | 7204 | REDIS_URL, DEFAULT_RPM, BURST_MULTIPLIER | (none) | Rate limiting |
| **observability-service** | 7301 | ALERT_EMAIL_FROM, SLO_BUDGETS_JSON | PAGERDUTY_ROUTING_KEY | Monitoring/alerts |
| **perf-coach-service** | 7302 | CDN_PROVIDER, REDIS_URL | PG_READONLY_DSN | Performance analytics |
| **finops-service** | 7401 | CLOUD_COST_EXPORT_S3, FORECAST_HORIZON_DAYS | STRIPE_API_KEY | Billing integration |
| **mvp-validator-service** | 7402 | DEFAULT_ACTIVATION_TARGET | FEATURE_FLAG_SDK_KEY | Feature flags |
| **positioning-engine-service** | 7501 | VERTICAL_CORPORA_S3 | EMBEDDING_API_KEY | AI/ML service |
| **launch-engine-service** | 7502 | (none) | TWITTER_API_KEY, REDDIT_CLIENT_ID, PH_TOKEN | Social media APIs |
| **legal-automator-service** | 7503 | JURISDICTIONS_JSON, DSAR_INBOX_EMAIL | (none) | Compliance |
| **motivation-engine-service** | 7601 | NUDGE_HEURISTICS_JSON, WS_CLUSTER_ID | (none) | User engagement |
| **billing-service** | 7901 | BILLING_TAX_REGION | STRIPE_WEBHOOK_SECRET | Payment processing |
| **notification-service** | 7902 | (none) | RESEND_API_KEY, TWILIO_ACCOUNT_SID, TWILIO_AUTH_TOKEN, SLACK_WEBHOOK_URL | Multi-channel notifications |
| **webhook-relay-service** | 7903 | DLQ_STREAM | WEBHOOK_SIGNING_SECRET | Webhook processing |

## Resource Allocation by Environment

### Development
- CPU: 50m request / 250m limit
- Memory: 64Mi request / 256Mi limit
- HPA: Disabled (single replica)

### Staging
- CPU: 200m request / 1 core limit
- Memory: 256Mi request / 1Gi limit
- HPA: Enabled (2-6 replicas)

### Production
- CPU: 300m request / 2 cores limit
- Memory: 512Mi request / 2Gi limit
- HPA: Enabled (3-10 replicas)

## External Secrets Configuration

### Vault Secret Stores

| Environment | ClusterSecretStore | Namespace |
|-------------|-------------------|-----------|
| Development | `vault-dev` | `lma-dev` |
| Staging | `vault-staging` | `lma-staging` |
| Production | `vault-prod` | `lma-prod` |

### Secret Path Structure

Secrets are stored in Vault at:
```
secret/data/lma/{service-name}/{environment}
```

Example:
- `secret/data/lma/publisher-service/dev`
- `secret/data/lma/projects-service/staging`
- `secret/data/lma/billing-service/prod`

## Deployment Configuration

All deployments have been patched with:

1. **Environment Variables**: Rendered from `.Values.env` map
   ```yaml
   env:
   {{- range $name, $val := .Values.env }}
     - name: {{ $name }}
       value: {{ $val | quote }}
   {{- end }}
   ```

2. **Secret Mounting**: Via `envFrom` referencing ExternalSecret-created Secret
   ```yaml
   {{- if .Values.externalSecrets.enabled }}
   envFrom:
     - secretRef:
         name: {{ .Values.externalSecrets.targetName }}
   {{- end }}
   ```

3. **Preserved Components**:
   - Existing health probes (readiness/liveness)
   - Resource requests/limits (now configurable via values)
   - Security contexts
   - Port configurations

## Image Repository Placeholders

All values files use placeholder format for easy replacement:
```yaml
image:
  repository: ghcr.io/{{ORG}}/{{REPO}}-{service-name}
```

To finalize:
```bash
find services -name "values*.yaml" -exec sed -i '' 's/{{ORG}}/your-org-name/g' {} \;
find services -name "values*.yaml" -exec sed -i '' 's/{{REPO}}/lastmile-accelerator/g' {} \;
```

## Usage

### Deploy to Development
```bash
helm upgrade --install publisher-service \
  services/publisher-service/helm \
  --values services/publisher-service/helm/values.dev.yaml \
  --namespace lma-dev
```

### Deploy to Staging
```bash
helm upgrade --install publisher-service \
  services/publisher-service/helm \
  --values services/publisher-service/helm/values.staging.yaml \
  --set image.tag=${CI_COMMIT_SHA} \
  --namespace lma-staging
```

### Deploy to Production
```bash
helm upgrade --install publisher-service \
  services/publisher-service/helm \
  --values services/publisher-service/helm/values.prod.yaml \
  --set image.tag=${CI_COMMIT_SHA} \
  --namespace lma-prod
```

## Validation

All 135 generated files have been validated:
- ✅ 27 services × 3 environments × 1 values file = 81 values files
- ✅ 27 services × 1 externalsecret template = 27 externalsecret files
- ✅ 27 services × 1 deployment template = 27 deployment files

Total: **135 files successfully generated and validated**

## Next Steps

1. **Replace Placeholders**: Update `{{ORG}}` and `{{REPO}}` in all values files
2. **Populate Vault Secrets**: Add actual secret values to Vault at the documented paths
3. **Install External Secrets Operator**: Ensure ESO is deployed in each cluster
4. **Create ClusterSecretStores**: Configure `vault-dev`, `vault-staging`, `vault-prod`
5. **Test Deployments**: Deploy each service to dev environment first
6. **Update CI/CD**: Configure pipelines to use appropriate values files per environment

## Maintenance

When adding a new service:
1. Add entry to `service_catalog.yaml`
2. Add env vars to `NON_SECRET_ENV` dict in `generate_helm_configs.py`
3. Add secrets to `SECRETS_MATRIX` dict in `generate_helm_configs.py`
4. Rerun `python3 generate_helm_configs.py`

## Security Notes

⚠️ **Important**: 
- No plaintext secrets are stored in Git
- All sensitive data flows through External Secrets Operator → Vault
- Secret keys are documented but values must be managed in Vault
- Update Vault paths/keys before deploying to ensure ExternalSecrets can sync
