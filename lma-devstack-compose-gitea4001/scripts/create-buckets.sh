#!/bin/sh
set -e

MC_HOST_ALIAS="dev"
MC_ALIAS_URL="http://minio:9000"
MC_ALIAS_USER="${MINIO_ROOT_USER:-minioadmin}"
MC_ALIAS_PASS="${MINIO_ROOT_PASSWORD:-minioadmin}"

echo "==> Waiting 3s for MinIO..."; sleep 3
mc alias set $MC_HOST_ALIAS $MC_ALIAS_URL $MC_ALIAS_USER $MC_ALIAS_PASS

for b in snapshots artifacts templates projects results backups; do
  echo "==> Ensuring bucket: $b"
  mc mb -p $MC_HOST_ALIAS/$b || true
  mc anonymous set download $MC_HOST_ALIAS/$b || true
done

echo "==> MinIO buckets ready."
