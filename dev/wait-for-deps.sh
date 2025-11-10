#!/bin/bash
# Wait for infrastructure dependencies to be ready
# Run this before starting services

set -e

DEVSTACK_DIR="../lma-devstack-compose-gitea4001"
TIMEOUT=60
INTERVAL=2

echo "🔍 Checking infrastructure dependencies..."
echo ""

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

wait_for_service() {
    local name=$1
    local host=$2
    local port=$3
    local timeout=$4
    
    echo -n "⏳ Waiting for $name ($host:$port)... "
    
    elapsed=0
    while ! nc -z $host $port >/dev/null 2>&1; do
        if [ $elapsed -ge $timeout ]; then
            echo -e "${RED}✗ TIMEOUT${NC}"
            return 1
        fi
        sleep $INTERVAL
        elapsed=$((elapsed + INTERVAL))
    done
    
    echo -e "${GREEN}✓${NC}"
    return 0
}

wait_for_http() {
    local name=$1
    local url=$2
    local timeout=$3
    
    echo -n "⏳ Waiting for $name ($url)... "
    
    elapsed=0
    while ! curl -sf $url >/dev/null 2>&1; do
        if [ $elapsed -ge $timeout ]; then
            echo -e "${RED}✗ TIMEOUT${NC}"
            return 1
        fi
        sleep $INTERVAL
        elapsed=$((elapsed + INTERVAL))
    done
    
    echo -e "${GREEN}✓${NC}"
    return 0
}

wait_for_postgres() {
echo -n "⏳ Waiting for PostgreSQL (localhost:55432)... "
    
    elapsed=0
while ! pg_isready -h localhost -p 55432 -U postgres >/dev/null 2>&1; do
        if [ $elapsed -ge $TIMEOUT ]; then
            echo -e "${RED}✗ TIMEOUT${NC}"
            return 1
        fi
        sleep $INTERVAL
        elapsed=$((elapsed + INTERVAL))
    done
    
    echo -e "${GREEN}✓${NC}"
    return 0
}

# Check if devstack is running
if ! docker ps | grep -q lma-postgres; then
    echo -e "${YELLOW}⚠️  DevStack doesn't appear to be running${NC}"
    echo ""
    echo "Start it with:"
    echo "  cd $DEVSTACK_DIR && docker compose up -d"
    echo ""
    exit 1
fi

# Wait for each service
failed=0

# Core infrastructure
wait_for_postgres || failed=1
wait_for_service "Redis" localhost 4050 $TIMEOUT || failed=1
wait_for_service "MinIO" localhost 9000 $TIMEOUT || failed=1
wait_for_service "NATS" localhost 4222 $TIMEOUT || failed=1

# Auth & Secrets
wait_for_http "Vault" "http://localhost:8200/v1/sys/health" $TIMEOUT || failed=1
wait_for_http "Keycloak" "http://localhost:8080" $TIMEOUT || failed=1

# Observability
wait_for_http "OpenTelemetry Collector" "http://localhost:4318" $TIMEOUT || failed=1
wait_for_http "Prometheus" "http://localhost:9090/-/healthy" $TIMEOUT || failed=1
wait_for_http "Grafana" "http://localhost:3000/api/health" $TIMEOUT || failed=1
wait_for_http "Loki" "http://localhost:3100/ready" $TIMEOUT || failed=1

# Dev tools
wait_for_http "Gitea" "http://localhost:4001" $TIMEOUT || failed=1
wait_for_http "MailHog" "http://localhost:8025" $TIMEOUT || failed=1

# Optional (may not be fully started)
wait_for_service "ClamAV" localhost 3310 10 || echo -e "${YELLOW}⚠️  ClamAV not ready (optional)${NC}"

echo ""

if [ $failed -eq 0 ]; then
    echo -e "${GREEN}✅ All required dependencies are ready!${NC}"
    echo ""
    echo "You can now start services with:"
    echo "  hivemind     # or: ./dev.sh start"
    echo ""
    exit 0
else
    echo -e "${RED}❌ Some dependencies failed to start${NC}"
    echo ""
    echo "Check docker compose logs:"
    echo "  cd $DEVSTACK_DIR && docker compose logs -f"
    echo ""
    exit 1
fi
