<template>
  <div class="wc-analytics-panel">
    <el-tabs v-model="analyticsTab" class="analytics-sub-tabs">
      <el-tab-pane v-if="!isAc" :label="t('wc.analytics.wc2026Tab')" name="tournament">
        <WcTournamentPanel
          :data="store.wc2026Data"
          :loading="store.wc2026Loading"
        />
      </el-tab-pane>
      <el-tab-pane :label="t('wc.analytics.myTab')" name="my">
        <MyAnalyticsPanel
          :data="store.myData"
          :loading="store.loading"
          @period-change="onPeriodChange"
        />
      </el-tab-pane>
      <el-tab-pane :label="t('wc.analytics.communityTab')" name="community">
        <CommunityPanel
          :data="store.communityData"
          :loading="store.loading"
        />
      </el-tab-pane>
      <el-tab-pane :label="t('wc.analytics.compareTab')" name="compare">
        <ComparePanel
          :data="store.compareData"
          :loading="store.loading"
        />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useWcAnalyticsStore } from '@/stores/wcAnalyticsStore'
import WcTournamentPanel from './analytics/WcTournamentPanel.vue'
import MyAnalyticsPanel from './analytics/MyAnalyticsPanel.vue'
import CommunityPanel from './analytics/CommunityPanel.vue'
import ComparePanel from './analytics/ComparePanel.vue'

const { t } = useI18n()
const route = useRoute()
const isAc = computed(() => route.meta?.tournamentType === 'asean_cup')
const store = useWcAnalyticsStore()
const analyticsTab = ref(isAc.value ? 'my' : 'tournament')

interface PeriodPayload {
  period: string
  dateFrom?: string
  dateTo?: string
}

async function onPeriodChange({ period, dateFrom, dateTo }: PeriodPayload) {
  await store.loadMyAnalytics(period, dateFrom, dateTo)
}

watch(analyticsTab, async (tab) => {
  if (tab === 'my' && !store.myData) await store.loadMyAnalytics()
  if (tab === 'community' && !store.communityData) await store.loadCommunityAnalytics()
  if (tab === 'compare' && !store.compareData) await store.loadCompareAnalytics()
})

onMounted(async () => {
  if (!isAc.value && !store.wc2026Data) await store.loadWC2026Analytics()
  if (!store.myData) await store.loadMyAnalytics()
})
</script>

<style scoped>
.wc-analytics-panel {
  padding-top: 4px;
}

.analytics-sub-tabs :deep(.el-tabs__nav-wrap::after) {
  height: 1px;
}

.analytics-sub-tabs :deep(.el-tabs__header) {
  margin-bottom: 16px;
}
</style>
