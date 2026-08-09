# POS frontend (React + Vite)

This app uses React + TypeScript + Vite with Tailwind CSS v4 and oxlint.

- Run lint with `bunx turbo run lint --filter=pos` (oxlint)
- Type-check + build with `bunx turbo run build --filter=pos` (`tsc -b && vite build`)
- Use `@/*` for imports into `src/`
- Follow the structure and conventions of `clients/backoffice` (feature modules under `src/features/`, api client under `src/lib/`)
