Run tests only for services affected by recent changes.

Steps:
1. Detect changed files since last commit:
   ```bash
   git diff --name-only HEAD^1
   ```
2. Determine which services are affected based on file paths
3. Use Turborepo to run tests only for affected services:
   ```bash
   bunx turbo run test --filter='...[HEAD^1]'
   ```
4. Show which services are being tested and why
5. Display test results for affected services only
6. Skip services with no changes

Example output:
```
Detecting affected services...

Changed files:
  - services/projects-service/src/routes/projects.ts
  - services/notification-service/src/templates/email.hbs
  - services/CLAUDE.md (docs only, no tests needed)

Affected services:
  ✓ projects-service (source code changed)
  ✓ notification-service (template changed)

Running tests for 2 affected services...

✓ projects-service - 32 tests passed
✓ notification-service - 15 tests passed

All affected services passing!
```

Optimization: If no services affected, skip running tests:
```
No service source code changed. Skipping tests.
```
