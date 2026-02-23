Perform a comprehensive security and quality code review of a pull request.

Usage: `/review-pr <pr-number>`

Steps:
1. Fetch PR details:
   ```bash
   gh pr view $ARGUMENTS --json title,body,files,commits,reviews
   ```
2. Checkout PR branch:
   ```bash
   gh pr checkout $ARGUMENTS
   ```
3. Review changed files systematically:

   **A. Code Quality Review:**
   - Check adherence to CLAUDE.md patterns for affected services
   - Verify conventional commit messages
   - Check for code duplication
   - Ensure proper error handling
   - Verify TypeScript/Go/Rust type safety
   - Check for commented-out code

   **B. Security Review:**
   - Run gitleaks: `gitleaks detect --source . --log-level info`
   - Check for hardcoded credentials, API keys, tokens
   - Verify input validation and sanitization
   - Check for SQL injection vulnerabilities
   - Review authentication/authorization changes
   - Check for XSS vulnerabilities
   - Verify secrets are in Vault, not committed

   **C. Testing Review:**
   - Check if new features have tests
   - Verify test coverage for changes
   - Check if tests follow TDD patterns
   - Run tests: `bunx turbo run test --filter=<affected-services>`
   - Check for flaky tests

   **D. Performance Review:**
   - Check for N+1 query problems
   - Verify database indexes for new queries
   - Check for inefficient loops or algorithms
   - Review caching strategy
   - Check for potential memory leaks

   **E. Documentation Review:**
   - Check if CLAUDE.md updated for new patterns
   - Verify CONTEXT.md updated if service config changed
   - Check if API endpoints documented
   - Verify environment variables documented

4. Run automated checks:
   ```bash
   # Type checking
   bunx turbo run typecheck --filter=<affected>

   # Linting
   bunx turbo run lint --filter=<affected>

   # Tests
   bunx turbo run test --filter=<affected>

   # Build
   bunx turbo run build --filter=<affected>

   # Security scanning
   gitleaks detect --source .

   # OPA policy checks (if K8s changes)
   conftest test services/*/helm/ --policy .github/policy/
   ```

5. Generate review report with:
   - **✅ Approved items:** What looks good
   - **⚠️ Warnings:** Minor issues or suggestions
   - **❌ Blocking issues:** Must be fixed before merge
   - **💡 Suggestions:** Optional improvements

6. Post review as PR comment:
   ```bash
   gh pr comment $ARGUMENTS --body "<review-report>"
   ```

Example review report:
```markdown
## Code Review Report

### ✅ Approved
- Code follows CLAUDE.md patterns for projects-service
- Comprehensive test coverage (92%)
- Proper error handling
- No security vulnerabilities detected

### ⚠️ Warnings
- Consider adding database index on `users.email` for performance
- Function `processProjects` is 65 lines (recommended max 50)

### ❌ Blocking Issues
- Missing test for `deleteProject` endpoint
- Hardcoded database URL in config file (should use env var)
- Missing input validation for project name

### 💡 Suggestions
- Consider extracting `validateProjectData` into shared util
- Add JSDoc comments for public API functions

### Automated Checks
- ✓ Type checking passed
- ✓ Linting passed
- ✗ Tests failed (2/45)
- ✓ Build succeeded
- ✓ No secrets detected

### Recommendation
**Request Changes** - Please address blocking issues before merge.
```

Return summary to user with link to full review comment.
