#!/bin/bash
SVC=$1
LOG_DIR="/tmp/pos"
SCRIPT_DIR="$(cd "$(dirname "$0")"/.. && pwd)"

if [ -z "$SVC" ]; then
	echo "Usage: make restart svc={iam|merchant|catalog}"
	exit 1
fi

echo "Restarting $SVC..."
pkill -f "go.*cmd/$SVC" 2>/dev/null
sleep 1
go -C "$SCRIPT_DIR/services" run "./cmd/$SVC/" > "$LOG_DIR/$SVC.log" 2>&1 &
echo "  $SVC restarted (PID $!)"
