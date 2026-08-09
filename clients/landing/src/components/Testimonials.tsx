import { useTranslations } from 'next-intl';
import { Reveal } from './Reveal';
import { SectionHead } from './Features';

export function Testimonials() {
  const t = useTranslations('testimonials');

  const items = [
    { key: 'one', initials: 'RA', bg: 'bg-brand-blue' },
    { key: 'two', initials: 'DP', bg: 'bg-brand-green' },
    { key: 'three', initials: 'SW', bg: 'bg-brand-navy' },
  ] as const;

  return (
    <section id="testimonials" className="py-[88px]">
      <div className="mx-auto max-w-container px-6">
        <SectionHead tag={t('tag')} title={t('title')} subtitle={t('subtitle')} />
        <div className="grid grid-cols-1 gap-[18px] md:grid-cols-3">
          {items.map((it) => (
            <Reveal
              key={it.key}
              className="flex flex-col rounded-2xl border border-line bg-white p-6"
            >
              <div className="mb-3.5 tracking-[2px] text-[#F5A623]">★★★★★</div>
              <p className="flex-1 text-[15.5px] text-ink">
                &ldquo;{t(`items.${it.key}.quote` as any)}&rdquo;
              </p>
              <div className="mt-5 flex items-center gap-3">
                <span
                  className={`grid h-10 w-10 place-items-center rounded-full text-[15px] font-bold text-white ${it.bg}`}
                >
                  {it.initials}
                </span>
                <div>
                  <b className="block text-[14.5px] font-bold">
                    {t(`items.${it.key}.name` as any)}
                  </b>
                  <span className="text-[13px] text-muted">
                    {t(`items.${it.key}.role` as any)}
                  </span>
                </div>
              </div>
            </Reveal>
          ))}
        </div>
      </div>
    </section>
  );
}
