#!/bin/bash
# Generate deploy/caddy/.env from Go service files
OUTPUT="${1:-deploy/caddy/.env}"
echo "# Auto-generated from services/cmd/*/main.go" > "$OUTPUT"
for dir in services/cmd/*/; do
	name=$(basename "$dir" | tr '[:lower:]' '[:upper:]')
	port=$(grep -oP ':\K\d+' "${dir}main.go" 2>/dev/null | head -1)
	[ -n "$port" ] && echo "${name}_PORT=$port" >> "$OUTPUT"
done
