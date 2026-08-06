'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { resources, supportedLocales, type Locale } from '@pos/i18n'
import { setLangCookie } from '@pos/i18n/cookie'
import i18n from '@/i18n/client'

export function LanguageSwitcher({ locale }: { locale: Locale }) {
  const [current, setCurrent] = useState(locale)
  const router = useRouter()
  const label = resources[current].translation.common.language
  return (
    <select
      aria-label={label}
      value={current}
      onChange={(event) => {
        const next = event.target.value as Locale
        setLangCookie(next)
        i18n.changeLanguage(next)
        setCurrent(next)
        router.refresh()
      }}
      className="rounded-md border border-neutral-200 bg-white px-2 py-1.5 text-sm text-neutral-700 transition-colors hover:bg-neutral-50 focus:border-neutral-300 focus:outline-none"
    >
      {supportedLocales.map((code) => (
        <option key={code} value={code}>
          {code.toUpperCase()}
        </option>
      ))}
    </select>
  )
}