Create a pull request with a comprehensive description.

Usage: `/pr-create <title>`

Steps:
1. Verify current branch is not main/master
2. Verify branch has commits to push
3. Get current branch name: `git branch --show-current`
4. Analyze changes:
   ```bash
   git diff main...HEAD --name-status
   git log main..HEAD --oneline
   ```
5. Determine affected services from changed files
6. For each affected service, check:
   - Run tests: MUST pass
   - Type checking: MUST pass
   - Linting: SHOULD pass (warn if fails)
   - Build: MUST succeed
7. Generate PR description with:
   - Summary of changes (from commits)
   - Affected services
   - Test plan checklist
   - Screenshots (if UI changes)
   - Breaking changes (if any)
   - Migration notes (if DB changes)
8. Push branch if not already pushed:
   ```bash
   git push -u origin $(git branch --show-current)
   ```
9. Create PR:
   ```bash
   gh pr create --title "$ARGUMENTS" --body "$(cat <<'EOF'
   ## Summary
   <AI-generated summary from commits>

   ## Affected Services
   - projects-service: <change description>
   - notification-service: <change description>

   ## Test Plan
   - [ ] Unit tests pass (32/32)
   - [ ] Integration tests pass (8/8)
   - [ ] Manual testing completed
   - [ ] Tested with real data
   - [ ] Edge cases covered

   ## Quality Checks
   - [x] Type checking passed
   - [x] Linting passed
   - [x] Tests pass
   - [x] Build succeeds
   - [x] No secrets committed
   - [x] Follows CLAUDE.md patterns

   ## Breaking Changes
   <List any breaking changes, or "None">

   ## Migration Notes
   <DB migrations or config changes, or "None">

   ## Screenshots
   <If UI changes, add screenshots>

   ---
   🤖 Generated with [Claude Code](https://claude.com/claude-code)

   Co-Authored-By: Claude <noreply@anthropic.com>
   EOF
   )"
   ```
10. Return PR URL to user

Pre-flight checks:
- ✅ All tests pass
- ✅ Type checking passes
- ✅ No secrets in changes (gitleaks)
- ✅ Commits follow conventional format
- ⚠️ Warn if no tests added for new code

If any critical checks fail, ask user if they want to proceed anyway.
