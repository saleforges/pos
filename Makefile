.PHONY: dev env caddy

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
