FROM oven/bun:1.3 AS build
WORKDIR /app

# Copy workspace root files first (better layer caching)
COPY package.json bun.lock turbo.json ./
COPY backoffice/package.json ./backoffice/
COPY pos/package.json ./pos/

RUN bun install --frozen-lockfile

# Now copy the rest of the source
COPY backoffice ./backoffice

ARG TURBO_TELEMETRY_DISABLED=1
RUN ./node_modules/.bin/turbo run build --filter=backoffice

FROM caddy:alpine AS runner
COPY --from=build /app/backoffice/dist /usr/share/caddy/html
COPY backoffice/Caddyfile /etc/caddy/Caddyfile
EXPOSE 80
