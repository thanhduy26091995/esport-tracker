<template>
  <div>
    <div v-if="loading" class="lb-loading">
      <el-icon class="animate-spin" :size="24" style="color:var(--text-muted)"><Loading /></el-icon>
    </div>
    <div v-else-if="displayUsers.length === 0" class="lb-empty">{{ t('leaderboard.noPlayers') }}</div>
    <div v-else>
      <el-table
        :data="displayUsers"
        :row-class-name="getRowClass"
        style="width:100%"
        class="lb-table"
      >
        <!-- Rank -->
        <el-table-column width="52" label="#">
          <template #default="{ $index }">
            <span v-if="$index === 0" class="lb-medal">🥇</span>
            <span v-else-if="$index === 1" class="lb-medal">🥈</span>
            <span v-else-if="$index === 2" class="lb-medal">🥉</span>
            <span v-else class="lb-rank-badge">#{{ $index + 1 }}</span>
          </template>
        </el-table-column>

        <!-- Player -->
        <el-table-column :label="t('users.colPlayer')" min-width="160">
          <template #default="{ row, $index }">
            <div class="player-cell">
              <div
                class="lb-avatar"
                :class="`lb-avatar--${$index < 3 ? ['gold','silver','bronze'][$index] : 'default'} ${$index < 3 ? 'lb-avatar--lg' : ''}`"
              >
                {{ row.name.charAt(0).toUpperCase() }}
              </div>
              <div class="lb-name-group">
                <div class="lb-name-row">
                  <span class="lb-name">{{ row.name }}</span>
                  <PlayerTierBadge :tier="row.tier || 'normal'" />
                </div>
                <div class="lb-bar-track">
                  <div class="lb-bar-fill" :style="barStyle(row.current_score)" />
                </div>
              </div>
            </div>
          </template>
        </el-table-column>

        <!-- Score -->
        <el-table-column :label="t('users.colScore')" width="90">
          <template #default="{ row }">
            <span class="score-pill" :class="row.current_score > 0 ? 'score-pill-positive' : row.current_score < 0 ? 'score-pill-negative' : 'score-pill-zero'">
              {{ row.current_score > 0 ? '+' : '' }}{{ row.current_score }}
            </span>
          </template>
        </el-table-column>

        <!-- Value / Total Paid -->
        <el-table-column v-if="showValue" :label="showTotalPaid ? t('users.colTotalPaid') : t('users.colValue')" width="150" align="right">
          <template #default="{ row }">
            <div class="vnd-cell">
              <span class="vnd-value" :class="{ 'vnd-value--paid': showTotalPaid }">{{ getRightValue(row) }}</span>
              <span v-if="showTotalPaid && getTotalDebtPoints(row) > 0" class="debt-pts">
                {{ getTotalDebtPoints(row) }} pts
              </span>
            </div>
          </template>
        </el-table-column>
      </el-table>

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
import type { User, UserWithPaymentTotal } from '@/types/user'
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
  showTotalPaid?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  loading: false, limit: 10, compact: false, showValue: true,
  showViewAll: false, showDebtThreshold: false, debtThreshold: -6, conversionRate: 22000,
  showTotalPaid: false
})

const { t } = useI18n()

defineEmits<{ viewAll: [] }>()

const displayUsers = computed(() =>
  props.limit ? props.users.slice(0, props.limit) : props.users
)

const maxAbsScore = computed(() =>
  Math.max(...displayUsers.value.map(u => Math.abs(u.current_score)), 1)
)

function getRowClass({ rowIndex }: { rowIndex: number }): string {
  if (rowIndex === 0) return 'lb-row--gold'
  if (rowIndex === 1) return 'lb-row--silver'
  if (rowIndex === 2) return 'lb-row--bronze'
  return ''
}

function barStyle(score: number) {
  const pct = (Math.abs(score) / maxAbsScore.value) * 100
  const color = score > 0 ? '#22c55e' : score < 0 ? '#ef4444' : '#94a3b8'
  return { width: `${pct}%`, background: color }
}

function getRightValue(user: User): string {
  if (props.showTotalPaid) {
    const totalPaid = (user as UserWithPaymentTotal).total_paid
    return totalPaid != null ? formatVND(totalPaid) : '—'
  }
  return formatVND(pointsToVND(user.current_score, props.conversionRate))
}

function getTotalDebtPoints(user: User): number {
  return (user as UserWithPaymentTotal).total_debt_points ?? 0
}
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

/* Medal / rank */
.lb-medal { font-size: 18px; line-height: 1; }
.lb-rank-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border-radius: 6px;
  background: var(--surface-subtle);
  border: 1px solid var(--border-default);
  font-size: 10px;
  font-weight: 700;
  color: var(--text-muted);
}

/* Player cell */
.player-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}

.lb-avatar {
  width: 28px; height: 28px;
  border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  font-size: 11px; font-weight: 700;
  flex-shrink: 0;
}
.lb-avatar--lg    { width: 32px; height: 32px; font-size: 13px; }
.lb-avatar--gold   { background: linear-gradient(135deg, #fbbf24, #f59e0b); color: #fff; box-shadow: 0 2px 6px rgba(245,158,11,0.35); }
.lb-avatar--silver { background: linear-gradient(135deg, #cbd5e1, #94a3b8); color: #fff; box-shadow: 0 2px 6px rgba(148,163,184,0.35); }
.lb-avatar--bronze { background: linear-gradient(135deg, #fb923c, #ea580c); color: #fff; box-shadow: 0 2px 6px rgba(234,88,12,0.3); }
.lb-avatar--default { background: #f1f5f9; color: #64748b; }

.lb-name-group {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.lb-name-row {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.lb-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.lb-bar-track {
  height: 3px;
  background: var(--border-subtle);
  border-radius: 2px;
  overflow: hidden;
}

.lb-bar-fill {
  height: 100%;
  border-radius: 2px;
  transition: width 0.5s cubic-bezier(0.4, 0, 0.2, 1);
}

/* Value cell */
.vnd-cell {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
}

.vnd-value {
  font-size: 12px;
  color: var(--text-muted);
  font-variant-numeric: tabular-nums;
}

.vnd-value--paid {
  color: var(--color-success);
  font-weight: 600;
}

.debt-pts {
  font-size: 10px;
  font-weight: 500;
  color: var(--text-muted);
}

/* Row tints via deep el-table row class */
:deep(.lb-row--gold   td) { background: linear-gradient(to right, rgba(251,191,36,0.08), transparent) !important; }
:deep(.lb-row--silver td) { background: linear-gradient(to right, rgba(148,163,184,0.08), transparent) !important; }
:deep(.lb-row--bronze td) { background: linear-gradient(to right, rgba(251,146,60,0.08), transparent) !important; }
:deep(.lb-row--gold   td:hover),
:deep(.lb-row--gold:hover td)   { background: linear-gradient(to right, rgba(251,191,36,0.14), transparent) !important; }
:deep(.lb-row--silver:hover td) { background: linear-gradient(to right, rgba(148,163,184,0.14), transparent) !important; }
:deep(.lb-row--bronze:hover td) { background: linear-gradient(to right, rgba(251,146,60,0.14), transparent) !important; }

/* Table overrides */
.lb-table {
  --el-table-border-color: var(--border-subtle);
  --el-table-row-hover-bg-color: var(--surface-page);
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
