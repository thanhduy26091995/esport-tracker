<template>
  <div class="wc-filter-bar">
    <button
      v-for="option in options"
      :key="option.value"
      class="wc-filter-pill"
      :class="{ 'wc-filter-pill--active': modelValue === option.value }"
      @click="$emit('update:modelValue', option.value)"
    >
      {{ option.label }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

defineProps<{ modelValue: string }>()
defineEmits<{ (e: 'update:modelValue', v: string): void }>()

const options = computed(() => [
  { value: '', label: t('wc.allMatches') },
  { value: 'group_A', label: 'A' },
  { value: 'group_B', label: 'B' },
  { value: 'group_C', label: 'C' },
  { value: 'group_D', label: 'D' },
  { value: 'group_E', label: 'E' },
  { value: 'group_F', label: 'F' },
  { value: 'group_G', label: 'G' },
  { value: 'group_H', label: 'H' },
  { value: 'group_I', label: 'I' },
  { value: 'group_J', label: 'J' },
  { value: 'group_K', label: 'K' },
  { value: 'group_L', label: 'L' },
  { value: 'r32', label: t('wc.stageR32') },
  { value: 'r16', label: t('wc.stageR16') },
  { value: 'qf', label: t('wc.stageQF') },
  { value: 'sf', label: t('wc.stageSF') },
  { value: 'third_place', label: t('wc.stageThirdPlace') },
  { value: 'final', label: t('wc.stageFinal') },
])
</script>

<style scoped>
.wc-filter-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding: 4px 0 12px;
}

.wc-filter-pill {
  padding: 5px 12px;
  border-radius: 20px;
  border: 1px solid var(--border-default);
  background: var(--surface-card);
  color: var(--text-secondary);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
  white-space: nowrap;
  font-family: 'Inter', system-ui, sans-serif;
}

.wc-filter-pill:hover {
  border-color: #16a34a;
  color: #16a34a;
}

.wc-filter-pill--active {
  background: #16a34a;
  border-color: #16a34a;
  color: #ffffff;
  box-shadow: 0 2px 8px rgba(22, 163, 74, 0.3);
}
</style>
