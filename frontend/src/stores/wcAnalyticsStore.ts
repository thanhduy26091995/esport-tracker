import { defineStore } from 'pinia'
import { ref } from 'vue'
import { wcAnalyticsService } from '@/services/wcAnalyticsService'
import type { MyAnalyticsResponse, CommunityAnalyticsResponse, CompareAnalyticsResponse, WcAnalyticsResponse } from '@/types/wc'

export const useWcAnalyticsStore = defineStore('wcAnalytics', () => {
  const myData = ref<MyAnalyticsResponse | null>(null)
  const communityData = ref<CommunityAnalyticsResponse | null>(null)
  const compareData = ref<CompareAnalyticsResponse | null>(null)
  const wc2026Data = ref<WcAnalyticsResponse | null>(null)
  const myPeriod = ref('30d')
  const myDateFrom = ref<string | undefined>(undefined)
  const myDateTo = ref<string | undefined>(undefined)
  const loading = ref(false)
  const wc2026Loading = ref(false)
  const error = ref<string | null>(null)

  async function loadMyAnalytics(period?: string, dateFrom?: string, dateTo?: string) {
    if (period) myPeriod.value = period
    myDateFrom.value = dateFrom
    myDateTo.value = dateTo
    loading.value = true
    error.value = null
    try {
      myData.value = await wcAnalyticsService.getMyAnalytics(myPeriod.value, myDateFrom.value, myDateTo.value)
    } catch {
      error.value = 'Failed to load analytics'
    } finally {
      loading.value = false
    }
  }

  async function loadCommunityAnalytics() {
    loading.value = true
    error.value = null
    try {
      communityData.value = await wcAnalyticsService.getCommunityAnalytics()
    } catch {
      error.value = 'Failed to load community analytics'
    } finally {
      loading.value = false
    }
  }

  async function loadCompareAnalytics() {
    loading.value = true
    error.value = null
    try {
      compareData.value = await wcAnalyticsService.getCompareAnalytics()
    } catch {
      error.value = 'Failed to load compare analytics'
    } finally {
      loading.value = false
    }
  }

  async function loadAll() {
    loading.value = true
    error.value = null
    try {
      const [my, community, compare] = await Promise.all([
        wcAnalyticsService.getMyAnalytics(myPeriod.value, myDateFrom.value, myDateTo.value),
        wcAnalyticsService.getCommunityAnalytics(),
        wcAnalyticsService.getCompareAnalytics(),
      ])
      myData.value = my
      communityData.value = community
      compareData.value = compare
    } catch {
      error.value = 'Failed to load analytics'
    } finally {
      loading.value = false
    }
  }

  async function loadWC2026Analytics() {
    wc2026Loading.value = true
    try {
      wc2026Data.value = await wcAnalyticsService.getWC2026Analytics()
    } catch {
      // fail silently — component shows empty state
    } finally {
      wc2026Loading.value = false
    }
  }

  return {
    myData,
    communityData,
    compareData,
    wc2026Data,
    myPeriod,
    myDateFrom,
    myDateTo,
    loading,
    wc2026Loading,
    error,
    loadMyAnalytics,
    loadCommunityAnalytics,
    loadCompareAnalytics,
    loadWC2026Analytics,
    loadAll,
  }
})
