# POS

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

A modern, multi-tenant Point of Sale backend built with Go.
Ready for production deployments.

## Features

- **IAM** — authentication, authorization, RBAC, JWT + JWKS
- **Merchant** — merchant & branch management, staff assignments
- **Catalog** — products, categories, variants, image upload (MinIO)
- **Multi-tenant** — isolated data per merchant with shared infrastructure
- **Observability** — OpenTelemetry tracing, Prometheus metrics, structured logging

## Quick Start

### Prerequisites

- Go 1.25+
- Docker
- dnsmasq

### 1. Setup local DNS

```bash
sudo cp deploy/dnsmasq/pos-local.conf /etc/dnsmasq.d/pos-local.conf
sudo systemctl restart dnsmasq
ping api.saleforges.local
# → 127.0.0.1
```

### 2. Start everything

```bash
make dev-up
```

Access at **`https://api.saleforges.local`** (self-signed cert — Advanced → Proceed).

### 3. Try it

```bash
curl -sk https://api.saleforges.local/v1/auth/me
# → 401 (expected, no token)
```

## Commands

| Command | Description |
|---------|-------------|
| `make dev-up` | Start all services + Caddy proxy |
| `make dev-down` | Stop all services + Caddy |
| `make dev-logs svc=iam` | Tail logs for a service |
| `make restart svc=iam` | Restart a single service |
| `make env` | Regenerate `.env` with service ports |
| `make lgtm` | Start Grafana + Loki + Tempo + Mimir |

## Services

Each service is a Go binary under `services/cmd/`:

```bash
cd services

# Run individually
go run ./cmd/iam/       # :8080
go run ./cmd/merchant/  # :8081
go run ./cmd/catalog/   # :8082

# Or all at once
make dev-up
```

## API

| Service | Endpoint |
|---------|----------|
| IAM | `https://api.saleforges.local/v1/auth/*`, `/users/*`, `/roles/*` |
| Merchant | `https://api.saleforges.local/v1/merchants/*`, `/branches/*` |
| Catalog | `https://api.saleforges.local/v1/catalog/*` |

API collections available in [`api/bruno/`](./api/bruno) (Bruno API client).

## Project Structure

```
├── services/           # Go microservices
│   ├── cmd/            # Entry points (iam, merchant, catalog)
│   └── internal/       # Domain logic (hexagonal)
├── clients/            # Frontend apps
│   ├── backoffice/     # React + Vite
│   └── pos/            # React + Vite (POS terminal)
├── deploy/             # Infrastructure
│   ├── caddy/          # Caddy reverse proxy
│   ├── dnsmasq/        # Local DNS config
│   ├── lgtm/           # Grafana + Loki + Tempo + Mimir
│   └── ansible/        # Production deployment
├── api/bruno/          # API request collections
└── scripts/            # Dev scripts (gen-env, dev-up, etc.)
```

## Tech Stack

- **Language:** Go 1.25+
- **HTTP:** Echo v4
- **Database:** PostgreSQL (pgx v5)
- **Cache:** Redis
- **Storage:** MinIO (S3-compatible)
- **Auth:** JWT + JWKS (RSA)
- **Observability:** OpenTelemetry, Prometheus, Grafana
- **Proxy:** Caddy (auto HTTPS)

## Contributing

PRs welcome. Run tests before submitting:

```bash
cd services && go test ./...
```

## License

MIT
