import { createI18n } from 'vue-i18n'
import en from '@/locales/en.json'
import vi from '@/locales/vi.json'
import { resolveInitialLocale } from '@/utils/locale'

const messages = {
  vi,
  en,
}

export const i18n = createI18n({
  legacy: false,
  globalInjection: true,
  locale: resolveInitialLocale(),
  fallbackLocale: 'vi',
  messages,
  missingWarn: import.meta.env.DEV,
  fallbackWarn: import.meta.env.DEV,
  missing: (locale, key) => {
    if (import.meta.env.DEV) {
      console.warn(`[i18n][missing] locale=${locale} key=${key}`)
    }
    return key
  },
})
