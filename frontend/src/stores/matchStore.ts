import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import type { Match, CreateMatchRequest, MatchStats } from '@/types/match'
import { matchService } from '@/services/matchService'
import { getErrorMessage, translate } from '@/utils/i18n'

export const useMatchStore = defineStore('match', () => {
  const matches = ref<Match[]>([])
  const stats = ref<MatchStats>({ total: 0, today: 0 })
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Getters
  const todayMatches = computed(() =>
    matches.value.filter(
      (m) => new Date(m.match_date).toDateString() === new Date().toDateString()
    )
  )

  const lockedMatches = computed(() => matches.value.filter((m) => m.is_locked))

  const recentMatches = computed(() => matches.value.slice(0, 5))

  // Actions
  async function fetchMatches(params?: { page?: number; limit?: number }) {
    loading.value = true
    error.value = null
    try {
      matches.value = await matchService.getAll(params)
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch matches'
      if (error.value) ElMessage.error(error.value)
    } finally {
      loading.value = false
    }
  }

  async function fetchStats() {
    try {
      stats.value = await matchService.getStats()
    } catch (err: any) {
      console.error('Failed to fetch match stats:', err)
    }
  }

  async function createMatch(data: CreateMatchRequest) {
    loading.value = true
    error.value = null
    try {
      const newMatch = await matchService.create(data)
      matches.value.unshift(newMatch)
      stats.value.total++
      if (new Date(newMatch.match_date).toDateString() === new Date().toDateString()) {
        stats.value.today++
      }
      ElMessage.success(translate('toast.matchCreated'))
      return newMatch
    } catch (err: any) {
      const errorMsg = getErrorMessage(err)
      error.value = errorMsg
      ElMessage.error(errorMsg)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function deleteMatch(id: string) {
    loading.value = true
    error.value = null
    try {
      await matchService.delete(id)
      const index = matches.value.findIndex((m) => m.id === id)
      if (index !== -1) {
        matches.value.splice(index, 1)
      }
      ElMessage.success(translate('toast.matchDeleted'))
    } catch (err: any) {
      const errorMsg = getErrorMessage(err)
      error.value = errorMsg
      ElMessage.error(errorMsg)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchMatchesByUser(userId: string) {
    loading.value = true
    error.value = null
    try {
      matches.value = await matchService.getByUser(userId)
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch user matches'
      if (error.value) ElMessage.error(error.value)
    } finally {
      loading.value = false
    }
  }

  return {
    matches,
    stats,
    loading,
    error,
    todayMatches,
    lockedMatches,
    recentMatches,
    fetchMatches,
    fetchStats,
    createMatch,
    deleteMatch,
    fetchMatchesByUser,
  }
})
