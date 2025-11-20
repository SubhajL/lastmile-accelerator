# .github - CI/CD and GitHub Configuration

**Parent Context:** This extends [../CLAUDE.md](../CLAUDE.md)

This directory contains GitHub-specific configuration including CI/CD workflows, security policies, issue templates, and GitHub Actions.

## Directory Structure

```
.github/
├── workflows/              # GitHub Actions workflows
│   ├── _service-ci.yml     # Reusable service CI template
│   ├── _frontend-ci.yml    # Reusable frontend CI template
│   ├── ci-*.yml            # Service-specific CI workflows (28 files)
│   └── pr-ci.yml           # Pull request checks
├── policy/                 # OPA security policies
│   └── root.rego           # Conftest policy for K8s resources
├── gitleaks.toml           # Secret scanning configuration
├── ISSUE_TEMPLATE/         # Issue templates (if any)
├── PULL_REQUEST_TEMPLATE.md # PR template (if any)
└── dependabot.yml          # Dependabot configuration (if any)
```

## CI/CD Architecture

### Workflow Strategy

**Reusable Workflows:**
- `_service-ci.yml` - Template for all backend services (28 services)
- `_frontend-ci.yml` - Template for all frontend apps (3 apps)

**Per-Service Workflows:**
- Each service has its own `ci-{service-name}.yml`
- Inherits from reusable template
- Passes service-specific parameters (language, directory, ports)

### Trigger Strategy

**Pull Requests:**
- Triggered on: PR creation, PR updates
- Runs: Tests, linting, type checking, security scans
- **Does NOT** push Docker images

**Push to Main:**
- Triggered on: Merge to `main` branch
- Runs: Full CI pipeline
- **DOES** build and push Docker images to GitHub Container Registry (GHCR)

**Path Filters:**
Each workflow only runs when files in the service directory change:
```yaml
on:
  pull_request:
    paths:
      - 'services/db-guardian-service/**'
      - '.github/workflows/ci-db-guardian-service.yml'
```

## Reusable Service CI Template

**File:** `workflows/_service-ci.yml`

### CI Gates (All Services)

**1. Language-Specific Testing:**
```yaml
# Node.js/TypeScript
- run: bun run typecheck
- run: bun run test

# Go
- run: go test ./... -cover

# Rust
- run: cargo test --all --locked

# Python
- run: pytest -q
```

**2. Security Scanning:**
```yaml
# Gitleaks - Secret scanning
- uses: gitleaks/gitleaks-action@v2
  with:
    config-path: .github/gitleaks.toml

# Hadolint - Dockerfile linting
- uses: hadolint/hadolint-action@v3
  with:
    dockerfile: services/${{ inputs.service_dir }}/Dockerfile

# OPA Conftest - K8s policy checks
- run: conftest test services/${{ inputs.service_dir }}/helm/ --policy .github/policy/
```

**3. Docker Build & Publish:**
```yaml
# Build image (always)
- run: docker build -t ${{ env.IMAGE_NAME }}:${{ github.sha }} .

# Push image (only on push to main)
- if: github.event_name == 'push'
  run: |
    echo "${{ secrets.GITHUB_TOKEN }}" | docker login ghcr.io -u ${{ github.actor }} --password-stdin
    docker tag ${{ env.IMAGE_NAME }}:${{ github.sha }} ${{ env.IMAGE_NAME }}:latest
    docker push ${{ env.IMAGE_NAME }}:${{ github.sha }}
    docker push ${{ env.IMAGE_NAME }}:latest
```

**4. Container Security (on push to main):**
```yaml
# Syft - Generate SBOM
- uses: anchore/sbom-action@v0
  with:
    image: ${{ env.IMAGE_NAME }}:${{ github.sha }}
    format: spdx-json

# Grype - Vulnerability scanning
- uses: anchore/scan-action@v3
  with:
    image: ${{ env.IMAGE_NAME }}:${{ github.sha }}
    fail-build: true
    severity-cutoff: high
```

### Example Service Workflow

**File:** `workflows/ci-db-guardian-service.yml`

```yaml
name: CI - DB Guardian Service

on:
  pull_request:
    paths:
      - 'services/db-guardian-service/**'
      - '.github/workflows/ci-db-guardian-service.yml'
  push:
    branches: [main]
    paths:
      - 'services/db-guardian-service/**'

jobs:
  ci:
    uses: ./.github/workflows/_service-ci.yml
    with:
      service_name: db-guardian-service
      service_dir: services/db-guardian-service
      language: go
      rest_port: 7105
      grpc_port: 50065
```

## Frontend CI Template

**File:** `workflows/_frontend-ci.yml`

### CI Gates (All Frontends)

```yaml
# Install dependencies
- run: pnpm install --frozen-lockfile

# Type checking (allowed to fail for now)
- run: pnpm typecheck || true

# Unit tests (allowed to fail for now)
- run: pnpm test || true

# Playwright E2E smoke tests
- run: pnpm test:e2e

# Build for production
- run: pnpm build

# Docker build & push
- run: docker build -t ${{ env.IMAGE_NAME }}:${{ github.sha }} .
- if: github.event_name == 'push'
  run: docker push ${{ env.IMAGE_NAME }}:${{ github.sha }}
```

## Security Configuration

### Gitleaks (Secret Scanning)

**File:** `gitleaks.toml`

**What it scans:**
- Git commits and history
- Staged files
- All file types

**What it catches:**
- JWT tokens
- API keys (Stripe, AWS, GitHub, etc.)
- Database connection strings with passwords
- OAuth client secrets
- SSH private keys

**Allowlist patterns:**
```toml
[allowlist]
paths = [
  '''.*_test\.go''',
  '''.*\.spec\.ts''',
  '''.*\.md''',
  '''example\.env''',
]

regexes = [
  '''http://localhost''',
  '''postgresql://.*@localhost''',
]
```

**Custom Rules:**
```toml
[[rules]]
id = "jwt-token"
description = "Detected JWT token"
regex = '''eyJ[A-Za-z0-9-_]+\.eyJ[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+'''
```

### OPA Policies (Conftest)

**File:** `policy/root.rego`

**Kubernetes Resource Policies:**

```rego
package main

# Deny LoadBalancer services without source IP restrictions
deny[msg] {
  input.kind == "Service"
  input.spec.type == "LoadBalancer"
  not input.spec.loadBalancerSourceRanges
  msg = "LoadBalancer services must specify loadBalancerSourceRanges"
}

# Deny containers running as root
deny[msg] {
  input.kind == "Deployment"
  container := input.spec.template.spec.containers[_]
  not container.securityContext.runAsNonRoot
  msg = sprintf("Container %s must run as non-root", [container.name])
}

# Deny :latest image tag
deny[msg] {
  input.kind == "Deployment"
  container := input.spec.template.spec.containers[_]
  endswith(container.image, ":latest")
  msg = sprintf("Container %s uses :latest tag (forbidden)", [container.name])
}

# Warn about missing resource limits
warn[msg] {
  input.kind == "Deployment"
  container := input.spec.template.spec.containers[_]
  not container.resources.limits
  msg = sprintf("Container %s should define resource limits", [container.name])
}
```

**Testing Policies:**
```bash
# Test all policies
conftest test services/*/helm/ --policy .github/policy/

# Test specific service
conftest test services/db-guardian-service/helm/ --policy .github/policy/

# Show all warnings
conftest test --all-namespaces services/*/helm/ --policy .github/policy/
```

## Workflow Best Practices

### 1. Adding a New Service Workflow

**Steps:**
1. Copy existing service workflow template
2. Update `service_name`, `service_dir`, `language`, ports
3. Add path filters to trigger on service changes only
4. Test workflow on feature branch before merging

**Example:**
```yaml
name: CI - My New Service

on:
  pull_request:
    paths:
      - 'services/my-new-service/**'
      - '.github/workflows/ci-my-new-service.yml'
  push:
    branches: [main]
    paths:
      - 'services/my-new-service/**'

jobs:
  ci:
    uses: ./.github/workflows/_service-ci.yml
    with:
      service_name: my-new-service
      service_dir: services/my-new-service
      language: node  # or go, rust, python
      rest_port: 7XXX
      grpc_port: 50XXX
```

### 2. Modifying Reusable Workflows

**⚠️ IMPORTANT:** Changes to `_service-ci.yml` or `_frontend-ci.yml` affect **ALL** services/frontends.

**Before modifying:**
- Test changes on a single service workflow first
- Ensure backward compatibility
- Document breaking changes
- Coordinate with team

**After modifying:**
- Monitor CI runs across all services
- Fix any broken workflows immediately

### 3. Security Scanning Best Practices

**Gitleaks:**
- Run locally before committing: `gitleaks detect --source .`
- Add false positives to allowlist in `gitleaks.toml`
- Never commit real secrets, even in tests

**Hadolint:**
- Run locally: `hadolint services/*/Dockerfile`
- Fix warnings before creating PR
- Add `.hadolint.yaml` for service-specific rules if needed

**OPA Conftest:**
- Run locally: `conftest test services/my-service/helm/ --policy .github/policy/`
- Add service-specific policies in `policy/` if needed
- Use `warn` for best practices, `deny` for hard requirements

### 4. Container Image Management

**Image Naming:**
```
ghcr.io/subhajl/lastmile-accelerator/<service-name>:<tag>

Examples:
ghcr.io/subhajl/lastmile-accelerator/db-guardian-service:abc1234
ghcr.io/subhajl/lastmile-accelerator/db-guardian-service:latest
ghcr.io/subhajl/lastmile-accelerator/mode-a-dashboard:def5678
```

**Image Tags:**
- Commit SHA (short): `abc1234` - Pushed on every merge to main
- `latest` - Also pushed on merge to main
- PR builds: Built but NOT pushed

**Cleaning Up Images:**
```bash
# List images
gh api /user/packages/container/lastmile-accelerator%2Fdb-guardian-service/versions

# Delete old images (manual or via retention policy)
gh api --method DELETE /user/packages/container/lastmile-accelerator%2F<service>/versions/<version-id>
```

## Quick Commands

### Run CI Locally

**Gitleaks:**
```bash
gitleaks detect --source . --config .github/gitleaks.toml
```

**Hadolint:**
```bash
hadolint services/db-guardian-service/Dockerfile
```

**OPA Conftest:**
```bash
conftest test services/db-guardian-service/helm/ --policy .github/policy/
```

**Syft (SBOM):**
```bash
syft packages <image> -o spdx-json
```

**Grype (Vulnerability Scan):**
```bash
grype <image> --fail-on high
```

### Trigger Workflow Manually

**Using GitHub CLI:**
```bash
gh workflow run "CI - DB Guardian Service"
```

**Using GitHub UI:**
1. Go to Actions tab
2. Select workflow
3. Click "Run workflow"

### View Workflow Runs

```bash
# List recent runs
gh run list --workflow="CI - DB Guardian Service"

# View specific run
gh run view <run-id>

# Watch run in real-time
gh run watch <run-id>
```

## Common Issues

### Issue: Workflow not triggering
**Solution:** Check path filters match your changes
```yaml
paths:
  - 'services/my-service/**'  # Must include trailing /**
```

### Issue: Gitleaks false positive
**Solution:** Add to allowlist in `gitleaks.toml`
```toml
[allowlist]
regexes = [
  '''your-pattern-here''',
]
```

### Issue: OPA policy blocking valid resource
**Solution:** Add exception in `policy/root.rego` or change policy
```rego
# Add condition to exclude specific cases
deny[msg] {
  input.kind == "Service"
  input.metadata.name != "legacy-service"  # Exception
  # ... rest of rule
}
```

### Issue: Docker build out of space
**Solution:** Clean up Docker cache in CI
```yaml
- name: Clean Docker
  run: docker system prune -af
```

## Related Documentation

- **Workflow Syntax:** https://docs.github.com/actions/reference/workflow-syntax-for-github-actions
- **Reusable Workflows:** https://docs.github.com/actions/using-workflows/reusing-workflows
- **Gitleaks:** https://github.com/gitleaks/gitleaks
- **Hadolint:** https://github.com/hadolint/hadolint
- **Conftest:** https://www.conftest.dev/
- **Syft:** https://github.com/anchore/syft
- **Grype:** https://github.com/anchore/grype

## Useful Links

- **Service Catalog:** `../service_catalog.yaml`
- **Root CLAUDE.md:** `../CLAUDE.md`
- **GitHub Actions:** https://github.com/SubhajL/lastmile-accelerator/actions
