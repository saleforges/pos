FROM oven/bun:1.3 AS build
WORKDIR /app

COPY package.json bun.lock turbo.json ./
COPY backoffice/package.json ./backoffice/
COPY pos/package.json ./pos/

RUN bun install --frozen-lockfile

COPY pos ./pos

# Override with Docker-specific config for standalone output
COPY pos/next.config.docker.ts pos/next.config.ts

RUN cd pos && bunx next build

FROM node:22-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

COPY --from=build /app/pos/public ./public

RUN mkdir -p .next
COPY --from=build /app/pos/.next/standalone ./
COPY --from=build /app/pos/.next/static ./.next/static

USER nextjs

EXPOSE 3000

CMD ["node", "server.js"]
