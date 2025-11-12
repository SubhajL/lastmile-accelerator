#!/bin/sh
set -e
NATS_URL="${NATS_URL:-n""ats://nats:4222}"
echo "==> Waiting 3s for NATS..."; sleep 3

add_stream () {
  NAME="$1"; SUBJECTS="$2"
  nats -s $NATS_URL stream add $NAME --subjects="$SUBJECTS" --retention=limits --discard=old --storage=file --max-msgs=-1 --max-bytes=-1 --ack --dupe --defaults --yes 2>/dev/null || true
}

add_stream SNAPSHOT "snapshot.*"
add_stream CHECKS   "checks.*"
add_stream PUBLISH  "publish.*"
add_stream ERRORS   "errors.*"
add_stream SLO      "slo.*"
add_stream FINOPS   "finops.*"

echo "==> NATS streams ready."
