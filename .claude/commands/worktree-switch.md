Switch to a service worktree by name.

Usage: `/worktree-switch <service-name>`

Steps:
1. Parse service name from $ARGUMENTS
2. Check if worktree exists at `~/dev/worktrees/<service-name>`
3. If worktree doesn't exist, offer to create it:
   - Ask for branch name (default: `<service-name>`)
   - Run `git worktree add ~/dev/worktrees/<service-name> -b <branch-name>`
4. Change directory to the worktree: `cd ~/dev/worktrees/<service-name>`
5. Show status:
   - Current branch
   - Last commit
   - Uncommitted changes
   - Service info from service_catalog.yaml (ports, language)

Example:
```
$ /worktree-switch db-guardian

✓ Switched to worktree: db-guardian-service
  Path: ~/dev/worktrees/db-guardian-service
  Branch: db-guardian-service
  Service: db-guardian-service (Go)
  Ports: REST 7105, gRPC 50065
  Status: 2 modified files

Ready to work on db-guardian-service!
```
