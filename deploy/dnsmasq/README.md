# Dnsmasq for Local Development

Resolve all `*.saleforges.local` domains to `127.0.0.1` without editing `/etc/hosts`.

## Prerequisites

```bash
sudo apt-get install -y dnsmasq
```

## Setup

```bash
# 1. Copy config
sudo cp deploy/dnsmasq/pos-local.conf /etc/dnsmasq.d/pos-local.conf

# 2. Replace systemd-resolved stub with dnsmasq
sudo rm -f /etc/resolv.conf
echo "nameserver 127.0.0.1" | sudo tee /etc/resolv.conf
sudo sed -i 's/^hosts:.*/hosts:          files dns/' /etc/nsswitch.conf
sudo systemctl restart dnsmasq

# 3. Test
ping api.saleforges.local
# Should resolve to 127.0.0.1
```

> **Note:** This replaces systemd-resolved's stub resolver. For a less invasive setup,
> keep systemd-resolved and use `resolvectl dns` instead.

## Usage

### Start everything

```bash
make dev-up
```

This will:
1. Generate `deploy/caddy/.env` with service ports
2. Start IAM (`:8080`), Merchant (`:8081`), Catalog (`:8082`)
3. Start Caddy (via Docker) on `https://api.saleforges.local`

### Stop everything

```bash
make dev-down
```

### View logs

```bash
make dev-logs svc=iam         # tail IAM logs
make dev-logs svc=merchant    # tail Merchant logs
make dev-logs svc=catalog     # tail Catalog logs
make dev-logs svc=caddy       # tail Caddy (Docker) logs
```

### Restart a single service

```bash
make restart svc=iam         # restart IAM only
make restart svc=merchant    # restart Merchant only
make restart svc=catalog     # restart Catalog only
```

Useful when you change code and need a quick reload without stopping everything.

### Generate env only

```bash
make env
```

Regenerates `deploy/caddy/.env` from service source files.

## Access

| URL | Backend |
|-----|---------|
| `https://api.saleforges.local/v1/auth/*` | IAM (`:8080`) |
| `https://api.saleforges.local/v1/merchants/*` | Merchant (`:8081`) |
| `https://api.saleforges.local/v1/catalog/*` | Catalog (`:8082`) |

HTTP (`http://api.saleforges.local`) automatically redirects to HTTPS.
Self-signed certificate (Caddy local CA) — click Advanced → Proceed in browser.

## How it works

```
Browser → api.saleforges.local:443 → Caddy (Docker, network=host)
  ├── /v1/auth/* → rewrite → /api/v1/auth/* → localhost:8080 (IAM)
  ├── /v1/merchants/* → rewrite → /api/v1/merchants/* → localhost:8081 (Merchant)
  └── /v1/catalog/* → rewrite → /api/v1/catalog/* → localhost:8082 (Catalog)
```

Caddy runs in Docker with `network_mode: host`, so it can reach Go services
running directly on your machine. No Caddy installation needed on the host.
