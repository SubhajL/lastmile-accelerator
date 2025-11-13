#!/usr/bin/env bash
set -euo pipefail

# Requires: gh CLI authenticated with a token that has admin:repo_hook/repo permissions.
# Usage: ./github/scripts/protect-db-guardian.sh [owner] [repo]

OWNER=${1:-$(git config --get remote.origin.url | sed -E 's#.*github.com[:/ ]([^/]+)/.*#\1#')}
REPO=${2:-$(git config --get remote.origin.url | sed -E 's#.*github.com[:/ ][^/]+/([^\.]+)(\.git)?#\1#')}
BRANCH="db-guardian-service"

echo "Setting branch protection on ${OWNER}/${REPO}@${BRANCH}..."

gh api \
  -X PUT \
  -H "Accept: application/vnd.github+json" \
  "/repos/${OWNER}/${REPO}/branches/${BRANCH}/protection" \
  -f required_status_checks.strict=true \
  -f required_status_checks.contexts[]="CI / ci" \
  -f enforce_admins=true \
  -f required_pull_request_reviews.required_approving_review_count=1 \
  -f restrictions='' \
  >/dev/null

echo "Branch protection configured."
