import { useTranslations } from 'next-intl';
import { Reveal } from './Reveal';
import {
  IconPOS,
  IconBox,
  IconCard,
  IconUsers,
  IconFlow,
  IconChart,
} from './icons';

export function SectionHead({
  tag,
  title,
  subtitle,
}: {
  tag: string;
  title: string;
  subtitle?: string;
}) {
  return (
    <div className="mx-auto mb-[54px] max-w-[660px] text-center">
      <span className="mb-3.5 inline-block text-[13px] font-bold uppercase tracking-[0.06em] text-brand-blue">
        {tag}
      </span>
      <h2 className="mb-3.5 text-[clamp(28px,4vw,42px)] font-extrabold leading-[1.1] tracking-[-0.03em]">
        {title}
      </h2>
      {subtitle && <p className="text-[17.5px] text-muted">{subtitle}</p>}
    </div>
  );
}

export function Features() {
  const t = useTranslations('features');

  const items = [
    { key: 'pos', Icon: IconPOS, tone: 'blue' },
    { key: 'inventory', Icon: IconBox, tone: 'green' },
    { key: 'payment', Icon: IconCard, tone: 'navy' },
    { key: 'customers', Icon: IconUsers, tone: 'blue' },
    { key: 'automation', Icon: IconFlow, tone: 'green' },
    { key: 'analytics', Icon: IconChart, tone: 'navy' },
  ] as const;

  const tone: Record<string, string> = {
    blue: 'bg-brand-blue/10 text-brand-blue',
    green: 'bg-brand-green/10 text-brand-green',
    navy: 'bg-brand-navy/[.08] text-brand-navy',
  };

  return (
    <section id="features" className="py-[88px]">
      <div className="mx-auto max-w-container px-6">
        <SectionHead tag={t('tag')} title={t('title')} subtitle={t('subtitle')} />
        <div className="grid grid-cols-1 gap-[18px] md:grid-cols-3">
          {items.map(({ key, Icon, tone: tn }) => (
            <Reveal
              key={key}
              className="group rounded-2xl border border-line bg-white p-6 transition hover:-translate-y-0.5 hover:border-[#D5DDF6] hover:shadow-card"
            >
              <div
                className={`mb-4 grid h-11 w-11 place-items-center rounded-[11px] ${tone[tn]}`}
              >
                <Icon className="h-[22px] w-[22px]" />
              </div>
              <h3 className="mb-2 text-lg font-bold tracking-tight">
                {t(`items.${key}.title` as any)}
              </h3>
              <p className="text-[15px] text-muted">
                {t(`items.${key}.desc` as any)}
              </p>
            </Reveal>
          ))}
        </div>
      </div>
    </section>
  );
}
