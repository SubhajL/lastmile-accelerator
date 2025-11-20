Generate Software Bill of Materials (SBOM) for a service.

Usage: `/sbom-generate <service-name>`

Steps:
1. Parse service name from $ARGUMENTS
2. Find service directory: `services/<service-name>`
3. Determine service language from service_catalog.yaml
4. Generate SBOM based on language:

   **Node/TypeScript:**
   ```bash
   cd services/<service-name>
   # Using CycloneDX
   npx @cyclonedx/cyclonedx-npm --output-file sbom.json
   # Or using Syft
   syft packages . -o spdx-json > sbom-spdx.json
   ```

   **Go:**
   ```bash
   cd services/<service-name>
   syft packages . -o cyclonedx-json > sbom.json
   ```

   **Rust:**
   ```bash
   cd services/<service-name>
   cargo sbom > sbom.json
   # Or using Syft
   syft packages . -o spdx-json > sbom-spdx.json
   ```

   **Python:**
   ```bash
   cd services/<service-name>
   pip freeze | cyclonedx-py -o sbom.json
   ```

5. Scan SBOM for vulnerabilities with Grype:
   ```bash
   grype sbom:./sbom.json --fail-on critical
   ```

6. Generate human-readable summary:
   ```
   SBOM Generated for db-guardian-service
   ======================================

   Total Dependencies: 87
   - Direct: 23
   - Transitive: 64

   By License:
   - MIT: 45
   - Apache-2.0: 25
   - BSD-3-Clause: 12
   - Other: 5

   Vulnerability Scan:
   - Critical: 0
   - High: 2
   - Medium: 5
   - Low: 12

   High Vulnerabilities:
   1. golang.org/x/net v0.10.0
      CVE: CVE-2023-39325
      Severity: HIGH
      Fixed in: v0.17.0

   2. github.com/lib/pq v1.10.7
      CVE: CVE-2023-5043
      Severity: HIGH
      Fixed in: v1.10.9

   Files:
   - SBOM: services/db-guardian-service/sbom.json (CycloneDX)
   - Vulnerability Report: services/db-guardian-service/sbom-vulnerabilities.txt

   Recommendation:
   Update dependencies to fix high-severity vulnerabilities.
   ```

7. Save SBOM to service directory: `services/<service-name>/sbom.json`
8. Optionally upload to dependency tracking system
9. Return summary with file paths

SBOM formats supported:
- CycloneDX JSON (default)
- SPDX JSON
- SPDX Tag-Value

Security scanning is automatic with SBOM generation.
