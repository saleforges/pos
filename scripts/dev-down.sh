#!/bin/bash
# Stop all Go services + Caddy
echo "Stopping Go services..."
pkill -f "go.*cmd/(iam|merchant|catalog)" 2>/dev/null && echo "  stopped" || echo "  (none running)"

echo "Stopping Caddy..."
SCRIPT_DIR="$(cd "$(dirname "$0")"/.. && pwd)"
docker compose -f "$SCRIPT_DIR/deploy/caddy/docker-compose.yml" down 2>/dev/null

echo "✓ All services stopped"
