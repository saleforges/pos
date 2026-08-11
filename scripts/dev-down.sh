#!/bin/bash
# Stop all Go services + Caddy
echo "Stopping Go services..."
pkill -f "go.*cmd/(iam|merchant|catalog|inventory|order|payment)" 2>/dev/null
# child binaries spawned by `go run` outlive the parent — kill them too
for p in 8080 8081 8082 8083 8084 8085; do
	pid=$(ss -tlnp 2>/dev/null | grep ":$p " | grep -oP 'pid=\K[0-9]+' | head -1)
	[ -n "$pid" ] && kill "$pid" 2>/dev/null
done
sleep 1
echo "  stopped"

echo "Stopping Caddy..."
SCRIPT_DIR="$(cd "$(dirname "$0")"/.. && pwd)"
docker compose -f "$SCRIPT_DIR/deploy/caddy/docker-compose.yml" down 2>/dev/null

echo "✓ All services stopped"
