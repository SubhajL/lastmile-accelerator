List all git worktrees and their current branches:

1. Run `git worktree list` to show all worktrees
2. For each worktree, show:
   - Path
   - Current branch
   - Last commit (short SHA and message)
3. Highlight which worktree is the current directory
4. Show if any worktrees have uncommitted changes

Example output format:
```
📁 Main Repository
   Path: /Users/.../lastmile-accelerator
   Branch: test-lab-service
   Commit: e5fb277 - test(test-lab-service): defang secret scanner
   Status: ✓ Clean

📁 Worktree: db-guardian-service
   Path: /Users/.../worktrees/db-guardian-service
   Branch: db-guardian-service
   Commit: 779be4e - feat(db): add migration analysis
   Status: ⚠ Modified files (3)

...
```
