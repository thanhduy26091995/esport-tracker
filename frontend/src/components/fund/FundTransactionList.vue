<template>
  <div>
    <div class="filter-bar">
      <el-select v-model="typeFilter" placeholder="Transaction Type" clearable class="w-44">
        <el-option label="All Types" value="" />
        <el-option label="Deposits" value="deposit" />
        <el-option label="Withdrawals" value="withdrawal" />
      </el-select>
      <span class="filter-count">{{ filteredTransactions.length }} of {{ transactions.length }}</span>
    </div>

    <div v-if="!loading && filteredTransactions.length === 0" class="empty-state">
      <svg class="empty-state-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 14l6-6m-5.5.5h.01m4.99 5h.01M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16l3.5-2 3.5 2 3.5-2 3.5 2z" />
      </svg>
      <p class="empty-state-title">{{ hasFilters ? 'No transactions found' : 'No transactions yet' }}</p>
      <p class="empty-state-desc">{{ hasFilters ? 'Try adjusting your filters' : 'Fund transactions will appear here' }}</p>
    </div>

    <div v-else class="tx-list">
      <div
        v-for="t in paginatedTransactions" :key="t.id"
        class="tx-row"
        :class="t.transaction_type === 'deposit' ? 'tx-row--deposit' : 'tx-row--withdraw'"
      >
        <div class="tx-icon" :class="t.transaction_type === 'deposit' ? 'tx-icon--deposit' : 'tx-icon--withdraw'">
          <el-icon :size="16"><component :is="t.transaction_type === 'deposit' ? Plus : Minus" /></el-icon>
        </div>
        <div class="tx-info">
          <div class="tx-top">
            <span class="tx-type" :class="t.transaction_type === 'deposit' ? 'tx-type--deposit' : 'tx-type--withdraw'">
              {{ t.transaction_type }}
            </span>
            <el-tag v-if="t.related_settlement_id" type="warning" size="small" effect="plain">Settlement</el-tag>
          </div>
          <p class="tx-desc">{{ t.description }}</p>
          <p class="tx-date">{{ formatDateTime(t.transaction_date) }}</p>
        </div>
        <div class="tx-amount" :class="t.transaction_type === 'deposit' ? 'tx-amount--deposit' : 'tx-amount--withdraw'">
          {{ t.transaction_type === 'deposit' ? '+' : '-' }}{{ formatVND(t.amount) }}
        </div>
      </div>
    </div>

    <div v-if="filteredTransactions.length > pageSize" class="mt-6 flex justify-center">
      <el-pagination v-model:current-page="currentPage" :page-size="pageSize" :total="filteredTransactions.length" layout="prev, pager, next" @current-change="handlePageChange" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Plus, Minus } from '@element-plus/icons-vue'
import type { FundTransaction } from '@/types/fund'
import { formatVND } from '@/utils/formatters'
import { formatDateTime } from '@/utils/date'

interface Props { transactions: FundTransaction[]; loading?: boolean }
const props = withDefaults(defineProps<Props>(), { loading: false })

const typeFilter = ref(''); const currentPage = ref(1); const pageSize = 20
const hasFilters = computed(() => !!typeFilter.value)
const filteredTransactions = computed(() => typeFilter.value ? props.transactions.filter(t => t.transaction_type === typeFilter.value) : props.transactions)
const paginatedTransactions = computed(() => { const s = (currentPage.value - 1) * pageSize; return filteredTransactions.value.slice(s, s + pageSize) })
const handlePageChange = (page: number) => { currentPage.value = page; window.scrollTo({ top: 0, behavior: 'smooth' }) }
</script>

<style scoped>
.tx-list { display: flex; flex-direction: column; gap: 8px; }

.tx-row {
  display: flex; align-items: center; gap: 14px;
  padding: 14px 16px; border-radius: 12px; border: 1px solid;
  transition: box-shadow 0.15s;
}
.tx-row:hover { box-shadow: var(--shadow-card); }
.tx-row--deposit  { background: #f0fdf4; border-color: #bbf7d0; }
.tx-row--withdraw { background: #fef2f2; border-color: #fecaca; }

.tx-icon {
  width: 38px; height: 38px; border-radius: 10px;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}
.tx-icon--deposit  { background: var(--color-success-bg); color: var(--color-success); }
.tx-icon--withdraw { background: var(--color-danger-bg);  color: var(--color-danger); }

.tx-info { flex: 1; min-width: 0; }

.tx-top { display: flex; align-items: center; gap: 8px; margin-bottom: 3px; }

.tx-type { font-size: 11px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.06em; }
.tx-type--deposit  { color: var(--color-success); }
.tx-type--withdraw { color: var(--color-danger); }

.tx-desc { font-size: 13px; color: var(--text-primary); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.tx-date { font-size: 11px; color: var(--text-muted); margin-top: 2px; }

.tx-amount { font-size: 15px; font-weight: 800; flex-shrink: 0; letter-spacing: -0.02em; }
.tx-amount--deposit  { color: var(--color-success); }
.tx-amount--withdraw { color: var(--color-danger); }
</style>
