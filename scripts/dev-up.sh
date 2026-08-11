#!/bin/bash
# Start all Go services + Caddy
LOG_DIR="/tmp/pos"
SERVICES="iam merchant catalog inventory order payment"
SCRIPT_DIR="$(cd "$(dirname "$0")"/.. && pwd)"

# Optional local secrets (JWT key etc.) — gitignored
if [ -f "$SCRIPT_DIR/.env.local" ]; then
	set -a; . "$SCRIPT_DIR/.env.local"; set +a
fi

# Local dev defaults (override via environment if needed)
export DATABASE_URL="${DATABASE_URL:-postgres://pos:devpassword@localhost:5432/pos?sslmode=disable}"
export IAM_BASE_URL="${IAM_BASE_URL:-http://localhost:8080}"
export INTERNAL_API_KEY="${INTERNAL_API_KEY:-dev-internal-key}"
export INVENTORY_BASE_URL="${INVENTORY_BASE_URL:-http://localhost:8083}"
export ORDER_BASE_URL="${ORDER_BASE_URL:-http://localhost:8084}"
export IPAYMU_BASE_URL="${IPAYMU_BASE_URL:-https://sandbox.ipaymu.com/api/v2}"
export IPAYMU_VA="${IPAYMU_VA:-}"
export IPAYMU_API_KEY="${IPAYMU_API_KEY:-}"
export IPAYMU_RETURN_URL="${IPAYMU_RETURN_URL:-http://localhost:5173/payment/return}"
export IPAYMU_CANCEL_URL="${IPAYMU_CANCEL_URL:-http://localhost:5173/payment/cancel}"
export IPAYMU_NOTIFY_URL="${IPAYMU_NOTIFY_URL:-http://localhost:8085/api/v1/payments/ipaymu/callback}"
export LOG_FORMAT="${LOG_FORMAT:-text}"
export LOG_LEVEL="${LOG_LEVEL:-info}"

mkdir -p "$LOG_DIR"
echo "Starting Go services..."
echo "  DB: $DATABASE_URL"

for svc in $SERVICES; do
	log="$LOG_DIR/$svc.log"
	go -C "$SCRIPT_DIR/services" run "./cmd/$svc/" > "$log" 2>&1 &
	echo "  $svc started (PID $!) → $log"
done

sleep 3
echo ""
echo "✓ All services running"
for svc in $SERVICES; do
	port=$(grep -oP ":\K\d+" "$SCRIPT_DIR/services/cmd/$svc/main.go" 2>/dev/null | head -1)
	echo "  $svc → http://localhost:$port (log: /tmp/pos/$svc.log)"
done
echo ""
echo "  Stop: make dev-down"
