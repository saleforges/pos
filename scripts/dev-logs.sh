#!/bin/bash
# Tail logs for a specific service
SVC="${1:-}"
LOG_DIR="/tmp/pos"

if [ -z "$SVC" ]; then
	echo "Usage: make dev-logs svc={iam|merchant|catalog|caddy}"
	exit 1
fi

if [ "$SVC" = "caddy" ]; then
	docker logs pos-caddy -n 10 2>&1 | sed 's/^/[caddy] /'
	echo "--- tailing ---"
	docker logs pos-caddy -f 2>&1 | sed 's/^/[caddy] /'
elif [ -f "$LOG_DIR/$SVC.log" ]; then
	tail -f "$LOG_DIR/$SVC.log" | sed "s/^/[$SVC] /"
else
	echo "Service '$SVC' not found. Use: iam, merchant, catalog, or caddy"
	exit 1
fi
