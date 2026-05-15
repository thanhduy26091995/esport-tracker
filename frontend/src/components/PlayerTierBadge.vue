<template>
  <div class="win-rate-cell">
    <!-- Win rate mode (UserTable): show chip + % or — when below min matches threshold -->
    <template v-if="showWinRate">
      <template v-if="isRanked">
        <el-tag :type="tagType" size="small" effect="light" class="tier-badge">
          {{ tierLabel }}
        </el-tag>
        <span class="win-rate-pct">{{ displayRate }}</span>
      </template>
      <span v-else class="win-rate-dash">—</span>
    </template>
    <!-- Tier-only mode (tournament/leaderboard views): always show tier chip -->
    <el-tag v-else :type="tagType" size="small" effect="light" class="tier-badge">
      {{ tierLabel }}
    </el-tag>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const props = withDefaults(defineProps<{
  tier: string
  winRate?: number
  totalMatches?: number
  minMatchesForRank?: number
}>(), {
  winRate: 0,
  totalMatches: undefined,
  minMatchesForRank: 5,
})

const { t } = useI18n()

// showWinRate is true only when totalMatches is explicitly provided (win rate mode).
const showWinRate = computed(() => props.totalMatches !== undefined)
const isRanked = computed(() => showWinRate.value && props.totalMatches! >= props.minMatchesForRank)

const tierLabel = computed(() => t(`tier.${props.tier}`, props.tier))

const tagType = computed(() => {
  switch (props.tier) {
    case 'pro':  return 'warning'
    case 'noob': return 'info'
    default:     return 'primary'  // normal → blue
  }
})

const displayRate = computed(() => `${Math.round(props.winRate * 100)}%`)
</script>

<style scoped>
.win-rate-cell {
  display: flex;
  align-items: center;
  gap: 6px;
}
.tier-badge {
  font-weight: 600;
  font-size: 11px;
}
.win-rate-pct {
  font-size: 12px;
  color: var(--text-muted, #9ca3af);
  font-variant-numeric: tabular-nums;
}
.win-rate-dash {
  color: var(--text-muted, #9ca3af);
}
</style>
