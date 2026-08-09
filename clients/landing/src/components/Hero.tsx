import { useTranslations } from 'next-intl';
import { Reveal } from './Reveal';

export function Hero() {
  const t = useTranslations('hero');
  const d = useTranslations('dashboard');

  const navItems = [
    { key: 'dashboard', active: true },
    { key: 'pos' },
    { key: 'inventory' },
    { key: 'payments' },
    { key: 'customers' },
    { key: 'analytics' },
  ];

  const bars = [
    { h: 52 },
    { h: 68 },
    { h: 45 },
    { h: 82, g: true },
    { h: 60 },
    { h: 74 },
    { h: 95, g: true },
  ];

  return (
    <section className="relative overflow-hidden pb-[72px] pt-24">
      {/* gradient glows */}
      <div
        className="pointer-events-none absolute inset-0"
        style={{
          background:
            'radial-gradient(560px 340px at 82% -8%, rgba(29,95,243,.10), transparent 60%), radial-gradient(520px 340px at 12% 8%, rgba(11,180,137,.08), transparent 62%)',
        }}
      />
      <div className="relative mx-auto max-w-container px-6 text-center">
        <span className="mb-6 inline-flex items-center gap-2 rounded-full border border-line bg-white px-3.5 py-1.5 text-[13px] font-semibold text-muted">
          <span className="h-[7px] w-[7px] rounded-full bg-brand-green shadow-[0_0_0_3px_rgba(11,180,137,.18)]" />
          {t('badge')}
        </span>

        <h1 className="mx-auto max-w-3xl text-[clamp(38px,6vw,66px)] font-extrabold leading-[1.04] tracking-[-0.035em] text-ink">
          {t('titleA')}{' '}
          <span className="bg-gradient-to-r from-brand-blue to-brand-green bg-clip-text text-transparent">
            {t('titleHighlight')}
          </span>
        </h1>

        <p className="mx-auto mt-5 max-w-[620px] text-[clamp(17px,2.2vw,20px)] text-muted">
          {t('subtitle')}
        </p>

        <div className="mt-8 flex flex-wrap justify-center gap-3.5">
          <a
            href="#demo"
            className="rounded-xl bg-brand-blue px-6 py-3.5 text-base font-semibold text-white shadow-sm transition hover:-translate-y-px hover:bg-[#1A54DB] hover:shadow-btn"
          >
            {t('ctaPrimary')}
          </a>
          <a
            href="#features"
            className="rounded-xl border border-line bg-white px-6 py-3.5 text-base font-semibold text-ink transition hover:bg-soft"
          >
            {t('ctaSecondary')}
          </a>
        </div>

        <p className="mt-5 text-[13.5px] text-muted2">{t('note')}</p>

        {/* Product showcase */}
        <Reveal className="mt-14">
          <div className="mx-auto max-w-[1000px] overflow-hidden rounded-[20px] border border-line bg-white shadow-frame">
            <div className="flex items-center gap-2 border-b border-line2 bg-soft px-[18px] py-3.5">
              <i className="inline-block h-[11px] w-[11px] rounded-full bg-[#FF6058]" />
              <i className="inline-block h-[11px] w-[11px] rounded-full bg-[#FFBE2E]" />
              <i className="inline-block h-[11px] w-[11px] rounded-full bg-[#2ACB42]" />
              <span className="ml-3 text-[12.5px] font-medium text-muted2">
                {d('url')}
              </span>
            </div>

            <div className="grid grid-cols-1 gap-[18px] p-[22px] md:grid-cols-[200px_1fr]">
              <aside className="flex flex-row gap-1.5 overflow-x-auto md:flex-col">
                {navItems.map((n) => (
                  <div
                    key={n.key}
                    className={`flex shrink-0 items-center gap-2.5 rounded-[9px] px-2.5 py-2.5 text-[13.5px] font-medium ${
                      n.active
                        ? 'bg-brand-blue/10 font-semibold text-brand-blue'
                        : 'text-muted'
                    }`}
                  >
                    <span className="h-4 w-4 rounded-[5px] bg-current opacity-85" />
                    {d(`nav.${n.key}` as any)}
                  </div>
                ))}
              </aside>

              <div className="flex flex-col gap-4">
                <div className="grid grid-cols-1 gap-3 md:grid-cols-3">
                  <div className="rounded-xl border border-line px-4 py-3.5">
                    <div className="text-xs font-semibold uppercase tracking-wide text-muted2">
                      {d('kpi.revenue')}
                    </div>
                    <div className="mt-1 text-2xl font-extrabold tracking-tight">
                      Rp 24,8jt
                    </div>
                    <div className="mt-0.5 text-[12.5px] font-semibold text-brand-green">
                      {d('kpi.revenueChange')}
                    </div>
                  </div>
                  <div className="rounded-xl border border-line px-4 py-3.5">
                    <div className="text-xs font-semibold uppercase tracking-wide text-muted2">
                      {d('kpi.transactions')}
                    </div>
                    <div className="mt-1 text-2xl font-extrabold tracking-tight">
                      1,284
                    </div>
                    <div className="mt-0.5 text-[12.5px] font-semibold text-brand-green">
                      {d('kpi.transactionsChange')}
                    </div>
                  </div>
                  <div className="rounded-xl border border-line px-4 py-3.5">
                    <div className="text-xs font-semibold uppercase tracking-wide text-muted2">
                      {d('kpi.stock')}
                    </div>
                    <div className="mt-1 text-2xl font-extrabold tracking-tight">7</div>
                    <div className="mt-0.5 text-[12.5px] font-semibold text-muted2">
                      {d('kpi.stockNote')}
                    </div>
                  </div>
                </div>

                <div className="rounded-xl border border-line p-4">
                  <div className="mb-3.5 flex items-baseline justify-between">
                    <b className="text-sm">{d('chartTitle')}</b>
                    <span className="text-xs text-muted2">{d('chartLive')}</span>
                  </div>
                  <div className="flex h-[120px] items-end gap-2.5">
                    {bars.map((b, i) => (
                      <div
                        key={i}
                        className={`flex-1 rounded-t-md ${
                          b.g
                            ? 'bg-gradient-to-b from-brand-green to-brand-green/35'
                            : 'bg-gradient-to-b from-brand-blue to-brand-blue/35'
                        }`}
                        style={{ height: `${b.h}%` }}
                      />
                    ))}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </Reveal>
      </div>
    </section>
  );
}
