<template>
  <div>
    <div v-if="showFilterBar" class="filter-bar">
      <el-input v-model="searchQuery" :placeholder="t('users.searchPlaceholder')" clearable class="w-64" :prefix-icon="Search" />
      <el-select v-model="scoreFilter" :placeholder="t('users.filterByScore')" clearable class="w-44">
        <el-option :label="t('common.all')" value="" />
        <el-option :label="t('users.scorePositive')" value="positive" />
        <el-option :label="t('users.scoreInDebt')" value="negative" />
        <el-option :label="t('users.scoreZero')" value="zero" />
      </el-select>
      <span class="filter-count">{{ t('users.filterCount', { filtered: filteredUsers.length, total: users.length }) }}</span>
    </div>

    <div v-if="!loading && filteredUsers.length === 0" class="empty-state">
      <svg class="empty-state-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
      </svg>
      <p class="empty-state-title">{{ searchQuery ? t('users.noPlayersFound') : t('users.noPlayers') }}</p>
      <p class="empty-state-desc">{{ searchQuery ? t('users.tryDifferentSearch') : t('users.addFirstPlayer') }}</p>
    </div>

    <div v-else class="user-table-wrap">
      <el-table :data="filteredUsers" stripe style="width:100%" class="user-table" v-loading="loading">
      <el-table-column type="index" label="#" width="55" />
      <el-table-column prop="name" :label="t('users.colPlayer')" min-width="180">
        <template #default="{ row }">
          <div class="player-cell">
            <div class="player-avatar">{{ row.name.charAt(0).toUpperCase() }}</div>
            <span class="player-name">{{ row.name }}</span>
            <el-tag v-if="!row.is_active" type="info" size="small">{{ t('users.inactive') }}</el-tag>
          </div>
        </template>
      </el-table-column>
      <el-table-column :label="t('users.colWinRate')" width="130">
        <template #default="{ row }">
          <PlayerTierBadge
            :tier="row.tier || 'normal'"
            :win-rate="row.win_rate ?? 0"
            :total-matches="row.total_matches ?? 0"
          />
        </template>
      </el-table-column>
      <el-table-column :label="showTotalPaid ? t('users.colSettled') : t('users.colScore')" width="110" :sortable="!showTotalPaid" :prop="showTotalPaid ? undefined : 'current_score'">
        <template #default="{ row }">
          <template v-if="showTotalPaid">
            <span class="settled-badge">{{ row.total_debt_points ?? 0 }} pts</span>
          </template>
          <template v-else>
            <span class="score-pill" :class="row.current_score > 0 ? 'score-pill-positive' : row.current_score < 0 ? 'score-pill-negative' : 'score-pill-zero'">
              {{ row.current_score > 0 ? '+' : '' }}{{ row.current_score }}
            </span>
          </template>
        </template>
      </el-table-column>
      <el-table-column :label="showTotalPaid ? t('users.colTotalPaid') : t('users.colValue')" width="170">
        <template #default="{ row }">
          <template v-if="showTotalPaid && row.total_paid != null">
            <div class="total-paid-cell">
              <span class="vnd-value vnd-value--paid">{{ formatVND(row.total_paid) }}</span>
              <span v-if="row.total_debt_points > 0" class="payment-count-badge">{{ row.total_debt_points }} pts</span>
            </div>
          </template>
          <template v-else>
            <span class="vnd-value">{{ formatVND(pointsToVND(row.current_score, conversionRate)) }}</span>
          </template>
        </template>
      </el-table-column>
      <el-table-column :label="t('users.colJoined')" width="130">
        <template #default="{ row }">
          <span class="date-value">{{ formatDate(row.created_at) }}</span>
        </template>
      </el-table-column>
      <el-table-column v-if="showActions" :label="t('common.actions')" min-width="260" align="right">
        <template #default="{ row }">
          <div class="actions-cell">
            <el-tooltip v-if="!showTotalPaid && row.current_score < 0 && row.current_score <= debtThreshold" :content="t('users.triggerSettlementTooltip')" placement="top">
              <el-button size="small" type="warning" plain @click="emit('triggerSettlement', row)" :icon="Warning">
                {{ t('users.settleDebt') }}
              </el-button>
            </el-tooltip>
            <el-button size="small" text @click="emit('edit', row)" :icon="Edit">{{ t('common.edit') }}</el-button>
            <el-button size="small" text type="danger" @click="emit('delete', row)" :icon="Delete">{{ t('common.delete') }}</el-button>
          </div>
        </template>
      </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
import { Edit, Delete, Search, Warning } from '@element-plus/icons-vue'
import type { UserWithStats } from '@/types/user'
import { formatVND, pointsToVND } from '@/utils/formatters'
import { formatDate } from '@/utils/date'
import PlayerTierBadge from '@/components/PlayerTierBadge.vue'

interface Props { users: UserWithStats[]; loading?: boolean; conversionRate?: number; debtThreshold?: number; showTotalPaid?: boolean; showFilterBar?: boolean; showActions?: boolean }
const props = withDefaults(defineProps<Props>(), { loading: false, conversionRate: 22000, debtThreshold: -6, showTotalPaid: false, showFilterBar: true, showActions: true })
const emit = defineEmits<{ edit: [user: UserWithStats]; delete: [user: UserWithStats]; triggerSettlement: [user: UserWithStats] }>()

const searchQuery = ref(''); const scoreFilter = ref('')

const filteredUsers = computed(() => {
  let r = props.users
  if (searchQuery.value) { const q = searchQuery.value.toLowerCase(); r = r.filter(u => u.name.toLowerCase().includes(q)) }
  if (scoreFilter.value === 'positive') r = r.filter(u => u.current_score > 0)
  else if (scoreFilter.value === 'negative') r = r.filter(u => u.current_score < 0)
  else if (scoreFilter.value === 'zero') r = r.filter(u => u.current_score === 0)
  return r
})
</script>

<style scoped>
.player-cell { display: flex; align-items: center; gap: 10px; }

.player-avatar {
  width: 30px; height: 30px; border-radius: 50%;
  background: linear-gradient(135deg, #dbeafe, #bfdbfe);
  color: #1d4ed8; font-size: 12px; font-weight: 700;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}

.player-name { font-size: 13px; font-weight: 600; color: var(--text-primary); }

.vnd-value { font-size: 12px; color: var(--text-muted); }
.vnd-value--paid { color: var(--color-success); font-weight: 600; }
.date-value { font-size: 12px; color: var(--text-muted); }

.settled-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 10px;
  background: var(--color-warning-bg);
  border: 1px solid var(--color-warning-border);
  font-size: 11px;
  font-weight: 700;
  color: var(--color-warning);
}

.total-paid-cell {
  display: flex;
  align-items: center;
  gap: 6px;
}

.payment-count-badge {
  display: inline-flex;
  align-items: center;
  padding: 1px 6px;
  border-radius: 10px;
  background: var(--color-success-bg);
  border: 1px solid var(--color-success-border);
  font-size: 10px;
  font-weight: 700;
  color: var(--color-success);
  white-space: nowrap;
}

.user-table-wrap {
  width: 100%;
  overflow-x: auto;
}

.user-table {
  min-width: 760px;
}

.actions-cell {
  display: flex;
  justify-content: flex-end;
  gap: 4px;
  flex-wrap: wrap;
}

@media (max-width: 640px) {
  .filter-count {
    width: 100%;
    margin-left: 0;
  }

  .actions-cell {
    justify-content: flex-start;
  }

  .user-table {
    min-width: 680px;
  }
}
</style>
