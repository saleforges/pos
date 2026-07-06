.PHONY: dev env caddy lgtm

ENV_FILE := deploy/caddy/.env
CADDYFILE := deploy/caddy/Caddyfile

env:
	@echo "# Auto-generated from services/*/.env" > $(ENV_FILE)
	@for dir in services/*/; do \
		name=$$(basename $$dir | tr '[:lower:]' '[:upper:]'); \
		port=$$(grep -oP '(?<=:)\d+' $$$$dir.env 2>/dev/null || echo "8080"); \
		echo "$${name}_PORT=$$port" >> $(ENV_FILE); \
	done

caddy: env
	caddy run --config $(CADDYFILE) --envfile $(ENV_FILE)

dev: env
	@echo "Starting Caddy reverse proxy on localhost:80"
	caddy run --config $(CADDYFILE) --envfile $(ENV_FILE)

lgtm:
	@docker compose -f deploy/lgtm/docker-compose.yml up -d
	@echo "LGTM stack started:"
	@echo "  Grafana : http://localhost:3000"
	@echo "  Tempo   : localhost:4317 (OTLP gRPC)"
	@echo "  Loki    : http://localhost:3100"
	@echo "  Mimir   : http://localhost:9009"
	@echo "  Prom    : http://localhost:9090"
