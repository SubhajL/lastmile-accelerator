# Helm Configuration Quick Start

## Before You Begin

Replace placeholders in all values files:
```bash
# Update organization name
find services -name "values*.yaml" -exec sed -i '' 's/{{ORG}}/your-github-org/g' {} \;

# Update repository name
find services -name "values*.yaml" -exec sed -i '' 's/{{REPO}}/lastmile-accelerator/g' {} \;
```

## Prerequisites

1. **External Secrets Operator** installed in your clusters
2. **ClusterSecretStores** configured:
   - `vault-dev` → Vault instance for development
   - `vault-staging` → Vault instance for staging
   - `vault-prod` → Vault instance for production
3. **Vault secrets** populated at paths like:
   ```
   secret/data/lma/{service-name}/dev
   secret/data/lma/{service-name}/staging
   secret/data/lma/{service-name}/prod
   ```

## Deploying a Service

### Development
```bash
helm upgrade --install {service-name} \
  services/{service-name}/helm \
  --values services/{service-name}/helm/values.dev.yaml \
  --namespace lma-dev \
  --create-namespace
```

### Staging (from CI)
```bash
helm upgrade --install {service-name} \
  services/{service-name}/helm \
  --values services/{service-name}/helm/values.staging.yaml \
  --set image.tag=${CI_COMMIT_SHA:0:7} \
  --namespace lma-staging \
  --create-namespace \
  --wait
```

### Production (with approval)
```bash
helm upgrade --install {service-name} \
  services/{service-name}/helm \
  --values services/{service-name}/helm/values.prod.yaml \
  --set image.tag=${RELEASE_TAG} \
  --namespace lma-prod \
  --create-namespace \
  --wait \
  --timeout 5m
```

## Debugging

### Check if ExternalSecret synced
```bash
kubectl get externalsecret -n lma-dev
kubectl describe externalsecret {service-name}-secrets -n lma-dev
```

### Check if Secret was created
```bash
kubectl get secret {service-name}-secrets -n lma-dev
```

### View environment variables in pod
```bash
kubectl exec -n lma-dev deploy/{service-name} -- env | grep -E "ENV|SERVICE_PORT|OTEL"
```

### Check Secret keys (without values)
```bash
kubectl get secret {service-name}-secrets -n lma-dev -o jsonpath='{.data}' | jq 'keys'
```

## Customizing Values

Override values at deployment time:
```bash
helm upgrade --install publisher-service \
  services/publisher-service/helm \
  --values services/publisher-service/helm/values.dev.yaml \
  --set resources.limits.memory=512Mi \
  --set hpa.enabled=true \
  --set hpa.minReplicas=2
```

## Environment-Specific Configurations

| Environment | Use Case | Resources | Replicas | Auto-scaling |
|-------------|----------|-----------|----------|--------------|
| **dev** | Local testing, debugging | Minimal (50m/64Mi) | 1 | No |
| **staging** | Integration tests, QA | Medium (200m/256Mi) | 2-6 | Yes |
| **prod** | Production traffic | High (300m/512Mi) | 3-10 | Yes |

## Common Issues

### ExternalSecret not syncing
```bash
# Check ClusterSecretStore
kubectl get clustersecretstore vault-dev

# Check ESO logs
kubectl logs -n external-secrets-system deploy/external-secrets
```

### Secret missing required keys
```bash
# Verify Vault path
vault kv get secret/data/lma/{service-name}/dev

# Check ExternalSecret spec
kubectl get externalsecret {service-name}-secrets -n lma-dev -o yaml
```

### Pod not starting - ImagePullBackOff
```bash
# Verify image exists
crane manifest ghcr.io/{org}/{repo}-{service}:dev

# Check imagePullSecrets if needed
kubectl get serviceaccount default -n lma-dev -o yaml
```

## File Structure

```
services/{service-name}/helm/
├── Chart.yaml                      # Helm chart metadata
├── values.yaml                     # Base values (original)
├── values.dev.yaml                # ✨ Dev environment overrides
├── values.staging.yaml            # ✨ Staging environment overrides
├── values.prod.yaml               # ✨ Production environment overrides
└── templates/
    ├── deployment.yaml            # ✨ Patched with env vars
    ├── externalsecret.yaml        # ✨ External Secrets integration
    ├── service.yaml               # Service definition
    ├── hpa.yaml                   # Horizontal Pod Autoscaler
    └── networkpolicy.yaml         # Network policies
```

Legend: ✨ = Newly generated/modified files

## Next Steps

1. Review `HELM_CONFIG_SUMMARY.md` for detailed configuration matrix
2. Populate Vault with actual secret values
3. Test deployments in dev environment
4. Update CI/CD pipelines to use appropriate values files
5. Configure monitoring alerts for ExternalSecret sync failures

## Support

- Configuration Matrix: See `HELM_CONFIG_SUMMARY.md`
- Service Catalog: See `service_catalog.yaml`
- WARP Guide: See `WARP.md`
