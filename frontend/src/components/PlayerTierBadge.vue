<template>
  <el-tag :type="tagType" size="small" effect="light" class="tier-badge">
    {{ tierLabel }}
  </el-tag>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ tier: string }>()

const tierLabel = computed(() => {
  switch (props.tier) {
    case 'pro': return 'Pro'
    case 'noop': return 'Noob'
    default: return 'Normal'
  }
})

const tagType = computed(() => {
  switch (props.tier) {
    case 'pro': return 'warning'
    case 'noop': return 'info'
    default: return 'success'  // 'normal' = green
  }
})
</script>

<style scoped>
.tier-badge {
  font-weight: 600;
  font-size: 11px;
}
</style>
