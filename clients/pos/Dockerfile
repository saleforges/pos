FROM oven/bun:1.3 AS build
WORKDIR /app

COPY package.json bun.lock turbo.json ./
COPY backoffice/package.json ./backoffice/
COPY pos/package.json ./pos/

RUN bun install --frozen-lockfile

COPY pos ./pos

RUN cd pos && bunx next build

FROM oven/bun:1.3 AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV HOSTNAME=0.0.0.0
ENV PORT=3000

COPY --from=build /app/pos/.next ./.next
COPY --from=build /app/pos/public ./public

# Install only production dependencies (excludes typescript, @types/*, etc.)
COPY package.json bun.lock turbo.json ./
COPY backoffice/package.json ./backoffice/
COPY pos/package.json ./pos/
RUN bun install --production --frozen-lockfile

EXPOSE 3000

CMD ["bun", "node_modules/next/dist/bin/next", "start"]
