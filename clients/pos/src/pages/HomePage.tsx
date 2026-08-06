import { useTranslation } from 'react-i18next'
import { LanguageSwitcher } from '@/components/ui/LanguageSwitcher'

export function HomePage() {
  const { t } = useTranslation()
  return (
    <div className="flex flex-col flex-1 items-center justify-center bg-neutral-50 font-sans">
      <main className="flex flex-1 w-full max-w-3xl flex-col items-center justify-center py-32 px-16 bg-white">
        <div className="flex flex-col items-center gap-4 text-center">
          <h1 className="text-5xl font-bold tracking-tight text-neutral-900">
            POS {t("common.appName")}
          </h1>
          <p className="max-w-md text-lg leading-8 text-neutral-600">
            {t("pos.tagline")}
          </p>
        </div>
        <div className="mt-8">
          <LanguageSwitcher />
        </div>
      </main>
    </div>
  )
}
