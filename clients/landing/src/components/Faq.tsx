import { useTranslations } from 'next-intl';
import { SectionHead } from './Features';

export function Faq() {
  const t = useTranslations('faq');
  const keys = ['offline', 'setup', 'multi', 'payment', 'security', 'switch'];

  return (
    <section id="faq" className="bg-soft py-[88px]">
      <div className="mx-auto max-w-container px-6">
        <SectionHead tag={t('tag')} title={t('title')} />
        <div className="mx-auto max-w-[760px]">
          {keys.map((k, i) => (
            <details
              key={k}
              className="group border-b border-line py-1"
              open={i === 0}
            >
              <summary className="flex cursor-pointer list-none items-center justify-between gap-5 py-5 text-[17px] font-semibold text-ink [&::-webkit-details-marker]:hidden">
                {t(`items.${k}.q` as any)}
                <span className="relative h-5 w-5 shrink-0">
                  <span className="absolute left-0.5 top-[9px] h-0.5 w-4 rounded bg-muted" />
                  <span className="absolute left-[9px] top-0.5 h-4 w-0.5 rounded bg-muted transition group-open:rotate-90 group-open:opacity-0" />
                </span>
              </summary>
              <div className="max-w-[660px] pb-[22px] text-[15.5px] text-muted">
                {t(`items.${k}.a` as any)}
              </div>
            </details>
          ))}
        </div>
      </div>
    </section>
  );
}
