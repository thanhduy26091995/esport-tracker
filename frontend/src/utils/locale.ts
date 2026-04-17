export const LOCALE_STORAGE_KEY = 'app.locale'

export const SUPPORTED_LOCALES = ['vi', 'en'] as const

export type LocaleCode = (typeof SUPPORTED_LOCALES)[number]

export function isLocaleCode(value: string): value is LocaleCode {
  return (SUPPORTED_LOCALES as readonly string[]).includes(value)
}

export function getStoredLocale(): LocaleCode | null {
  if (typeof window === 'undefined') return null

  const rawLocale = window.localStorage.getItem(LOCALE_STORAGE_KEY)
  if (!rawLocale) return null

  return isLocaleCode(rawLocale) ? rawLocale : null
}

export function setStoredLocale(locale: LocaleCode): void {
  if (typeof window === 'undefined') return
  window.localStorage.setItem(LOCALE_STORAGE_KEY, locale)
}

export function resolveInitialLocale(defaultLocale: LocaleCode = 'vi'): LocaleCode {
  const storedLocale = getStoredLocale()
  return storedLocale ?? defaultLocale
}
