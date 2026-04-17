import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { LocaleCode } from '@/utils/locale'
import { resolveInitialLocale, setStoredLocale } from '@/utils/locale'

export const useLocaleStore = defineStore('locale', () => {
  const locale = ref<LocaleCode>(resolveInitialLocale())

  function setLocale(nextLocale: LocaleCode) {
    locale.value = nextLocale
    setStoredLocale(nextLocale)
  }

  function getLocale(): LocaleCode {
    return locale.value
  }

  return {
    locale,
    setLocale,
    getLocale,
  }
})
