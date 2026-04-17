<template>
  <div>
    <div v-if="loading" class="lb-loading">
      <el-icon class="animate-spin" :size="24" style="color:var(--text-muted)"><Loading /></el-icon>
    </div>
    <div v-else-if="displayUsers.length === 0" class="lb-empty">{{ t('leaderboard.noPlayers') }}</div>
    <div v-else class="lb-list">
      <div v-for="(user, index) in displayUsers" :key="user.id" class="lb-row">
        <!-- Rank -->
        <div class="lb-rank">
          <span v-if="index === 0" class="lb-medal">🥇</span>
          <span v-else-if="index === 1" class="lb-medal">🥈</span>
          <span v-else-if="index === 2" class="lb-medal">🥉</span>
          <span v-else class="lb-rank-num">#{{ index + 1 }}</span>
        </div>

        <!-- Avatar -->
        <div class="lb-avatar" :class="`lb-avatar--${index < 3 ? ['gold','silver','bronze'][index] : 'default'}`">
          {{ user.name.charAt(0).toUpperCase() }}
        </div>

        <!-- Name -->
        <span class="lb-name">{{ user.name }}</span>
        <PlayerTierBadge :tier="user.tier || 'normal'" />

        <!-- Score -->
        <span class="score-pill" :class="user.current_score > 0 ? 'score-pill-positive' : user.current_score < 0 ? 'score-pill-negative' : 'score-pill-zero'">
          {{ user.current_score > 0 ? '+' : '' }}{{ user.current_score }}
        </span>

        <!-- VND -->
        <span v-if="showValue" class="lb-vnd">{{ formatVND(pointsToVND(user.current_score, conversionRate)) }}</span>
      </div>

      <div v-if="showDebtThreshold && debtThreshold < 0" class="lb-threshold">
        <el-icon :size="12"><Warning /></el-icon>
        {{ t('leaderboard.settlementAt', { threshold: debtThreshold }) }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Loading, Warning } from '@element-plus/icons-vue'
import type { User } from '@/types/user'
import { formatVND, pointsToVND } from '@/utils/formatters'
import PlayerTierBadge from '@/components/PlayerTierBadge.vue'

interface Props {
  users: User[]
  loading?: boolean
  limit?: number
  compact?: boolean
  showValue?: boolean
  showViewAll?: boolean
  showDebtThreshold?: boolean
  debtThreshold?: number
  conversionRate?: number
}

const props = withDefaults(defineProps<Props>(), {
  loading: false, limit: 10, compact: false, showValue: true,
  showViewAll: false, showDebtThreshold: false, debtThreshold: -6, conversionRate: 22000
})

const { t } = useI18n()

defineEmits<{ viewAll: [] }>()

const displayUsers = computed(() => {
  const sorted = [...props.users].sort((a, b) => b.current_score - a.current_score)
  return props.limit ? sorted.slice(0, props.limit) : sorted
})
</script>

<style scoped>
.lb-loading, .lb-empty {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 40px 0;
  font-size: 13px;
  color: var(--text-muted);
}

.lb-list { display: flex; flex-direction: column; gap: 2px; }

.lb-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 6px;
  border-radius: 10px;
  transition: background 0.15s;
}
.lb-row:hover { background: var(--surface-page); }

.lb-rank { width: 28px; text-align: center; flex-shrink: 0; }
.lb-medal { font-size: 18px; line-height: 1; }
.lb-rank-num { font-size: 11px; font-weight: 700; color: var(--text-muted); }

.lb-avatar {
  width: 28px; height: 28px;
  border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  font-size: 11px; font-weight: 700;
  flex-shrink: 0;
}
.lb-avatar--gold    { background: #fef9c3; color: #854d0e; }
.lb-avatar--silver  { background: #f1f5f9; color: #475569; }
.lb-avatar--bronze  { background: #fff7ed; color: #9a3412; }
.lb-avatar--default { background: #f1f5f9; color: #64748b; }

.lb-name {
  flex: 1;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.lb-vnd {
  font-size: 11px;
  color: var(--text-muted);
  flex-shrink: 0;
  min-width: 80px;
  text-align: right;
}

.lb-threshold {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--color-danger-border);
  font-size: 11px;
  color: var(--color-danger);
}
</style>
