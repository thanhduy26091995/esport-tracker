<template>
  <div class="language-switcher">
    <span class="language-label">{{ t('common.language') }}</span>
    <el-segmented
      v-model="selectedLocale"
      :options="languageOptions"
      size="small"
      class="language-segmented"
    >
      <template #default="{ item }">
        <span class="lang-option">
          <span class="lang-flag">{{ item.flag }}</span>
          <span class="lang-code">{{ item.code }}</span>
        </span>
      </template>
    </el-segmented>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useLocaleStore } from '@/stores/localeStore'
import type { LocaleCode } from '@/utils/locale'
const { t, locale } = useI18n()
const localeStore = useLocaleStore()

const languageOptions = computed(() => [
  { label: '🇻🇳 VI', value: 'vi', flag: '🇻🇳', code: 'VI' },
  { label: '🇬🇧 EN', value: 'en', flag: '🇬🇧', code: 'EN' },
])

const selectedLocale = computed<LocaleCode>({
  get: () => localeStore.locale,
  set: (nextLocale) => {
    localeStore.setLocale(nextLocale)
    locale.value = nextLocale
  },
})
</script>

<style scoped>
.language-switcher {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.language-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-muted);
  white-space: nowrap;
}

.language-segmented {
  --el-segmented-item-selected-color: #1d4ed8;
}

.lang-option {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.lang-flag {
  font-size: 14px;
  line-height: 1;
}

.lang-code {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.03em;
}
</style>
