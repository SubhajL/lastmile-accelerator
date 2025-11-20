Commit CLAUDE.md files and sync to all worktrees.

Steps:
1. **Verify we're in the main repository:**
   ```bash
   if [ ! -d ".claude" ]; then
     echo "Error: Must run from main repository root"
     exit 1
   fi
   ```

2. **Check for uncommitted CLAUDE.md files:**
   ```bash
   git status --porcelain | grep -E "(CLAUDE\.md|\.claude/|MCP_SETUP\.md)"
   ```

3. **Stage all CLAUDE.md related files:**
   ```bash
   git add CLAUDE.md
   git add MCP_SETUP.md
   git add CLAUDE_CODE_SETUP_COMPLETE.md
   git add .claude/settings.json
   git add .claude/commands/*.md
   git add services/CLAUDE.md
   git add services/*/CLAUDE.md
   git add frontends/CLAUDE.md
   git add .github/CLAUDE.md
   git add infra/CLAUDE.md
   git add dev/CLAUDE.md
   ```

4. **Show what will be committed:**
   ```bash
   echo "Files to commit:"
   git diff --cached --name-only

   echo ""
   echo "Summary:"
   echo "  CLAUDE.md files: $(git diff --cached --name-only | grep -c CLAUDE.md)"
   echo "  Slash commands: $(git diff --cached --name-only | grep -c .claude/commands)"
   echo "  Config files: $(git diff --cached --name-only | grep -c settings.json)"
   ```

5. **Commit the files:**
   ```bash
   git commit -m "$(cat <<'EOF'
   docs: add comprehensive CLAUDE.md system for Claude Code

   - Add root CLAUDE.md with universal project rules
   - Add CLAUDE.md for all 28 services (comprehensive + streamlined)
   - Add directory-level CLAUDE.md (frontends, .github, infra, dev)
   - Configure hooks for auto-formatting and safety guards
   - Add 12 custom slash commands for workflows
   - Add MCP server setup guide

   Generated following best practices from generate-claude.md

   🤖 Generated with Claude Code
   https://claude.com/claude-code

   Co-Authored-By: Claude <noreply@anthropic.com>
   EOF
   )"
   ```

6. **Get current branch name:**
   ```bash
   MAIN_BRANCH=$(git branch --show-current)
   echo "Main branch: $MAIN_BRANCH"
   ```

7. **Push to remote:**
   ```bash
   git push origin $MAIN_BRANCH
   ```

8. **Find all worktrees:**
   ```bash
   echo ""
   echo "Syncing to worktrees..."
   echo ""

   git worktree list --porcelain | grep -E "^worktree " | cut -d' ' -f2 | while read worktree_path; do
     # Skip main repository
     if [ "$worktree_path" = "$(pwd)" ]; then
       continue
     fi

     echo "📁 Processing: $worktree_path"

     cd "$worktree_path"

     # Get worktree branch
     worktree_branch=$(git branch --show-current)
     echo "   Branch: $worktree_branch"

     # Check for uncommitted changes
     if [ -n "$(git status --porcelain)" ]; then
       echo "   ⚠️  Warning: Uncommitted changes detected"
       echo "   Creating stash..."
       git stash push -m "Auto-stash before CLAUDE.md sync"
     fi

     # Fetch latest
     git fetch origin

     # Try to merge main branch
     echo "   Merging $MAIN_BRANCH..."
     if git merge origin/$MAIN_BRANCH --no-edit; then
       echo "   ✓ Merged successfully"
     else
       echo "   ⚠️  Merge conflict detected"
       echo "   Attempting auto-resolution..."

       # For CLAUDE.md files, prefer incoming version (from main)
       git status --porcelain | grep "^UU" | grep "CLAUDE.md" | while read status file; do
         echo "      Accepting incoming: $file"
         git checkout --theirs "$file"
         git add "$file"
       done

       # For .claude/ directory, prefer incoming
       git status --porcelain | grep "^UU" | grep ".claude/" | while read status file; do
         echo "      Accepting incoming: $file"
         git checkout --theirs "$file"
         git add "$file"
       done

       # Try to complete merge
       if git commit --no-edit; then
         echo "   ✓ Conflicts resolved and committed"
       else
         echo "   ✗ Manual resolution needed"
         echo "   Run: cd $worktree_path && git status"
         git merge --abort
       fi
     fi

     # Pop stash if we created one
     if git stash list | grep -q "Auto-stash before CLAUDE.md sync"; then
       echo "   Restoring stashed changes..."
       git stash pop
     fi

     echo ""
   done
   ```

9. **Summary report:**
   ```bash
   echo "═══════════════════════════════════════"
   echo "CLAUDE.md Sync Complete!"
   echo "═══════════════════════════════════════"
   echo ""
   echo "✓ Committed to main repository ($MAIN_BRANCH)"
   echo "✓ Pushed to remote"
   echo ""
   echo "Worktree sync status:"

   git worktree list --porcelain | grep -E "^worktree " | cut -d' ' -f2 | while read worktree_path; do
     if [ "$worktree_path" = "$(pwd)" ]; then
       continue
     fi

     cd "$worktree_path"
     worktree_branch=$(git branch --show-current)
     service_name=$(basename "$worktree_path")

     if [ -f "services/$service_name/CLAUDE.md" ] || [ -f "CLAUDE.md" ]; then
       echo "  ✓ $service_name ($(basename $worktree_path))"
     else
       echo "  ⚠️  $service_name - CLAUDE.md not found"
     fi
   done

   echo ""
   echo "Next steps:"
   echo "  1. Verify CLAUDE.md files in each worktree"
   echo "  2. Run '/worktree-list' to check status"
   echo "  3. Test custom commands: '/test-service <name>'"
   echo ""
   ```

10. **Handle errors gracefully:**
    - If commit fails (nothing to commit): Skip and show message
    - If push fails: Show error, suggest pulling first
    - If worktree merge fails: Show conflict details, don't abort other worktrees
    - If worktree has uncommitted changes: Stash, merge, pop stash

**Safety checks:**
- Don't run if not in main repository
- Stash uncommitted changes before merging
- Auto-resolve CLAUDE.md conflicts (prefer incoming from main)
- Don't force push
- Show detailed status for each worktree

**Example output:**
```
Syncing CLAUDE.md system to all worktrees...

Files to commit:
  CLAUDE.md
  MCP_SETUP.md
  .claude/settings.json
  .claude/commands/worktree-list.md
  ... (48 files total)

Summary:
  CLAUDE.md files: 34
  Slash commands: 12
  Config files: 1

✓ Committed to main repository (test-lab-service)
✓ Pushed to remote

Syncing to worktrees...

📁 Processing: /Users/.../worktrees/db-guardian-service
   Branch: db-guardian-service
   Merging test-lab-service...
   ✓ Merged successfully

📁 Processing: /Users/.../worktrees/test-lab-service
   Branch: test-lab/auth-wire-jwks-plugin
   Merging test-lab-service...
   ✓ Merged successfully

... (5 more worktrees)

═══════════════════════════════════════
CLAUDE.md Sync Complete!
═══════════════════════════════════════

✓ Committed to main repository (test-lab-service)
✓ Pushed to remote

Worktree sync status:
  ✓ db-guardian-service
  ✓ test-lab-service
  ✓ dep-governance-service
  ✓ notification-service
  ✓ observability-service
  ✓ projects-service
  ✓ secrets-env-service

Next steps:
  1. Verify CLAUDE.md files in each worktree
  2. Run '/worktree-list' to check status
  3. Test custom commands: '/test-service <name>'
```
