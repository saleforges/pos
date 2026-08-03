import { defaultLocale, isSupportedLocale, normalizeLocale, type Locale } from './index'

export const LANG_COOKIE = 'sf_lang'

const MAX_AGE_SECONDS = 60 * 60 * 24 * 365

function parentDomain(hostname: string): string | null {
  const labels = hostname.split('.')
  if (labels.length < 2) return null
  if (labels.at(-1) === 'localhost') return null
  if (/^\d+(\.\d+){3}$/.test(hostname)) return null
  if (/^\d+$/.test(labels.at(-1) ?? '')) return null
  return `.${labels.slice(1).join('.')}`
}

export function getCookieDomain(): string | null {
  if (typeof window === 'undefined') return null
  return parentDomain(window.location.hostname)
}

export function readLangCookie(): Locale | null {
  if (typeof document === 'undefined') return null
  const match = document.cookie
    .split(';')
    .map((part) => part.trim())
    .find((part) => part.startsWith(`${LANG_COOKIE}=`))
  if (!match) return null
  const value = decodeURIComponent(match.slice(LANG_COOKIE.length + 1))
  return isSupportedLocale(value) ? value : null
}

export function getLangCookie(): Locale {
  return readLangCookie() ?? defaultLocale
}

export function setLangCookie(locale: Locale): void {
  if (typeof document === 'undefined') return
  const parts = [`${LANG_COOKIE}=${encodeURIComponent(locale)}`]
  const domain = getCookieDomain()
  if (domain) parts.push(`Domain=${domain}`)
  parts.push('Path=/')
  parts.push('SameSite=Lax')
  if (typeof location !== 'undefined' && location.protocol === 'https:') {
    parts.push('Secure')
  }
  parts.push(`Max-Age=${MAX_AGE_SECONDS}`)
  document.cookie = parts.join('; ')
}

export function deleteLangCookie(): void {
  if (typeof document === 'undefined') return
  const parts = [`${LANG_COOKIE}=`, 'Path=/', 'Max-Age=0']
  const domain = getCookieDomain()
  if (domain) parts.push(`Domain=${domain}`)
  document.cookie = parts.join('; ')
}

export function getBrowserLocale(): Locale {
  const fromCookie = readLangCookie()
  if (fromCookie) return fromCookie
  if (typeof navigator === 'undefined') return defaultLocale
  return normalizeLocale(navigator.language?.slice(0, 2) ?? null)
}