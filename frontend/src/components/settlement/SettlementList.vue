<template>
  <div>
    <div class="filter-bar">
      <el-select v-model="dateFilter" placeholder="Date Range" clearable class="w-44">
        <el-option label="All Time" value="" />
        <el-option label="Today" value="today" />
        <el-option label="This Week" value="week" />
        <el-option label="This Month" value="month" />
      </el-select>
      <span class="filter-count">{{ filteredSettlements.length }} of {{ settlements.length }}</span>
    </div>

    <div v-if="!loading && filteredSettlements.length === 0" class="empty-state">
      <svg class="empty-state-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      <p class="empty-state-title">{{ hasFilters ? 'No settlements found' : 'No settlements yet' }}</p>
      <p class="empty-state-desc">{{ hasFilters ? 'Try adjusting your filters' : 'Settlements trigger automatically at the debt threshold' }}</p>
    </div>

    <div v-else class="sl-list">
      <div
        v-for="s in filteredSettlements" :key="s.id"
        class="sl-card"
        @click="emit('settlementClick', s)"
      >
        <div class="sl-header">
          <div class="sl-who">
            <div class="sl-who-icon">
              <el-icon :size="18" color="white"><Warning /></el-icon>
            </div>
            <div>
              <p class="sl-name">{{ s.debtor.name }}</p>
              <p class="sl-date">{{ formatDateTime(s.settlement_date) }}</p>
            </div>
          </div>
          <div class="sl-total">{{ formatVND(s.money_amount) }}</div>
        </div>

        <div class="sl-grid">
          <div class="sl-stat">
            <dt>Debt Points</dt>
            <dd class="sl-stat-val sl-stat-val--red">{{ s.original_debt_points }} pts</dd>
          </div>
          <div class="sl-stat">
            <dt>Total Amount</dt>
            <dd class="sl-stat-val">{{ formatVND(s.money_amount) }}</dd>
          </div>
          <div class="sl-stat">
            <dt>To Fund</dt>
            <dd class="sl-stat-val sl-stat-val--green">{{ formatVND(s.fund_amount) }}</dd>
          </div>
          <div class="sl-stat">
            <dt>To Winners</dt>
            <dd class="sl-stat-val sl-stat-val--blue">{{ formatVND(s.winner_distribution) }}</dd>
          </div>
        </div>

        <div v-if="s.winners?.length" class="sl-winners">
          <span v-for="w in s.winners" :key="w.id" class="sl-winner-pill">
            {{ w.winner.name }} · {{ formatVND(w.money_amount) }}
          </span>
        </div>

        <p class="sl-hint">Tap for details →</p>
      </div>
    </div>

    <div v-if="filteredSettlements.length > 0" class="mt-6 flex justify-center">
      <el-pagination v-model:current-page="currentPage" :page-size="pageSize" :total="filteredSettlements.length" layout="prev, pager, next" @current-change="handlePageChange" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Warning } from '@element-plus/icons-vue'
import type { DebtSettlement } from '@/types/settlement'
import { formatVND } from '@/utils/formatters'
import { formatDateTime, isToday } from '@/utils/date'

interface Props { settlements: DebtSettlement[]; loading?: boolean }
const props = withDefaults(defineProps<Props>(), { loading: false })
const emit = defineEmits<{ settlementClick: [settlement: DebtSettlement] }>()

const dateFilter = ref(''); const currentPage = ref(1); const pageSize = 20
const hasFilters = computed(() => !!dateFilter.value)

const filteredSettlements = computed(() => {
  if (!dateFilter.value) return props.settlements
  const now = new Date()
  return props.settlements.filter(s => {
    const d = new Date(s.settlement_date)
    if (dateFilter.value === 'today') return isToday(s.settlement_date)
    if (dateFilter.value === 'week') return d >= new Date(now.getTime() - 7 * 86400000)
    if (dateFilter.value === 'month') return d >= new Date(now.getTime() - 30 * 86400000)
    return true
  })
})

const handlePageChange = (page: number) => { currentPage.value = page; window.scrollTo({ top: 0, behavior: 'smooth' }) }
</script>

<style scoped>
.sl-list { display: flex; flex-direction: column; gap: 12px; }

.sl-card {
  background: var(--surface-card);
  border: 1px solid #fecaca;
  border-radius: 16px;
  padding: 18px;
  cursor: pointer;
  box-shadow: var(--shadow-card);
  transition: box-shadow 0.15s, transform 0.15s;
}
.sl-card:hover { box-shadow: var(--shadow-card-hover); transform: translateY(-1px); }

.sl-header { display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: 14px; }

.sl-who { display: flex; align-items: center; gap: 12px; }

.sl-who-icon {
  width: 40px; height: 40px; border-radius: 12px;
  background: linear-gradient(135deg, #ef4444, #b91c1c);
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0; box-shadow: 0 4px 10px rgba(239,68,68,0.3);
}

.sl-name { font-size: 15px; font-weight: 700; color: var(--text-primary); }
.sl-date { font-size: 11px; color: var(--text-muted); margin-top: 2px; }
.sl-total { font-size: 17px; font-weight: 800; color: var(--color-danger); }

.sl-grid {
  display: grid; grid-template-columns: repeat(2, 1fr); gap: 10px;
  background: var(--surface-page); border-radius: 12px; padding: 14px;
  margin-bottom: 12px;
}
@media (min-width: 640px) { .sl-grid { grid-template-columns: repeat(4, 1fr); } }

.sl-stat dt { font-size: 10px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.06em; color: var(--text-muted); margin-bottom: 4px; }

.sl-stat-val { font-size: 14px; font-weight: 700; color: var(--text-primary); }
.sl-stat-val--red   { color: var(--color-danger); }
.sl-stat-val--green { color: var(--color-success); }
.sl-stat-val--blue  { color: var(--color-info); }

.sl-winners { display: flex; flex-wrap: wrap; gap: 6px; margin-bottom: 10px; }

.sl-winner-pill {
  font-size: 11px; font-weight: 600;
  background: var(--color-info-bg); color: var(--color-info);
  border: 1px solid var(--color-info-border);
  padding: 4px 10px; border-radius: 20px;
}

.sl-hint { font-size: 11px; color: var(--text-muted); text-align: right; }
</style>
