import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import { defaultLocale, resources } from '@pos/i18n'

i18n.use(initReactI18next).init({
  lng: defaultLocale,
  fallbackLng: defaultLocale,
  resources,
  initImmediate: false,
  interpolation: { escapeValue: false },
})

export default i18n