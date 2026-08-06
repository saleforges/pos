import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import { defaultLocale, normalizeLocale, resources } from '@pos/i18n'
import { getBrowserLocale, setLangCookie } from '@pos/i18n/cookie'

i18n.use(initReactI18next).init({
  lng: getBrowserLocale(),
  fallbackLng: defaultLocale,
  resources,
  initImmediate: false,
  interpolation: { escapeValue: false },
})

i18n.on('languageChanged', (lng) => {
  setLangCookie(normalizeLocale(lng))
})

export default i18n
