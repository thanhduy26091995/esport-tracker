import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import type { DebtSettlement, SettlementStats } from '@/types/settlement'
import { settlementService } from '@/services/settlementService'
import { getErrorMessage, translate } from '@/utils/i18n'

export const useSettlementStore = defineStore('settlement', () => {
  const settlements = ref<DebtSettlement[]>([])
  const stats = ref<SettlementStats>({ total: 0, today: 0 })
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Getters
  const todaySettlements = computed(() =>
    settlements.value.filter(
      (s) => new Date(s.settlement_date).toDateString() === new Date().toDateString()
    )
  )

  const recentSettlements = computed(() => settlements.value.slice(0, 5))

  // Actions
  async function fetchSettlements(params?: { page?: number; limit?: number }) {
    loading.value = true
    error.value = null
    try {
      settlements.value = await settlementService.getAll(params)
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch settlements'
      if (error.value) ElMessage.error(error.value)
    } finally {
      loading.value = false
    }
  }

  async function fetchStats() {
    try {
      stats.value = await settlementService.getStats()
    } catch (err: any) {
      console.error('Failed to fetch settlement stats:', err)
    }
  }

  async function fetchSettlementById(id: string) {
    loading.value = true
    error.value = null
    try {
      return await settlementService.getById(id)
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch settlement'
      if (error.value) ElMessage.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchSettlementsByDebtor(debtorId: string) {
    loading.value = true
    error.value = null
    try {
      settlements.value = await settlementService.getByDebtorId(debtorId)
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch debtor settlements'
      if (error.value) ElMessage.error(error.value)
    } finally {
      loading.value = false
    }
  }

  async function triggerSettlement(debtorId: string, winnerIds?: string[]) {
    loading.value = true
    error.value = null
    try {
      const settlement = await settlementService.trigger({
        debtor_id: debtorId,
        winner_ids: winnerIds,
      })
      settlements.value.unshift(settlement)
      stats.value.total++
      if (new Date(settlement.settlement_date).toDateString() === new Date().toDateString()) {
        stats.value.today++
      }
      ElMessage.success(translate('toast.settlementTriggered'))
      return settlement
    } catch (err: any) {
      const errorMsg = getErrorMessage(err)
      error.value = errorMsg
      ElMessage.error(errorMsg)
      throw err
    } finally {
      loading.value = false
    }
  }

  return {
    settlements,
    stats,
    loading,
    error,
    todaySettlements,
    recentSettlements,
    fetchSettlements,
    fetchStats,
    fetchSettlementById,
    fetchSettlementsByDebtor,
    triggerSettlement,
  }
})
