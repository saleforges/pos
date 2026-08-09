# SaleForges — Landing Page (Next.js + i18n)

Product-showcase landing page for **SaleForges**, built with **Next.js 14 (App Router)**, **next-intl** (Indonesian + English), and **Tailwind CSS**.

## Getting started

```bash
npm install
npm run dev
```

Open http://localhost:3000 — you'll be redirected to `/id` (default locale).
Switch languages via the globe dropdown in the navbar, or go directly to `/en`.

## Scripts

- `npm run dev` — start dev server
- `npm run build` — production build
- `npm run start` — run the production build

## Editing content

All copy lives in translation files — no need to touch components:

- `messages/id.json` — Bahasa Indonesia
- `messages/en.json` — English

To add a language: add its code to `locales` in `src/i18n/routing.ts` and create `messages/<code>.json`.

## Project structure

```
src/
├── app/[locale]/       # localized layout + page
│   ├── fonts/          # bundled Inter (no Google Fonts dependency)
│   ├── layout.tsx      # NextIntlClientProvider + metadata
│   └── page.tsx        # assembles all sections
├── components/         # Navbar, Hero, Features, Pricing,
│                       #   Testimonials, Faq, CtaFooter, Logo, etc.
├── i18n/               # routing + request config
└── middleware.ts       # locale routing
```

## Notes

- The SaleForges logo is an inline SVG React component (`src/components/Logo.tsx`)
  with color / blue / white / navy variants.
- Buttons currently link to `#` placeholders — point them to your real
  demo/signup URLs.

## Deploy

Works out of the box on **Vercel** (or any Node host):

```bash
# Vercel
vercel

# or any host
npm run build && npm run start
```
