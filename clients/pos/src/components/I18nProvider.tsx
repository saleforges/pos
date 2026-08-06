'use client'

import { useState, type ReactNode } from 'react'
import { I18nextProvider } from 'react-i18next'
import type { Locale } from '@pos/i18n'
import i18n from '@/i18n/client'

export function I18nProvider({
  initialLocale,
  children,
}: {
  initialLocale: Locale
  children: ReactNode
}) {
  const [instance] = useState(() => {
    i18n.changeLanguage(initialLocale)
    return i18n
  })
  return <I18nextProvider i18n={instance}>{children}</I18nextProvider>
}