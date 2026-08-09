'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import { Logo } from './Logo';
import { LanguageSwitcher } from './LanguageSwitcher';

export function Navbar() {
  const t = useTranslations('nav');
  const [open, setOpen] = useState(false);

  const links = [
    { href: '#features', label: t('features') },
    { href: '#pricing', label: t('pricing') },
    { href: '#testimonials', label: t('testimonials') },
    { href: '#faq', label: t('faq') },
  ];

  return (
    <header className="sticky top-0 z-50 border-b border-line bg-white/80 backdrop-blur-md backdrop-saturate-150">
      <div className="mx-auto flex h-[68px] max-w-container items-center justify-between px-6">
        <a href="#top" aria-label="SaleForges home">
          <Logo variant="blue" />
        </a>

        <nav className="hidden items-center gap-8 md:flex">
          {links.map((l) => (
            <a
              key={l.href}
              href={l.href}
              className="text-[15px] font-medium text-muted transition hover:text-ink"
            >
              {l.label}
            </a>
          ))}
        </nav>

        <div className="hidden items-center gap-3 md:flex">
          <LanguageSwitcher />
          <a
            href="#"
            className="rounded-lg border border-line bg-white px-[18px] py-2.5 text-[15px] font-semibold text-ink transition hover:bg-soft"
          >
            {t('signin')}
          </a>
          <a
            href="#demo"
            className="rounded-lg bg-brand-blue px-[18px] py-2.5 text-[15px] font-semibold text-white shadow-sm transition hover:-translate-y-px hover:bg-[#1A54DB] hover:shadow-btn"
          >
            {t('demo')}
          </a>
        </div>

        <button
          className="p-2 md:hidden"
          aria-label="Menu"
          onClick={() => setOpen((v) => !v)}
        >
          <span className="mb-1 block h-0.5 w-5.5 w-[22px] rounded bg-ink" />
          <span className="mb-1 block h-0.5 w-[22px] rounded bg-ink" />
          <span className="block h-0.5 w-[22px] rounded bg-ink" />
        </button>
      </div>

      {open && (
        <div className="border-t border-line bg-white px-6 py-4 md:hidden">
          <nav className="flex flex-col gap-1">
            {links.map((l) => (
              <a
                key={l.href}
                href={l.href}
                onClick={() => setOpen(false)}
                className="rounded-lg px-2 py-2.5 text-[15px] font-medium text-muted transition hover:bg-soft hover:text-ink"
              >
                {l.label}
              </a>
            ))}
          </nav>
          <div className="mt-3 flex items-center gap-3">
            <LanguageSwitcher />
            <a
              href="#demo"
              onClick={() => setOpen(false)}
              className="flex-1 rounded-lg bg-brand-blue px-4 py-2.5 text-center text-[15px] font-semibold text-white"
            >
              {t('demo')}
            </a>
          </div>
        </div>
      )}
    </header>
  );
}
