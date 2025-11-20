# Infrastructure - Kubernetes, Helm, and Service Mesh

**Parent Context:** This extends [../CLAUDE.md](../CLAUDE.md)

This directory contains infrastructure-as-code for deploying LMA services to Kubernetes, including Helm charts, Envoy proxy configurations, and service mesh setup.

## Directory Structure

```
infra/
├── envoy/                  # Envoy proxy configurations
│   ├── envoy.yaml          # Main Envoy config
│   └── filters/            # Custom Envoy filters
├── helm/                   # Shared Helm charts and values
│   ├── common/             # Common chart templates
│   └── values/             # Shared values files
└── k8s/                    # Base Kubernetes manifests
    ├── namespaces/         # Namespace definitions
    ├── rbac/               # RBAC roles and bindings
    └── network-policies/   # Network policies
```

## Deployment Architecture

### Kubernetes Cluster
- **Platform:** GKE, EKS, or self-managed K8s
- **Namespaces:** Separate namespaces per environment (dev, staging, prod)
- **Service Mesh:** Envoy sidecar for mTLS and traffic management

### Service Deployment Pattern
Each service is deployed using Helm:
```
services/<service-name>/helm/
├── Chart.yaml              # Helm chart metadata
├── values.yaml             # Default values
├── values-dev.yaml         # Dev environment overrides
├── values-staging.yaml     # Staging overrides
├── values-prod.yaml        # Production overrides
└── templates/
    ├── deployment.yaml     # Kubernetes Deployment
    ├── service.yaml        # Kubernetes Service (REST + gRPC)
    ├── configmap.yaml      # Configuration
    ├── secret.yaml         # Secrets (from Vault)
    ├── hpa.yaml            # Horizontal Pod Autoscaler
    ├── pdb.yaml            # Pod Disruption Budget
    └── ingress.yaml        # Ingress (if external)
```

## Helm Chart Conventions

### Standard Chart Structure

**Chart.yaml:**
```yaml
apiVersion: v2
name: <service-name>
description: <Service description>
version: 0.1.0
appVersion: "1.0"
```

**values.yaml (common structure):**
```yaml
replicaCount: 2

image:
  repository: ghcr.io/subhajl/lastmile-accelerator/<service-name>
  pullPolicy: IfNotPresent
  tag: ""  # Overridden by CI/CD

service:
  type: ClusterIP
  rest:
    port: <rest-port>
  grpc:
    port: <grpc-port>

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 250m
    memory: 256Mi

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80

env:
  - name: SERVICE_NAME
    value: <service-name>
  - name: DATABASE_URL
    valueFrom:
      secretKeyRef:
        name: postgres-credentials
        key: url
  - name: REDIS_URL
    value: redis://redis:6379
  - name: NATS_URL
    value: nats://nats:4222

securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - ALL

livenessProbe:
  httpGet:
    path: /healthz
    port: rest
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /healthz
    port: rest
  initialDelaySeconds: 5
  periodSeconds: 5
```

### Environment-Specific Values

**values-dev.yaml:**
```yaml
replicaCount: 1

resources:
  limits:
    cpu: 200m
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 128Mi

autoscaling:
  enabled: false

env:
  - name: LOG_LEVEL
    value: debug
```

**values-prod.yaml:**
```yaml
replicaCount: 3

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 500m
    memory: 512Mi

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 20

env:
  - name: LOG_LEVEL
    value: info
```

## Deployment Commands

### Install/Upgrade Service

```bash
# Development
helm upgrade --install <service-name> \
  services/<service-name>/helm/ \
  -f services/<service-name>/helm/values-dev.yaml \
  --namespace lma-dev \
  --create-namespace

# Staging
helm upgrade --install <service-name> \
  services/<service-name>/helm/ \
  -f services/<service-name>/helm/values-staging.yaml \
  --namespace lma-staging \
  --create-namespace

# Production
helm upgrade --install <service-name> \
  services/<service-name>/helm/ \
  -f services/<service-name>/helm/values-prod.yaml \
  --namespace lma-prod \
  --create-namespace \
  --atomic \
  --timeout 5m
```

### Set Image Tag (CI/CD)

```bash
helm upgrade --install <service-name> \
  services/<service-name>/helm/ \
  --set image.tag=abc1234 \
  --namespace lma-prod
```

### Rollback

```bash
# List releases
helm history <service-name> --namespace lma-prod

# Rollback to previous version
helm rollback <service-name> --namespace lma-prod

# Rollback to specific revision
helm rollback <service-name> <revision> --namespace lma-prod
```

### Uninstall

```bash
helm uninstall <service-name> --namespace lma-dev
```

## Service Mesh (Envoy)

### Envoy Sidecar Pattern

Each service pod includes an Envoy sidecar for:
- **mTLS:** Mutual TLS for service-to-service communication
- **Traffic Management:** Load balancing, retries, timeouts
- **Observability:** Metrics, tracing, logging

**Deployment with Envoy Sidecar:**
```yaml
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    metadata:
      annotations:
        envoy.inject: "true"  # Auto-inject Envoy sidecar
    spec:
      containers:
        - name: <service-name>
          # ... service container
        - name: envoy
          image: envoyproxy/envoy:v1.27-latest
          volumeMounts:
            - name: envoy-config
              mountPath: /etc/envoy
      volumes:
        - name: envoy-config
          configMap:
            name: envoy-config
```

### Envoy Configuration

**File:** `infra/envoy/envoy.yaml`

**Key Features:**
- HTTP/2 and gRPC support
- Circuit breaking
- Retry policies
- Request timeouts
- Distributed tracing (OpenTelemetry)
- Access logging

**Example Cluster Config:**
```yaml
clusters:
  - name: projects-service
    connect_timeout: 0.25s
    type: STRICT_DNS
    lb_policy: ROUND_ROBIN
    load_assignment:
      cluster_name: projects-service
      endpoints:
        - lb_endpoints:
            - endpoint:
                address:
                  socket_address:
                    address: projects-service
                    port_value: 7002
    circuit_breakers:
      thresholds:
        - max_connections: 1000
          max_pending_requests: 1000
          max_requests: 1000
          max_retries: 3
```

## Kubernetes Resources

### Namespaces

```yaml
# infra/k8s/namespaces/lma-dev.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: lma-dev
  labels:
    environment: dev
```

### RBAC

```yaml
# infra/k8s/rbac/service-account.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: lma-service-account
  namespace: lma-prod
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: lma-service-role
  namespace: lma-prod
rules:
  - apiGroups: [""]
    resources: ["secrets", "configmaps"]
    verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: lma-service-rolebinding
  namespace: lma-prod
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: lma-service-role
subjects:
  - kind: ServiceAccount
    name: lma-service-account
    namespace: lma-prod
```

### Network Policies

```yaml
# infra/k8s/network-policies/allow-internal.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-internal-only
  namespace: lma-prod
spec:
  podSelector: {}
  policyTypes:
    - Ingress
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              environment: prod
```

## Secrets Management

### Vault Integration

Services access secrets via Vault:

**Deployment with Vault Agent:**
```yaml
spec:
  template:
    metadata:
      annotations:
        vault.hashicorp.com/agent-inject: "true"
        vault.hashicorp.com/role: "lma-service"
        vault.hashicorp.com/agent-inject-secret-config: "secret/data/lma/<service-name>"
```

**Vault Secret Path Convention:**
```
secret/data/lma/<service-name>/<environment>/<secret-name>

Examples:
secret/data/lma/db-guardian-service/prod/database-url
secret/data/lma/notification-service/prod/smtp-password
```

## Monitoring and Observability

### Prometheus Scraping

**Service Annotations:**
```yaml
metadata:
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "<rest-port>"
    prometheus.io/path: "/metrics"
```

### Distributed Tracing

**OpenTelemetry Collector:**
- Deployed as DaemonSet in each namespace
- Collects traces from all services
- Exports to backend (Jaeger, Tempo)

**Service Configuration:**
```yaml
env:
  - name: OTEL_EXPORTER_OTLP_ENDPOINT
    value: http://otel-collector:4318
  - name: OTEL_SERVICE_NAME
    value: <service-name>
```

## Best Practices

### 1. Resource Limits
**✅ DO:** Always set resource limits and requests
```yaml
resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 250m
    memory: 256Mi
```

**❌ DON'T:** Deploy without resource limits
```yaml
resources: {}  # Bad!
```

### 2. Security Context
**✅ DO:** Run as non-root with read-only filesystem
```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
```

### 3. Health Checks
**✅ DO:** Configure both liveness and readiness probes
```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: rest
readinessProbe:
  httpGet:
    path: /healthz
    port: rest
```

### 4. Autoscaling
**✅ DO:** Enable HPA for production
```yaml
autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 20
```

### 5. Pod Disruption Budgets
**✅ DO:** Set PDB to ensure availability during updates
```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: <service-name>
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app: <service-name>
```

## Troubleshooting

### Check Pod Status
```bash
kubectl get pods -n lma-prod -l app=<service-name>
kubectl describe pod <pod-name> -n lma-prod
```

### View Logs
```bash
# Service container logs
kubectl logs <pod-name> -n lma-prod -c <service-name>

# Envoy sidecar logs
kubectl logs <pod-name> -n lma-prod -c envoy
```

### Check Service Endpoints
```bash
kubectl get endpoints <service-name> -n lma-prod
```

### Port Forward for Local Testing
```bash
# Forward REST port
kubectl port-forward service/<service-name> 7002:7002 -n lma-prod

# Forward gRPC port
kubectl port-forward service/<service-name> 50052:50052 -n lma-prod
```

### Debug with Ephemeral Container
```bash
kubectl debug <pod-name> -n lma-prod --image=busybox -it -- sh
```

## Common Issues

### Issue: ImagePullBackOff
**Cause:** Cannot pull Docker image
**Solution:** Check image tag exists in GHCR
```bash
gh api /user/packages/container/lastmile-accelerator%2F<service>/versions
```

### Issue: CrashLoopBackOff
**Cause:** Container exits immediately
**Solution:** Check logs for errors
```bash
kubectl logs <pod-name> -n lma-prod --previous
```

### Issue: Service Not Responding
**Cause:** Readiness probe failing
**Solution:** Check health endpoint
```bash
kubectl exec <pod-name> -n lma-prod -- curl http://localhost:7002/healthz
```

## Useful Links

- **Kubernetes Docs:** https://kubernetes.io/docs
- **Helm Docs:** https://helm.sh/docs
- **Envoy Docs:** https://www.envoyproxy.io/docs
- **Vault K8s Integration:** https://www.vaultproject.io/docs/platform/k8s
- **Service Catalog:** `../service_catalog.yaml`
- **Root CLAUDE.md:** `../CLAUDE.md`
