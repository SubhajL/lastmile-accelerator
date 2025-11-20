Run tests for all services in the monorepo.

Steps:
1. Use Turborepo to run tests across all workspaces:
   ```bash
   bunx turbo run test
   ```
2. Show progress for each service
3. Collect results:
   - Services with passing tests (✓)
   - Services with failing tests (✗)
   - Services with no tests (⚠)
   - Total test count
   - Average coverage
4. If any tests fail, list which services failed
5. Offer to re-run specific failed service with verbose output

Alternative: Run tests only for services with changes:
```bash
bunx turbo run test --filter='...[HEAD^1]'
```

Example output:
```
Running tests for all services...

✓ projects-service (Node) - 32 tests passed, 84% coverage
✓ db-guardian-service (Go) - 45 tests passed, 83% coverage
✗ test-lab-service (Node) - 28/30 tests passed, 2 failed
✓ dep-governance-service (Rust) - 18 tests passed, 91% coverage
...

Summary:
  Total: 25 services tested
  Passed: 23 services
  Failed: 2 services (test-lab-service, notification-service)
  No tests: 0 services
  Average coverage: 82.4%

Re-run failed tests? [y/n]
```
