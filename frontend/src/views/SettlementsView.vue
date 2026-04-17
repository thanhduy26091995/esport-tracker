<template>
  <div class="page-wrapper">
    <div class="page-container">
      <!-- Header -->
      <div class="page-header">
        <div class="page-header-left">
          <h1 class="page-title">{{ t('settlements.pageTitle') }}</h1>
          <p class="page-subtitle">{{ t('settlements.pageSubtitle') }}</p>
        </div>
      </div>

      <!-- Stats -->
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
        <StatCard
          :title="t('settlements.statTotal')"
          :value="settlementStore.stats.total"
          :icon="Document"
          :loading="settlementStore.loading"
          type="info"
        />
        <StatCard
          :title="t('settlements.statToday')"
          :value="settlementStore.stats.today"
          :icon="Calendar"
          :loading="settlementStore.loading"
          type="warning"
        />
        <StatCard
          :title="t('settlements.statCurrentDebtors')"
          :value="currentDebtors"
          :icon="Warning"
          :loading="userStore.loading"
          :type="currentDebtors > 0 ? 'danger' : 'success'"
        />
      </div>

      <!-- Info banner -->
      <div class="flex items-start gap-3 bg-blue-50 border border-blue-100 rounded-xl px-4 py-3 mb-6 text-sm text-blue-700">
        <el-icon class="mt-0.5 flex-shrink-0"><InfoFilled /></el-icon>
        <div>
          <p class="font-medium mb-2">{{ t('settlements.infoTitle') }}</p>
          <ul class="list-disc ml-5 space-y-1">
            <li>
              <strong>{{ t('settlements.autoTriggerLabel') }}</strong>
              {{ t('settlements.autoTriggerDesc', { threshold: configStore.debtThreshold }) }}
            </li>
            <li>
              <strong>{{ t('settlements.manualTriggerLabel') }}</strong>
              {{ t('settlements.manualTriggerDesc', { action: t('users.settleDebt') }) }}
            </li>
            <li>
              <strong>{{ t('settlements.processLabel') }}</strong>
              <ul class="list-circle ml-5 mt-1">
                <li>{{ t('settlements.processDebtorPays') }}</li>
                <li>{{ t('settlements.processFundShare', { percent: configStore.fundSplitPercent }) }}</li>
                <li>{{ t('settlements.processWinnerShare', { percent: 100 - configStore.fundSplitPercent }) }}</li>
                <li>{{ t('settlements.processDebtorReset') }}</li>
                <li>{{ t('settlements.processWinnerReduction') }}</li>
              </ul>
            </li>
          </ul>
        </div>
      </div>

      <!-- Settlement List -->
      <div class="card">
        <div class="card-body">
          <SettlementList
            :settlements="settlementStore.settlements"
            :loading="settlementStore.loading"
            @settlement-click="handleSettlementClick"
          />
        </div>
      </div>

      <!-- Settlement Details Dialog -->
      <SettlementDetails
        v-model="showDetails"
        :settlement="selectedSettlement"
        :fund-split-percent="configStore.fundSplitPercent"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { Document, Calendar, Warning, InfoFilled } from '@element-plus/icons-vue'
import { useSettlementStore } from '@/stores/settlementStore'
import { useUserStore } from '@/stores/userStore'
import { useConfigStore } from '@/stores/configStore'
import SettlementList from '@/components/settlement/SettlementList.vue'
import SettlementDetails from '@/components/settlement/SettlementDetails.vue'
import StatCard from '@/components/shared/StatCard.vue'
import type { DebtSettlement } from '@/types/settlement'

const { t } = useI18n()
const settlementStore = useSettlementStore()
const userStore = useUserStore()
const configStore = useConfigStore()
const showDetails = ref(false)
const selectedSettlement = ref<DebtSettlement | null>(null)

const currentDebtors = computed(() =>
  userStore.users.filter(u => u.is_active && u.current_score <= configStore.debtThreshold).length
)

onMounted(async () => {
  await Promise.all([
    settlementStore.fetchSettlements(),
    settlementStore.fetchStats(),
    userStore.fetchUsers(),
    configStore.fetchConfigs()
  ])
})

const handleSettlementClick = (settlement: DebtSettlement) => {
  selectedSettlement.value = settlement
  showDetails.value = true
}
</script>
