#!/bin/bash
# Start all Go services + Caddy
LOG_DIR="/tmp/pos"
SERVICES="iam merchant catalog"
SCRIPT_DIR="$(cd "$(dirname "$0")"/.. && pwd)"

mkdir -p "$LOG_DIR"
echo "Starting Go services..."

for svc in $SERVICES; do
	log="$LOG_DIR/$svc.log"
	go -C "$SCRIPT_DIR/services" run "./cmd/$svc/" > "$log" 2>&1 &
	echo "  $svc started (PID $!) → $log"
done

sleep 2
echo "Starting Caddy on https://api.saleforges.local"
docker compose -f "$SCRIPT_DIR/deploy/caddy/docker-compose.yml" up -d --remove-orphans >/dev/null 2>&1
sleep 1

echo ""
echo "✓ All services running"
echo "  https://api.saleforges.local"
echo ""
for svc in $SERVICES caddy; do echo "  Logs: make dev-logs svc=$svc"; done
echo "  Stop: make dev-down"
