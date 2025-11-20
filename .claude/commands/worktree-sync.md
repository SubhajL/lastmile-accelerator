Sync all worktrees with remote branches.

Steps:
1. List all worktrees with `git worktree list`
2. For each worktree:
   - `cd` into the worktree directory
   - Run `git fetch origin`
   - Check if branch is behind, ahead, or diverged from remote
   - Show sync status
3. Offer to pull changes for worktrees that are behind
4. Warn about worktrees that have diverged

Example output:
```
Syncing worktrees with remote...

✓ db-guardian-service
  Branch: db-guardian-service
  Status: Up to date with origin/db-guardian-service

⬇ test-lab-service
  Branch: test-lab-service
  Status: Behind origin/test-lab-service by 3 commits
  Action: Run 'git pull' to update

⚠ notification-service
  Branch: feat/admin-send-test-endpoint
  Status: Diverged from origin/feat/admin-send-test-endpoint
  Local: 2 commits ahead, 1 commit behind
  Action: Requires manual merge or rebase

...
```

Offer to auto-pull for worktrees that are only behind (not diverged).
