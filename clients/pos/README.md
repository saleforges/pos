# POS (React + Vite)

The POS terminal frontend, built with React + TypeScript + Vite. Currently a minimal landing page; the POS terminal UI will be developed on top of this stack.

## Stack

- React 19 + Vite 8 (Oxc-based)
- Tailwind CSS v4 (`@tailwindcss/vite`)
- react-router-dom v7 (available for upcoming pages)
- i18next + react-i18next with the shared `@pos/i18n` package (client-only, cookie-driven locale)
- oxlint

## Getting Started

From the `clients` workspace root:

```bash
bun install
bunx turbo run dev --filter=pos
```

Open [http://localhost:5173](http://localhost:5173).

The dev server proxies `/api/*` to `VITE_API_PROXY_TARGET` (see `.env.development`).

## Scripts

| Command | Description |
|---------|-------------|
| `bunx turbo run dev --filter=pos` | Start Vite dev server |
| `bunx turbo run lint --filter=pos` | Lint with oxlint |
| `bunx turbo run build --filter=pos` | Type-check + build to `dist/` |
| `bunx turbo run preview --filter=pos` | Preview the production build |

## Structure

```
src/
  main.tsx               React entry
  App.tsx                App composition (i18n provider)
  index.css              Tailwind + design tokens (@theme)
  i18n/index.ts          Client-only i18n init
  components/ui/         Reusable UI primitives
  pages/                 Route pages
  features/              (reserved) feature modules
  lib/                   (reserved) shared helpers, api client
```

## Deployment

Built with the repo `Dockerfile` → static `dist/` served by Caddy (SPA fallback + `/api/*` proxy). CI/CD in `.github/workflows/pos.yml`.
