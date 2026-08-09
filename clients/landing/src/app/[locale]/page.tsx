import { setRequestLocale } from 'next-intl/server';
import { useTranslations } from 'next-intl';
import { Navbar } from '@/components/Navbar';
import { Hero } from '@/components/Hero';
import { Features } from '@/components/Features';
import { Pricing } from '@/components/Pricing';
import { Testimonials } from '@/components/Testimonials';
import { Faq } from '@/components/Faq';
import { CtaBand, Footer } from '@/components/CtaFooter';

function TrustStrip() {
  const t = useTranslations('trust');
  const logos = ['Kopi Nusa', 'RetailX', 'Salon Aura', 'WarungPintar', 'Bistro 88', 'MartaShop'];
  return (
    <div className="mx-auto max-w-container px-6 pb-2 pt-[76px]">
      <p className="mb-[22px] text-center text-[13px] font-semibold uppercase tracking-[0.06em] text-muted2">
        {t('label')}
      </p>
      <div className="flex flex-wrap items-center justify-center gap-x-14 gap-y-10 opacity-70">
        {logos.map((l) => (
          <span key={l} className="text-[19px] font-bold tracking-tight text-muted">
            {l}
          </span>
        ))}
      </div>
    </div>
  );
}

export default async function Home({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);

  return (
    <>
      <a id="top" />
      <Navbar />
      <main>
        <Hero />
        <TrustStrip />
        <Features />
        <Pricing />
        <Testimonials />
        <Faq />
        <CtaBand />
      </main>
      <Footer />
    </>
  );
}
