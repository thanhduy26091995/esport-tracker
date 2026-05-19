<template>
  <div>
    <div v-if="loading" class="wc-loading">
      <el-icon class="animate-spin" :size="24" style="color:var(--text-muted)"><Loading /></el-icon>
    </div>
    <div v-else-if="contributors.length === 0" class="wc-empty">
      <el-icon :size="36" class="wc-empty-icon"><Star /></el-icon>
      <p>{{ t('dashboard.noWinnerContributors') }}</p>
    </div>
    <div v-else class="wc-list">
      <div class="wc-header">
        <span class="wc-col-rank">{{ t('dashboard.colRank') }}</span>
        <span class="wc-col-name">{{ t('dashboard.colName') }}</span>
        <span class="wc-col-points">{{ t('dashboard.colPointsContributed') }}</span>
        <span class="wc-col-money">{{ t('dashboard.colMoneyContributed') }}</span>
      </div>
      <div v-for="item in contributors" :key="item.user_id" class="wc-row">
        <span class="wc-col-rank">
          <span class="wc-rank-badge" :class="rankClass(item.rank)">{{ item.rank }}</span>
        </span>
        <span class="wc-col-name">{{ item.user_name }}</span>
        <span class="wc-col-points wc-points">
          {{ item.total_points_contributed }} {{ t('common.pointsUnit') }}
        </span>
        <span class="wc-col-money wc-money">{{ formatVND(item.total_fund_amount) }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Loading, Star } from '@element-plus/icons-vue'
import { formatVND } from '@/utils/formatters'
import type { WinnerContributor } from '@/types/settlement'

const { t } = useI18n()

interface Props {
  contributors: WinnerContributor[]
  loading?: boolean
}

withDefaults(defineProps<Props>(), { loading: false })

function rankClass(rank: number) {
  if (rank === 1) return 'wc-rank--gold'
  if (rank === 2) return 'wc-rank--silver'
  if (rank === 3) return 'wc-rank--bronze'
  return 'wc-rank--plain'
}
</script>

<style scoped>
.wc-loading {
  display: flex; justify-content: center; padding: 40px 0;
}

.wc-empty {
  display: flex; flex-direction: column; align-items: center;
  gap: 8px; padding: 32px 0;
  font-size: 13px; color: var(--text-muted);
}
.wc-empty-icon { color: var(--text-muted); }

.wc-list { display: flex; flex-direction: column; }

.wc-header {
  display: grid;
  grid-template-columns: 40px 1fr 90px 100px;
  padding: 4px 10px 8px;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.wc-row {
  display: grid;
  grid-template-columns: 40px 1fr 90px 100px;
  align-items: center;
  padding: 8px 10px;
  border-radius: 8px;
  transition: background 0.15s;
}
.wc-row:hover { background: var(--surface-page); }

.wc-col-rank  { display: flex; align-items: center; }
.wc-col-name  { font-size: 13px; font-weight: 600; color: var(--text-primary); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.wc-col-points { font-size: 12px; font-weight: 700; text-align: right; }
.wc-col-money  { font-size: 12px; font-weight: 700; text-align: right; }

.wc-rank-badge {
  display: inline-flex; align-items: center; justify-content: center;
  width: 24px; height: 24px; border-radius: 6px;
  font-size: 11px; font-weight: 800;
}
.wc-rank--gold   { background: #fef3c7; color: #d97706; }
.wc-rank--silver { background: #f1f5f9; color: #64748b; }
.wc-rank--bronze { background: #fef0e7; color: #c2672a; }
.wc-rank--plain  { background: var(--surface-page); color: var(--text-muted); }

.wc-points { color: var(--color-danger); }
.wc-money  { color: var(--color-success); }
</style>
