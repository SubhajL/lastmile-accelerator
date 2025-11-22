# Git Hooks for Graphite Workflow Enforcement

This directory contains Git hooks that enforce the Graphite workflow by blocking direct `git commit` and `git push` operations.

## Overview

These hooks ensure all changes go through the Graphite CLI (`gt`) for proper stack management, preventing direct git operations that could break stack linearity.

## Hooks Implemented

### 1. `pre-commit`
- **Triggers on:** `git commit`
- **Action:** Blocks the commit and displays instructions to use Graphite
- **Suggests:** Using `gt branch create` or `gt modify` instead

### 2. `pre-push`
- **Triggers on:** `git push`
- **Action:** Blocks the push and displays instructions to use Graphite
- **Suggests:** Using `gt submit` or `gt submit --update` instead
- **Smart:** Detects if PR already exists and provides appropriate guidance

### 3. `commit-msg`
- **Triggers on:** Commit message processing (if pre-commit is bypassed)
- **Action:** Secondary block to ensure commits are prevented
- **Purpose:** Catches edge cases where pre-commit might be skipped

## Setup Instructions

These hooks are already configured for this repository. The setup process was:

```bash
# 1. Create hooks directory
mkdir -p .git-hooks

# 2. Create hook scripts (pre-commit, pre-push, commit-msg)
# 3. Make hooks executable
chmod +x .git-hooks/*

# 4. Configure Git to use this hooks directory
git config core.hooksPath .git-hooks
```

## Workflow Guide

### Instead of `git commit`:
```bash
# Stage your changes
git add .

# Create new branch with commit
gt branch create feature-name --parent main

# Or modify existing branch (amend)
gt modify
```

### Instead of `git push`:
```bash
# Submit new PR
gt submit

# Update existing PR
gt submit --update

# Submit as draft
gt submit --draft
```

### Managing Your Stack:
```bash
# View stack status
gt status

# View stack graph
gt log --graph

# Navigate stack
gt upstack
gt downstack

# Sync with trunk
gt sync trunk
```

## Troubleshooting

### If hooks aren't working:
1. Check hooks path: `git config --get core.hooksPath`
2. Verify hooks are executable: `ls -la .git-hooks/`
3. Re-run setup: `git config core.hooksPath .git-hooks`

### To temporarily bypass (NOT RECOMMENDED):
```bash
# Bypass hooks (emergency only)
git commit --no-verify
git push --no-verify
```

⚠️ **WARNING:** Bypassing hooks can break stack linearity and cause issues with Graphite.

## Integration with Claude Code

These hooks work seamlessly with Claude Code's g-stack and g-submit skills:
- When a commit/push is blocked, ask Claude to use the appropriate skill
- Example: "use g-stack to create my feature branch and commit these changes"
- Example: "use g-submit to submit my changes"

## Maintenance

To update hooks across all worktrees:
```bash
# Copy updated hooks to all worktrees
for worktree in ~/dev/worktrees/*; do
    cp -r .git-hooks/* $worktree/.git-hooks/
done
```