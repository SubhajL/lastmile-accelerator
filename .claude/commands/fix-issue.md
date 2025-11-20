Fetch a GitHub issue, implement a fix, and create a pull request.

Usage: `/fix-issue <issue-number>`

Steps:
1. Fetch issue details using GitHub CLI:
   ```bash
   gh issue view $ARGUMENTS --json title,body,labels,assignees
   ```
2. Analyze the issue:
   - Understand the problem
   - Identify affected service(s)
   - Determine required changes
3. Search codebase for relevant files:
   - Use `rg` to find related code
   - Read service CLAUDE.md for patterns
4. Create feature branch:
   ```bash
   git checkout -b fix/issue-$ARGUMENTS
   ```
5. Implement the fix:
   - Follow TDD: write failing test first
   - Implement fix
   - Ensure tests pass
   - Follow service's CLAUDE.md patterns
6. Run quality gates:
   - Type checking: `pnpm typecheck` or `make test`
   - Linting: `pnpm lint` or `make lint`
   - Tests: `pnpm test` or `make test`
   - Build: `pnpm build` or `make build`
7. Commit with conventional commit message:
   ```bash
   git add .
   git commit -m "fix(service-name): <issue title>

   Fixes #$ARGUMENTS

   - Detail about what was changed
   - Why this fixes the issue
   "
   ```
8. Push branch:
   ```bash
   git push -u origin fix/issue-$ARGUMENTS
   ```
9. Create PR using GitHub CLI:
   ```bash
   gh pr create --title "Fix: <issue title>" --body "Fixes #$ARGUMENTS

   ## Summary
   <explanation>

   ## Test Plan
   - [ ] Unit tests pass
   - [ ] Integration tests pass
   - [ ] Manual testing completed

   ## Checklist
   - [x] Tests added/updated
   - [x] Documentation updated
   - [x] Follows CLAUDE.md patterns
   "
   ```
10. Return PR URL to user

Remember:
- Follow existing code patterns from service's CLAUDE.md
- Write tests before implementation (TDD)
- Run all quality gates before committing
- Use conventional commit format
- Link PR to issue with "Fixes #<issue-number>"
