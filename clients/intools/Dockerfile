FROM oven/bun:1.3 AS build
WORKDIR /app

COPY package.json bun.lock turbo.json ./
COPY backoffice/package.json ./backoffice/
COPY intools/package.json ./intools/
COPY pos/package.json ./pos/

RUN bun install --frozen-lockfile

COPY intools ./intools

ARG TURBO_TELEMETRY_DISABLED=1
RUN ./node_modules/.bin/turbo run build --filter=intools

FROM caddy:alpine AS runner
COPY --from=build /app/intools/dist /usr/share/caddy/html
COPY intools/Caddyfile /etc/caddy/Caddyfile
EXPOSE 80
