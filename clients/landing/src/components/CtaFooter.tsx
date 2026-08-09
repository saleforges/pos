import { useTranslations } from 'next-intl';
import { Reveal } from './Reveal';
import { Logo } from './Logo';
import { IconX, IconLinkedIn, IconInstagram } from './icons';

export function CtaBand() {
  const t = useTranslations('cta');
  return (
    <section id="demo" className="px-6 pb-[96px] pt-10">
      <div className="mx-auto max-w-container">
        <Reveal>
          <div
            className="relative overflow-hidden rounded-[24px] px-8 py-16 text-center text-white"
            style={{
              background:
                'linear-gradient(120deg, #192539 0%, #1E3A8A 55%, #1D5FF3 130%)',
            }}
          >
            <div
              className="pointer-events-none absolute inset-0"
              style={{
                background:
                  'radial-gradient(500px 240px at 85% 120%, rgba(11,180,137,.35), transparent 60%)',
              }}
            />
            <div className="relative">
              <h2 className="mb-3.5 text-[clamp(28px,4vw,40px)] font-extrabold tracking-[-0.03em]">
                {t('title')}
              </h2>
              <p className="mx-auto mb-7 max-w-[520px] text-lg text-white/80">
                {t('subtitle')}
              </p>
              <div className="flex flex-wrap justify-center gap-3.5">
                <a
                  href="#"
                  className="rounded-xl bg-white px-6 py-3.5 text-base font-semibold text-brand-blue transition hover:bg-[#F0F4FF] hover:shadow-[0_8px_28px_rgba(0,0,0,.28)]"
                >
                  {t('primary')}
                </a>
                <a
                  href="#"
                  className="rounded-xl border border-white/30 bg-white/10 px-6 py-3.5 text-base font-semibold text-white transition hover:bg-white/20"
                >
                  {t('secondary')}
                </a>
              </div>
            </div>
          </div>
        </Reveal>
      </div>
    </section>
  );
}

export function Footer() {
  const t = useTranslations('footer');
  const productLinks = t.raw('productLinks') as string[];
  const companyLinks = t.raw('companyLinks') as string[];

  return (
    <footer className="border-t border-line py-[62px_0_40px]">
      <div className="mx-auto max-w-container px-6">
        <div className="grid grid-cols-1 gap-8 md:grid-cols-2 lg:grid-cols-[1.6fr_1fr_1fr_1fr]">
          <div>
            <a href="#top" className="mb-3.5 inline-block">
              <Logo variant="blue" />
            </a>
            <p className="max-w-[300px] text-[14.5px] text-muted">{t('desc')}</p>
          </div>

          <div>
            <h4 className="mb-4 text-[13px] font-bold uppercase tracking-[0.05em] text-muted2">
              {t('product')}
            </h4>
            {productLinks.map((l) => (
              <a
                key={l}
                href="#features"
                className="mb-[11px] block text-[14.5px] text-muted transition hover:text-ink"
              >
                {l}
              </a>
            ))}
          </div>

          <div>
            <h4 className="mb-4 text-[13px] font-bold uppercase tracking-[0.05em] text-muted2">
              {t('company')}
            </h4>
            {companyLinks.map((l) => (
              <a
                key={l}
                href="#"
                className="mb-[11px] block text-[14.5px] text-muted transition hover:text-ink"
              >
                {l}
              </a>
            ))}
          </div>

          <div>
            <h4 className="mb-4 text-[13px] font-bold uppercase tracking-[0.05em] text-muted2">
              {t('contact')}
            </h4>
            <a
              href="mailto:hello@saleforges.com"
              className="mb-[11px] block text-[14.5px] text-muted transition hover:text-ink"
            >
              hello@saleforges.com
            </a>
            <a
              href="tel:+62215550100"
              className="mb-[11px] block text-[14.5px] text-muted transition hover:text-ink"
            >
              +62 21 5550 0100
            </a>
            <span className="mb-[11px] block text-[14.5px] text-muted">
              {t('location')}
            </span>
            <a
              href="#demo"
              className="block text-[14.5px] text-muted transition hover:text-ink"
            >
              {t('demo')}
            </a>
          </div>
        </div>

        <div className="mt-[46px] flex flex-wrap items-center justify-between gap-3.5 border-t border-line pt-6">
          <p className="text-[13.5px] text-muted2">{t('rights')}</p>
          <div className="flex gap-3">
            {[IconX, IconLinkedIn, IconInstagram].map((Icon, i) => (
              <a
                key={i}
                href="#"
                className="grid h-[34px] w-[34px] place-items-center rounded-[9px] border border-line text-muted transition hover:border-[#D5DDF6] hover:text-brand-blue"
              >
                <Icon className="h-4 w-4" />
              </a>
            ))}
          </div>
        </div>
      </div>
    </footer>
  );
}
