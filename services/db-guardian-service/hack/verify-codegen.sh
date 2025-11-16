#!/usr/bin/env bash
set -euo pipefail

svc_root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")"/.. && pwd)
cd "$svc_root"

echo "Regenerating gRPC code..."
make grpc-codegen

echo "Checking for uncommitted changes in generated package..."
git --no-pager diff --exit-code -- "dbguardian/v1" || {
  echo "\nCodegen is stale. Run 'make grpc-codegen' and commit changes." >&2
  exit 1
}

echo "Codegen verification passed."
