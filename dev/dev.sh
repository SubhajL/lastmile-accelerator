#!/bin/bash
# LMA Development Environment Orchestrator
# Manages infrastructure + Sprint 1 services

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEVSTACK_DIR="$SCRIPT_DIR/../lma-devstack-compose-gitea4001"
PROJECT_ROOT="$SCRIPT_DIR/.."

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

usage() {
    cat <<EOF
${CYAN}LMA Development Environment${NC}

${GREEN}Usage:${NC}
  ./dev.sh <command>

${GREEN}Commands:${NC}
  ${YELLOW}infra${NC}         Start infrastructure (docker-compose)
  ${YELLOW}start${NC}         Start all services (infra + apps with hot-reload)
  ${YELLOW}stop${NC}          Stop all services
  ${YELLOW}restart${NC}       Restart all services
  ${YELLOW}status${NC}        Show status of all services
  ${YELLOW}logs${NC}          Tail logs from infrastructure
  ${YELLOW}clean${NC}         Stop everything and clean volumes
  ${YELLOW}deps${NC}          Check if dependencies are ready
  ${YELLOW}install${NC}       Install required dev tools

${GREEN}Examples:${NC}
  ./dev.sh start          # Start everything
  ./dev.sh stop           # Stop everything
  ./dev.sh logs           # View infra logs
  ./dev.sh deps           # Check if infra is ready

${GREEN}After starting:${NC}
  Services will run with hot-reload in tmux/hivemind
  Use Ctrl+C to stop, or run: ./dev.sh stop

${GREEN}Infrastructure URLs:${NC}
  Grafana:     http://localhost:3000 (admin/admin)
  Prometheus:  http://localhost:9090
  Gitea:       http://localhost:4001
  MailHog:     http://localhost:8025
  MinIO:       http://localhost:9001 (minioadmin/minioadmin)
  Vault:       http://localhost:8200 (token: lma-root)

EOF
}

check_docker() {
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}❌ Docker not found${NC}"
        echo "Install Docker Desktop: https://www.docker.com/products/docker-desktop"
        exit 1
    fi
}

check_hivemind() {
    if ! command -v hivemind &> /dev/null; then
        echo -e "${YELLOW}⚠️  hivemind not found${NC}"
        echo ""
        echo "Install with:"
        echo "  brew install hivemind"
        echo ""
        exit 1
    fi
}

start_infra() {
    echo -e "${BLUE}🚀 Starting infrastructure...${NC}"
    cd "$DEVSTACK_DIR"
    docker compose up -d
    echo -e "${GREEN}✅ Infrastructure started${NC}"
    echo ""
}

stop_infra() {
    echo -e "${BLUE}🛑 Stopping infrastructure...${NC}"
    cd "$DEVSTACK_DIR"
    docker compose stop
    echo -e "${GREEN}✅ Infrastructure stopped${NC}"
}

start_services() {
    check_hivemind
    
    echo -e "${BLUE}🔍 Checking dependencies...${NC}"
    "$SCRIPT_DIR/wait-for-deps.sh"
    
    echo ""
    echo -e "${BLUE}🚀 Starting Sprint 1 services with hot-reload...${NC}"
    echo ""
    echo -e "${YELLOW}Services:${NC}"
    echo "  • projects-service (Node.js) :7002"
    echo "  • notification-service (Node.js) :7902"
    echo "  • test-lab-service (Node.js) :7202"
    echo "  • observability-service (Go) :7301"
    echo "  • db-guardian-service (Go) :7105"
    echo "  • secrets-env-service (Go) :7104"
    echo "  • dep-governance-service (Rust) :7106"
    echo ""
    echo -e "${CYAN}Press Ctrl+C to stop all services${NC}"
    echo ""
    
    cd "$SCRIPT_DIR"
    # Source env vars and start hivemind
    export $(grep -v '^#' .env.local | xargs)
    hivemind
}

stop_services() {
    echo -e "${BLUE}🛑 Stopping services...${NC}"
    pkill -f hivemind || true
    pkill -f "tsx watch" || true
    pkill -f "air -c" || true
    pkill -f "cargo watch" || true
    echo -e "${GREEN}✅ Services stopped${NC}"
}

show_status() {
    echo -e "${CYAN}=== Infrastructure Status ===${NC}"
    cd "$DEVSTACK_DIR"
    docker compose ps
    
    echo ""
    echo -e "${CYAN}=== Service Processes ===${NC}"
    if pgrep -f hivemind > /dev/null; then
        echo -e "${GREEN}✓ hivemind is running${NC}"
        echo ""
        ps aux | grep -E "(tsx watch|air -c|cargo watch)" | grep -v grep || echo "No service processes found"
    else
        echo -e "${YELLOW}✗ No services running${NC}"
    fi
}

show_logs() {
    echo -e "${BLUE}📋 Infrastructure logs (Ctrl+C to stop)...${NC}"
    cd "$DEVSTACK_DIR"
    docker compose logs -f
}

clean_all() {
    echo -e "${RED}⚠️  This will delete all volumes (databases, etc.)${NC}"
    echo -n "Are you sure? (y/N): "
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        echo "Cancelled."
        exit 0
    fi
    
    stop_services
    echo ""
    echo -e "${BLUE}🧹 Cleaning everything...${NC}"
    cd "$DEVSTACK_DIR"
    docker compose down -v
    echo -e "${GREEN}✅ All cleaned${NC}"
}

check_deps() {
    "$SCRIPT_DIR/wait-for-deps.sh"
}

install_tools() {
    echo -e "${BLUE}🔧 Installing development tools...${NC}"
    echo ""
    
    # Check what's missing
    missing=()
    
    if ! command -v hivemind &> /dev/null; then
        missing+=("hivemind")
    fi
    
    if ! command -v air &> /dev/null; then
        missing+=("air")
    fi
    
    if ! command -v cargo-watch &> /dev/null; then
        missing+=("cargo-watch")
    fi
    
    if [ ${#missing[@]} -eq 0 ]; then
        echo -e "${GREEN}✅ All tools already installed!${NC}"
        exit 0
    fi
    
    echo -e "${YELLOW}Missing tools: ${missing[*]}${NC}"
    echo ""
    
    # Install with Homebrew (macOS)
    if command -v brew &> /dev/null; then
        for tool in "${missing[@]}"; do
            case $tool in
                hivemind)
                    echo "Installing hivemind..."
                    brew install hivemind
                    ;;
                air)
                    echo "Installing air (Go hot-reload)..."
                    go install github.com/air-verse/air@latest
                    ;;
                cargo-watch)
                    echo "Installing cargo-watch (Rust hot-reload)..."
                    cargo install cargo-watch
                    ;;
            esac
        done
        
        echo ""
        echo -e "${GREEN}✅ Tools installed!${NC}"
    else
        echo -e "${RED}Homebrew not found. Install manually:${NC}"
        echo ""
        echo "  hivemind:     brew install hivemind"
        echo "  air:          go install github.com/air-verse/air@latest"
        echo "  cargo-watch:  cargo install cargo-watch"
    fi
}

# Main command routing
case "${1:-}" in
    infra)
        check_docker
        start_infra
        ;;
    start)
        check_docker
        start_infra
        start_services
        ;;
    stop)
        stop_services
        stop_infra
        ;;
    restart)
        stop_services
        sleep 2
        start_services
        ;;
    status)
        show_status
        ;;
    logs)
        show_logs
        ;;
    clean)
        clean_all
        ;;
    deps)
        check_deps
        ;;
    install)
        install_tools
        ;;
    ""|help|-h|--help)
        usage
        ;;
    *)
        echo -e "${RED}Unknown command: $1${NC}"
        echo ""
        usage
        exit 1
        ;;
esac
