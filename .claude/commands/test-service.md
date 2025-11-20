Run tests for a specific service.

Usage: `/test-service <service-name>`

Steps:
1. Parse service name from $ARGUMENTS
2. Find service directory: `services/<service-name>`
3. Read `service_catalog.yaml` to determine language
4. Run language-appropriate test command:
   - **Node:** `cd services/<service-name> && pnpm test`
   - **Go:** `cd services/<service-name> && make test`
   - **Rust:** `cd services/<service-name> && cargo test --all --locked`
   - **Python:** `cd services/<service-name> && pytest -q`
5. Show test results summary:
   - Pass/fail count
   - Coverage percentage (if available)
   - Failed tests details
6. If tests fail, offer to show full output or specific failed test

Example:
```
$ /test-service db-guardian

Running tests for db-guardian-service (Go)...

✓ 45 tests passed
✗ 2 tests failed
Coverage: 83.2%

Failed tests:
  - TestMigrationValidator_ValidateUnsafeOperation
  - TestIndexAdvisor_RecommendIndexes

Run with -v for full output? [y/n]
```
