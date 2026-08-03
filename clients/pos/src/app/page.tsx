import { getLocaleFromCookies, getServerT } from "@/i18n/server";
import { LanguageSwitcher } from "@/components/LanguageSwitcher";

export default async function Home() {
  const locale = await getLocaleFromCookies();
  const { t } = await getServerT(locale);
  return (
    <div className="flex flex-col flex-1 items-center justify-center bg-zinc-50 font-sans dark:bg-black">
      <main className="flex flex-1 w-full max-w-3xl flex-col items-center justify-center py-32 px-16 bg-white dark:bg-black">
        <div className="flex flex-col items-center gap-4 text-center">
          <h1 className="text-5xl font-bold tracking-tight text-black dark:text-zinc-50">
            POS {t("common.appName")}
          </h1>
          <p className="max-w-md text-lg leading-8 text-zinc-600 dark:text-zinc-400">
            {t("pos.tagline")}
          </p>
        </div>
        <div className="mt-8">
          <LanguageSwitcher locale={locale} />
        </div>
      </main>
    </div>
  );
}