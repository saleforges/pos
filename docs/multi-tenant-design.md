# Multi-Tenant POS System Design

## Overview

POS SaaS with multi-tenant architecture: IAM (auth/roles/users), Merchant, Catalog, and
Inventory services deployed as microservices or combined binary.

## Tenant Model

- **Platform level** (superadmin): manages all merchants, users, roles
- **Merchant level**: owner/admin/cashier/viewer scoped to a single merchant

## User Types

- `platform`: superadmin — belongs to platform, not any merchant
- `merchant`: regular user — must be associated with at least one merchant via staff

## Auth Flow

1. Login → `POST /auth/login` → `{ access_token, refresh_token, expires_in }`
2. Get user + merchants → `GET /auth/me` → `{ id, username, ..., merchants: [...] }`
3. FE picks merchant → all subsequent requests include `merchant_id` in URL path
4. No select-merchant endpoint (merchant_id is URL-based)

## API Response Format

All responses wrapped in `{ message, data }`:
- Success: `{ message: "success", data: { ... } }`
- Error: `{ message: "<error>" }` (no data field)

## Roles

- Default roles: `superadmin`, `owner`, `admin`, `supervisor`, `cashier`, `viewer`
- `superadmin`: all permissions, platform-level only
- `owner`: full merchant access
- `viewer`: read-only
- Custom roles per merchant supported

## Services

- **IAM**: auth, users, roles, permissions — port 8080
- **Merchant**: merchants, branches, staff — port 8081
- **Catalog**: categories, products, variants — port 8082
- **Inventory** (planned): stock, warehouse, adjustment

## Infrastructure

- Docker + Ansible deployment on single VM
- PostgreSQL (shared `pos` database for IAM + Merchant, app-level FK validation)
- Redis (optional, for IAM cache)
- MinIO (image storage, bucket: catalog-dev)
- LGTM stack: Loki (logs), Grafana (dashboards), Tempo (traces), Mimir (metrics)
- Caddy reverse proxy with auto-HTTPS

## Domains (Dev)

- `grafana.saleforges.com` → Grafana
- `api-dev.saleforges.com` → IAM + Merchant + Catalog
- `minio.saleforges.com` → MinIO API
- `minio-console.saleforges.com` → MinIO Console

## Tracing

- OTEL gRPC to Tempo at `tempo:4317`
- Each service uses `otelecho.Middleware("service-name")`
- Custom span events for cache hit/miss

## Image Storage

- Bucket: `catalog-dev` (public-read)
- DB stores relative path: `products/abc123.jpg`
- Full URL: `https://minio.saleforges.com/catalog-dev/products/abc123.jpg`
- Upload endpoint: `POST /api/v1/merchants/:merchantID/images`
