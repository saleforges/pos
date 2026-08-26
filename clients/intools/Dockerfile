FROM oven/bun:1.3 AS base

# 1. Prune the monorepo to just what 'intools' needs
FROM base AS pruner
WORKDIR /app
RUN bun add -g turbo
COPY . .
RUN turbo prune intools --docker

# 2. Install deps + build
FROM base AS installer
WORKDIR /app
COPY --from=pruner /app/out/json/ .
RUN bun install --frozen-lockfile
COPY --from=pruner /app/out/full/ .
ARG TURBO_TELEMETRY_DISABLED=1
RUN ./node_modules/.bin/turbo run build --filter=intools

# 3. Minimal runtime image
FROM caddy:alpine AS runner
COPY --from=installer /app/intools/dist /usr/share/caddy/html
COPY intools/Caddyfile /etc/caddy/Caddyfile
EXPOSE 80
