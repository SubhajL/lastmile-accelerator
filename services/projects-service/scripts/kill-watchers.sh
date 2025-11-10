#!/usr/bin/env bash
set -euo pipefail

PATTERN='(vitest|pnpm dev|tsx .*watch|@playwright/test/.*/test-server)'
PIDS=$(ps -Ao pid,command | grep -E "$PATTERN" | grep -v grep | awk '{print $1}')

if [ -z "${PIDS:-}" ]; then
  echo "No watcher/test processes found."
  exit 0
fi

echo "Terminating PIDs: $PIDS"
for pid in $PIDS; do
  kill -TERM "$pid" 2>/dev/null || true
done

sleep 1

PIDS2=$(ps -Ao pid,command | grep -E "$PATTERN" | grep -v grep | awk '{print $1}')
if [ -n "${PIDS2:-}" ]; then
  echo "Force killing: $PIDS2"
  for pid in $PIDS2; do
    kill -KILL "$pid" 2>/dev/null || true
  done
fi

echo "Done."
