Run security scans on the codebase.

Steps:
1. Run gitleaks to detect secrets:
   ```bash
   gitleaks detect --source . --config .github/gitleaks.toml --verbose
   ```
2. If Dockerfiles changed, run Hadolint:
   ```bash
   find . -name Dockerfile | xargs -I {} hadolint {}
   ```
3. If Kubernetes YAML changed, run OPA Conftest:
   ```bash
   conftest test services/*/helm/ --policy .github/policy/ --all-namespaces
   ```
4. Check for vulnerable dependencies:
   - **Node:** `pnpm audit` in each service
   - **Go:** `go list -json -m all | nancy sleuth` (if Nancy installed)
   - **Rust:** `cargo audit` in each service
   - **Python:** `safety check` in each service
5. Scan Docker images (if built):
   ```bash
   # For each service with Docker image
   grype <image-name>:latest --fail-on high
   ```
6. Check for common security issues:
   - Search for `eval(` usage in JS/TS
   - Search for SQL string concatenation
   - Check for missing authentication middleware
   - Verify HTTPS enforced
7. Generate security report:
   ```
   Security Scan Report
   ====================

   🔐 Secrets Scan (Gitleaks)
   ✓ No secrets detected

   🐳 Dockerfile Lint (Hadolint)
   ✓ All Dockerfiles pass best practices

   ☸️ Kubernetes Policies (OPA)
   ✓ All Helm charts pass security policies
   ⚠ Warning: 3 charts missing resource limits

   📦 Dependency Vulnerabilities
   ✓ Node services: No vulnerabilities
   ✗ Go services: 2 moderate vulnerabilities in db-guardian
      - golang.org/x/net: CVE-2023-xxxxx
      - github.com/lib/pq: CVE-2023-yyyyy
   ✓ Rust services: No vulnerabilities
   ✓ Python services: No vulnerabilities

   🐳 Container Image Scan (Grype)
   ⏭ Skipped (no images built locally)

   🔍 Code Security Patterns
   ✓ No eval() usage found
   ✓ No SQL concatenation found
   ✓ Authentication middleware present

   Recommendation:
   Address moderate vulnerabilities in Go dependencies before production deploy.
   ```

8. If critical vulnerabilities found, fail with exit code 1
9. If moderate/low vulnerabilities, warn but don't fail

Return security report summary to user.
