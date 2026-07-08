.PHONY: dev-up dev-down dev-logs restart env lgtm

env:
	@scripts/gen-env.sh

dev-up: env
	@scripts/dev-up.sh

dev-down:
	@scripts/dev-down.sh

dev-logs:
	@scripts/dev-logs.sh $(svc)

restart:
	@scripts/dev-restart.sh $(svc)

lgtm:
	@docker compose -f deploy/lgtm/docker-compose.yml up -d
	@echo "LGTM stack started:"
	@echo "  Grafana : http://localhost:3000"
	@echo "  Tempo   : localhost:4317 (OTLP gRPC)"
	@echo "  Loki    : http://localhost:3100"
	@echo "  Mimir   : http://localhost:9009"
	@echo "  Prom    : http://localhost:9090"
