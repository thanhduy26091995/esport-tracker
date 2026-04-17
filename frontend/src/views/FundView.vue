<template>
  <div class="page-wrapper">
    <div class="page-container">
      <!-- Header -->
      <div class="page-header">
        <div class="page-header-left">
          <h1 class="page-title">{{ t('fund.pageTitle') }}</h1>
          <p class="page-subtitle">{{ t('fund.pageSubtitle') }}</p>
        </div>
        <div class="flex gap-2">
          <el-button type="success" plain @click="handleDeposit" :icon="Plus">{{ t('fund.deposit') }}</el-button>
          <el-button type="danger" plain @click="handleWithdraw" :icon="Minus" :disabled="fundStore.balance === 0">{{ t('fund.withdraw') }}</el-button>
        </div>
      </div>

      <!-- Balance hero -->
      <div class="balance-hero">
        <div class="balance-deco balance-deco--1" />
        <div class="balance-deco balance-deco--2" />
        <div class="balance-content">
          <div>
            <p class="balance-label">{{ t('fund.currentBalance') }}</p>
            <p class="balance-value">{{ formatVND(fundStore.balance) }}</p>
            <p class="balance-sub">{{ t('fund.balanceSubtitle') }}</p>
          </div>
          <div class="balance-icon">
            <el-icon :size="32" color="white"><Wallet /></el-icon>
          </div>
        </div>
      </div>

      <!-- Stats -->
      <div class="stats-grid-3 mb-6">
        <StatCard :title="t('fund.statTotalDeposits')"     :value="formatVND(fundStore.stats.total_deposits)"    :icon="TrendCharts" :loading="fundStore.loading" type="success" />
        <StatCard :title="t('fund.statTotalWithdrawals')"  :value="formatVND(fundStore.stats.total_withdrawals)" :icon="TrendCharts" :loading="fundStore.loading" type="danger"  />
        <StatCard :title="t('fund.statSettlementDeposits')" :value="fundStore.stats.settlement_deposits"          :icon="Document"   :loading="fundStore.loading" type="warning" />
      </div>

      <!-- Transactions -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">{{ t('fund.transactionHistory') }}</span>
        </div>
        <div class="card-body">
          <FundTransactionList :transactions="fundStore.transactions" :loading="fundStore.loading" />
        </div>
      </div>

      <FundForm v-model="showFundForm" :type="fundFormType" :current-balance="fundStore.balance"
        :loading="fundStore.loading" @submit="handleSubmitTransaction" @cancel="() => showFundForm = false" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { Plus, Minus, Wallet, TrendCharts, Document } from '@element-plus/icons-vue'
import { useFundStore } from '@/stores/fundStore'
import FundTransactionList from '@/components/fund/FundTransactionList.vue'
import FundForm from '@/components/fund/FundForm.vue'
import StatCard from '@/components/shared/StatCard.vue'
import { formatVND } from '@/utils/formatters'

const { t } = useI18n()
const fundStore = useFundStore()
const showFundForm = ref(false)
const fundFormType = ref<'deposit' | 'withdrawal'>('deposit')

onMounted(async () => { await Promise.all([fundStore.fetchStats(), fundStore.fetchTransactions()]) })

const handleDeposit = () => { fundFormType.value = 'deposit'; showFundForm.value = true }
const handleWithdraw = () => { fundFormType.value = 'withdrawal'; showFundForm.value = true }
const handleSubmitTransaction = async (data: { amount: number; description: string; date?: string }) => {
  try {
    fundFormType.value === 'deposit' ? await fundStore.deposit(data) : await fundStore.withdraw(data)
    showFundForm.value = false
  } catch {}
}
</script>

<style scoped>
.stats-grid-3 {
  display: grid;
  grid-template-columns: repeat(1, 1fr);
  gap: 16px;
}
@media (min-width: 640px) { .stats-grid-3 { grid-template-columns: repeat(3, 1fr); } }

.balance-hero {
  position: relative;
  overflow: hidden;
  border-radius: 20px;
  background: linear-gradient(135deg, #1e40af 0%, #2563eb 50%, #3b82f6 100%);
  padding: 32px;
  margin-bottom: 24px;
  box-shadow: 0 8px 32px rgba(37, 99, 235, 0.35);
}

.balance-deco {
  position: absolute;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.08);
}
.balance-deco--1 { width: 200px; height: 200px; top: -60px; right: -40px; }
.balance-deco--2 { width: 120px; height: 120px; bottom: -40px; right: 80px; }

.balance-content {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.balance-label {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: rgba(255,255,255,0.65);
  margin-bottom: 8px;
}

.balance-value {
  font-size: 36px;
  font-weight: 800;
  color: #ffffff;
  letter-spacing: -0.03em;
  line-height: 1.1;
}

.balance-sub {
  margin-top: 6px;
  font-size: 13px;
  color: rgba(255,255,255,0.55);
}

.balance-icon {
  width: 64px; height: 64px;
  border-radius: 16px;
  background: rgba(255,255,255,0.15);
  backdrop-filter: blur(4px);
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
}
</style>
