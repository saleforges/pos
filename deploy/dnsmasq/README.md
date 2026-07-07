# Dnsmasq for Local Development

Resolve all `*.pos.local` domains to `127.0.0.1` without editing `/etc/hosts`.

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
ping iam.pos.local
# Should resolve to 127.0.0.1
```

> **Note:** This replaces systemd-resolved's stub resolver. For a less invasive setup,
> keep systemd-resolved and use `resolvectl dns` instead.

## Usage

Access services directly via domain (requires a reverse proxy on port 80):

```caddy
iam.pos.local {
    reverse_proxy localhost:8080
}
merchant.pos.local {
    reverse_proxy localhost:8081
}
```
