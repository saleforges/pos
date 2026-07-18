FROM oven/bun:1.3 AS base

# 1. Prune the monorepo to just what 'pos' needs
FROM base AS pruner
WORKDIR /app
RUN bun add -g turbo
COPY . .
RUN turbo prune pos --docker

# 2. Install deps + build
FROM base AS installer
WORKDIR /app
COPY --from=pruner /app/out/json/ .
RUN bun install --frozen-lockfile
COPY --from=pruner /app/out/full/ .
RUN bunx turbo run build --filter=pos

# 3. Minimal runtime image
FROM base AS runner
WORKDIR /app
ENV NODE_ENV=production
RUN addgroup --system --gid 1001 nodejs && adduser --system --uid 1001 nextjs

COPY --from=installer /app/pos/public ./pos/public
COPY --from=installer --chown=nextjs:nodejs /app/pos/.next/standalone ./
COPY --from=installer --chown=nextjs:nodejs /app/pos/.next/static ./pos/.next/static

USER nextjs
EXPOSE 3000
ENV PORT=3000
CMD ["bun", "pos/server.js"]