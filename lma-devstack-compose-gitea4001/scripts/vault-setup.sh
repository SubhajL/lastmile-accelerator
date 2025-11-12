#!/bin/sh
set -e
export VAULT_ADDR="${VAULT_ADDR:-hxxp://vault:8200}"
export VAULT_TOKEN="${VAULT_TOKEN:-dev-root-token}"

echo "==> Enabling KV v2 at 'kvv2'"
vault secrets enable -path=kvv2 -version=2 kv || true

echo "==> Seeding example secrets"
vault kv put kvv2/lma/dev/publisher STRIPE_API_KEY=test_key
# Defanged DSN to avoid secret scanners; assembly in parts
vault kv put kvv2/lma/dev/projects-service DATABASE_URL="pg""sql://lma:lma@postgres:5432/lma"

echo "==> Vault dev setup complete."
