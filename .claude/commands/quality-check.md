Run all quality gates (typecheck, lint, test, build) for the repository or specific service.

Usage: `/quality-check [service-name]`

Steps:
1. If service name provided in $ARGUMENTS:
   - Run quality checks for specific service
   - Use service-specific commands based on language
2. If no arguments:
   - Run quality checks for all services using Turbo
3. Run checks in order (fail fast):

   **A. Type Checking**
   ```bash
   # All services
   bunx turbo run typecheck

   # Specific service (Node)
   cd services/<service> && pnpm typecheck

   # Specific service (Go)
   cd services/<service> && go vet ./...

   # Specific service (Rust)
   cd services/<service> && cargo check
   ```

   **B. Linting**
   ```bash
   # All services
   bunx turbo run lint

   # Specific service (Node)
   cd services/<service> && pnpm lint

   # Specific service (Go)
   cd services/<service> && golangci-lint run

   # Specific service (Rust)
   cd services/<service> && cargo clippy
   ```

   **C. Tests**
   ```bash
   # All services
   bunx turbo run test

   # Specific service (Node)
   cd services/<service> && pnpm test

   # Specific service (Go)
   cd services/<service> && make test

   # Specific service (Rust)
   cd services/<service> && cargo test --all --locked
   ```

   **D. Build**
   ```bash
   # All services
   bunx turbo run build

   # Specific service (Node)
   cd services/<service> && pnpm build

   # Specific service (Go)
   cd services/<service> && make build

   # Specific service (Rust)
   cd services/<service> && cargo build --release
   ```

4. Track progress and show results:
   ```
   Running quality checks...

   [1/4] Type checking... ✓ (3.2s)
   [2/4] Linting...       ✓ (2.1s)
   [3/4] Tests...         ✗ (15.3s) - 2 tests failed
   [4/4] Build...         ⏭ Skipped due to test failures

   Summary:
   ✓ Type checking passed
   ✓ Linting passed
   ✗ Tests failed (45/47 passed)
   ⏭ Build skipped

   Failed tests:
   - projects-service: TestCreateProject_Validation
   - test-lab-service: TestGenerateTests_EmptyRepo

   Fix test failures before creating PR.
   ```

5. Return status code:
   - 0: All checks passed
   - 1: One or more checks failed

Optional flags (parse from $ARGUMENTS):
- `--no-build`: Skip build step
- `--parallel`: Run checks in parallel (faster but less clear output)
- `--verbose`: Show full command output

Example:
```
$ /quality-check projects-service

Running quality checks for projects-service (Node)...

✓ Type checking passed (2.1s)
✓ Linting passed (1.3s)
✓ Tests passed - 32/32 (8.7s)
✓ Build succeeded (5.2s)

All checks passed! ✓
```
