<template>
  <div>
    <div v-if="loading" class="fc-loading">
      <el-icon class="animate-spin" :size="24" style="color:var(--text-muted)"><Loading /></el-icon>
    </div>
    <div v-else-if="contributors.length === 0" class="fc-empty">
      <el-icon :size="36" class="fc-empty-icon"><Medal /></el-icon>
      <p>{{ t('dashboard.noContributors') }}</p>
    </div>
    <div v-else class="fc-list">
      <div class="fc-header">
        <span class="fc-col-rank">{{ t('dashboard.colRank') }}</span>
        <span class="fc-col-name">{{ t('dashboard.colName') }}</span>
        <span class="fc-col-points">{{ t('dashboard.colPointsContributed') }}</span>
        <span class="fc-col-money">{{ t('dashboard.colMoneyContributed') }}</span>
      </div>
      <div v-for="item in contributors" :key="item.user_id" class="fc-row">
        <span class="fc-col-rank">
          <span class="fc-rank-badge" :class="rankClass(item.rank)">{{ item.rank }}</span>
        </span>
        <span class="fc-col-name">{{ item.user_name }}</span>
        <span class="fc-col-points fc-points">
          {{ pointsContributed(item.total_fund_amount) }} {{ t('common.pointsUnit') }}
        </span>
        <span class="fc-col-money fc-money">{{ formatVND(item.total_fund_amount) }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Loading, Medal } from '@element-plus/icons-vue'
import { formatVND } from '@/utils/formatters'
import type { FundContributor } from '@/types/settlement'

const { t } = useI18n()

interface Props {
  contributors: FundContributor[]
  loading?: boolean
  pointToVnd: number
}

const props = withDefaults(defineProps<Props>(), { loading: false })

function pointsContributed(fundAmount: number): string {
  if (!props.pointToVnd) return '0'
  return (fundAmount / props.pointToVnd).toFixed(1)
}

function rankClass(rank: number) {
  if (rank === 1) return 'fc-rank--gold'
  if (rank === 2) return 'fc-rank--silver'
  if (rank === 3) return 'fc-rank--bronze'
  return 'fc-rank--plain'
}
</script>

<style scoped>
.fc-loading {
  display: flex; justify-content: center; padding: 40px 0;
}

.fc-empty {
  display: flex; flex-direction: column; align-items: center;
  gap: 8px; padding: 32px 0;
  font-size: 13px; color: var(--text-muted);
}
.fc-empty-icon { color: var(--text-muted); }

.fc-list { display: flex; flex-direction: column; }

.fc-header {
  display: grid;
  grid-template-columns: 40px 1fr 90px 100px;
  padding: 4px 10px 8px;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.fc-row {
  display: grid;
  grid-template-columns: 40px 1fr 90px 100px;
  align-items: center;
  padding: 8px 10px;
  border-radius: 8px;
  transition: background 0.15s;
}
.fc-row:hover { background: var(--surface-page); }

.fc-col-rank { display: flex; align-items: center; }
.fc-col-name { font-size: 13px; font-weight: 600; color: var(--text-primary); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.fc-col-points { font-size: 12px; font-weight: 700; text-align: right; }
.fc-col-money  { font-size: 12px; font-weight: 700; text-align: right; }

.fc-rank-badge {
  display: inline-flex; align-items: center; justify-content: center;
  width: 24px; height: 24px; border-radius: 6px;
  font-size: 11px; font-weight: 800;
}
.fc-rank--gold   { background: #fef3c7; color: #d97706; }
.fc-rank--silver { background: #f1f5f9; color: #64748b; }
.fc-rank--bronze { background: #fef0e7; color: #c2672a; }
.fc-rank--plain  { background: var(--surface-page); color: var(--text-muted); }

.fc-points { color: var(--color-warning); }
.fc-money  { color: var(--color-success); }
</style>
