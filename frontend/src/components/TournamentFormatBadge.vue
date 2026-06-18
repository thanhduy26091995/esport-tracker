<template>
  <el-tag :type="tagType" size="small" effect="light" class="format-badge">
    {{ label }}
  </el-tag>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { TournamentFormat } from '@/types/tournament'

const props = defineProps<{ format: TournamentFormat }>()
const { t } = useI18n()

const label = computed(() =>
  props.format === 'round_robin_top4'
    ? t('tournaments.format.roundRobinTop4', 'Round Robin + Top 4')
    : t('tournaments.format.classic', 'Classic')
)

const tagType = computed(() =>
  props.format === 'round_robin_top4' ? 'success' : 'default'
)
</script>

<style scoped>
.format-badge {
  font-weight: 600;
  font-size: 11px;
}
</style>
