import { useTranslations } from 'next-intl';
import { Reveal } from './Reveal';
import { SectionHead } from './Features';
import { IconCheck } from './icons';

export function Pricing() {
  const t = useTranslations('pricing');

  const plans = [
    { key: 'starter', popular: false, ctaStyle: 'ghost' },
    { key: 'growth', popular: true, ctaStyle: 'primary' },
    { key: 'enterprise', popular: false, ctaStyle: 'ghost' },
  ] as const;

  return (
    <section id="pricing" className="bg-soft py-[88px]">
      <div className="mx-auto max-w-container px-6">
        <SectionHead tag={t('tag')} title={t('title')} subtitle={t('subtitle')} />
        <div className="grid grid-cols-1 items-stretch gap-5 md:grid-cols-3">
          {plans.map((p) => {
            const features = t.raw(`plans.${p.key}.features`) as string[];
            return (
              <Reveal
                key={p.key}
                className={`relative flex flex-col rounded-[18px] bg-white p-[26px_26px_30px] ${
                  p.popular
                    ? 'border border-brand-blue shadow-[0_20px_48px_-20px_rgba(29,95,243,.32)]'
                    : 'border border-line'
                } ${p.popular ? 'md:-order-none max-md:order-first' : ''}`}
              >
                {p.popular && (
                  <span className="absolute -top-3 left-1/2 -translate-x-1/2 rounded-full bg-brand-blue px-3.5 py-1.5 text-xs font-bold text-white">
                    {t('popular')}
                  </span>
                )}
                <h3 className="mb-1.5 text-lg font-bold">
                  {t(`plans.${p.key}.name` as any)}
                </h3>
                <p className="mb-[18px] min-h-[40px] text-sm text-muted">
                  {t(`plans.${p.key}.desc` as any)}
                </p>
                <div className="text-[40px] font-extrabold tracking-[-0.03em]">
                  {t(`plans.${p.key}.price` as any)}
                </div>
                <div className="mb-[22px] text-[13.5px] text-muted2">
                  {t(`plans.${p.key}.per` as any)}
                </div>
                <ul className="mb-[26px] flex flex-1 flex-col gap-[11px]">
                  {features.map((f, i) => (
                    <li key={i} className="flex items-start gap-2.5 text-[14.5px]">
                      <IconCheck className="mt-px h-[18px] w-[18px] shrink-0 text-brand-green" />
                      <span>{f}</span>
                    </li>
                  ))}
                </ul>
                <a
                  href="#demo"
                  className={`w-full rounded-[10px] px-[18px] py-2.5 text-center text-[15px] font-semibold transition ${
                    p.ctaStyle === 'primary'
                      ? 'bg-brand-blue text-white hover:-translate-y-px hover:bg-[#1A54DB] hover:shadow-btn'
                      : 'border border-line bg-white text-ink hover:bg-soft'
                  }`}
                >
                  {t(`plans.${p.key}.cta` as any)}
                </a>
              </Reveal>
            );
          })}
        </div>
      </div>
    </section>
  );
}
