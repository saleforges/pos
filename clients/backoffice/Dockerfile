FROM oven/bun:1.3 AS build
WORKDIR /app

# Copy workspace root files first (better layer caching)
COPY package.json bun.lock turbo.json ./
COPY backoffice/package.json ./backoffice/
# COPY packages ./packages (will used later after adding it)

RUN bun install --frozen-lockfile

# Now copy the rest of the source
COPY backoffice ./backoffice

RUN bunx turbo run build --filter=backoffice

FROM nginx:alpine AS runner
COPY --from=build /app/backoffice/dist /usr/share/nginx/html
COPY backoffice/nginx.conf.template /etc/nginx/templates/default.conf.template
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]