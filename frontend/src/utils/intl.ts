import { i18n } from '@/plugins/i18n'
import type { LocaleCode } from '@/utils/locale'

const INTL_LOCALE_MAP: Record<LocaleCode, string> = {
  vi: 'vi-VN',
  en: 'en-US',
}

export function getCurrentLocaleCode(): LocaleCode {
  const locale = i18n.global.locale
  const current = typeof locale === 'string' ? locale : locale.value
  return current === 'en' ? 'en' : 'vi'
}

export function getIntlLocale(localeCode: LocaleCode = getCurrentLocaleCode()): string {
  return INTL_LOCALE_MAP[localeCode]
}