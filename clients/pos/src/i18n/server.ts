import { createInstance, type TFunction } from 'i18next'
import { cookies } from 'next/headers'
import { defaultLocale, normalizeLocale, resources, type Locale } from '@pos/i18n'
import { LANG_COOKIE } from '@pos/i18n/cookie'

const instancesByLocale = new Map<Locale, ReturnType<typeof createInstance>>()

function getInstance(locale: Locale): ReturnType<typeof createInstance> {
  let instance = instancesByLocale.get(locale)
  if (!instance) {
    instance = createInstance()
    instance.init({
      lng: locale,
      fallbackLng: defaultLocale,
      resources,
      initImmediate: false,
      interpolation: { escapeValue: false },
    })
    instancesByLocale.set(locale, instance)
  }
  return instance
}

export async function getLocaleFromCookies(): Promise<Locale> {
  const store = await cookies()
  return normalizeLocale(store.get(LANG_COOKIE)?.value ?? null)
}

export async function getServerT(
  locale?: Locale,
): Promise<{ locale: Locale; t: TFunction }> {
  const resolved = locale ?? (await getLocaleFromCookies())
  const instance = getInstance(resolved)
  return { locale: resolved, t: instance.t.bind(instance) as TFunction }
}