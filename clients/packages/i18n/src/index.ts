import en from './locales/en.json'
import id from './locales/id.json'

export const defaultLocale = 'en' as const
export const supportedLocales = ['en', 'id'] as const

export type Locale = (typeof supportedLocales)[number]

export type Translations = {
  common: {
    appName: string
    language: string
    switchBranch: string
    logout: string
  }
  pos: {
    tagline: string
  }
  login: {
    heroTitle1: string
    heroTitle2: string
    heroSubtitle: string
    welcomeBack: string
    subtitle: string
    username: string
    password: string
    logIn: string
    loggingIn: string
    invalidCredentials: string
    somethingWentWrong: string
    footer: string
  }
  branchSelect: {
    heroTitle1: string
    heroTitle2: string
    heroSubtitle: string
    title: string
    signedInAs: string
    active: string
    selectError: string
    loading: string
  }
  pages: {
    dashboard: string
    orders: string
    products: string
    staff: string
    users: string
    roles: string
    merchants: string
    settings: string
  }
}

export const resources: Record<Locale, { translation: Translations }> = {
  en: { translation: en as Translations },
  id: { translation: id as Translations },
}

export function isSupportedLocale(locale: string): locale is Locale {
  return (supportedLocales as readonly string[]).includes(locale)
}

export function normalizeLocale(locale: string | null | undefined): Locale {
  if (locale && isSupportedLocale(locale)) return locale
  return defaultLocale
}