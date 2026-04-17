import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Tournament, RecordMatchResultRequest, CreateTournamentRequest } from '@/types/tournament'
import { tournamentService } from '@/services/tournamentService'
import { ElMessage } from 'element-plus'
import { getErrorMessage, translate } from '@/utils/i18n'

export const useTournamentStore = defineStore('tournament', () => {
  const tournaments = ref<Tournament[]>([])
  const currentTournament = ref<Tournament | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchTournaments() {
    loading.value = true
    error.value = null
    try {
      tournaments.value = await tournamentService.getAll()
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch tournaments'
      ElMessage.error(error.value!)
    } finally {
      loading.value = false
    }
  }

  async function fetchTournament(id: string) {
    loading.value = true
    error.value = null
    try {
      currentTournament.value = await tournamentService.getById(id)
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch tournament'
      ElMessage.error(error.value!)
    } finally {
      loading.value = false
    }
  }

  async function createTournament(data: CreateTournamentRequest) {
    loading.value = true
    error.value = null
    try {
      const tournament = await tournamentService.create(data)
      tournaments.value.unshift(tournament)
      ElMessage.success(translate('toast.tournamentCreated', { name: tournament.name }))
      return tournament
    } catch (err: any) {
      const msg = getErrorMessage(err)
      error.value = msg
      ElMessage.error(msg)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function deleteTournament(id: string) {
    loading.value = true
    try {
      await tournamentService.delete(id)
      tournaments.value = tournaments.value.filter(t => t.id !== id)
      if (currentTournament.value?.id === id) currentTournament.value = null
      ElMessage.success(translate('toast.tournamentDeleted'))
    } catch (err: any) {
      const msg = getErrorMessage(err)
      ElMessage.error(msg)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function completeTournament(id: string) {
    loading.value = true
    try {
      const updated = await tournamentService.complete(id)
      if (currentTournament.value?.id === id) currentTournament.value = updated
      const idx = tournaments.value.findIndex(t => t.id === id)
      if (idx !== -1) tournaments.value[idx] = updated
      ElMessage.success(translate('toast.tournamentCompleted'))
      return updated
    } catch (err: any) {
      const msg = getErrorMessage(err)
      ElMessage.error(msg)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function recordResult(tournamentId: string, matchId: string, data: RecordMatchResultRequest) {
    loading.value = true
    try {
      await tournamentService.recordResult(tournamentId, matchId, data)
      await fetchTournament(tournamentId)
      ElMessage.success(translate('toast.resultRecorded'))
    } catch (err: any) {
      const msg = getErrorMessage(err)
      ElMessage.error(msg)
      throw err
    } finally {
      loading.value = false
    }
  }

  return {
    tournaments,
    currentTournament,
    loading,
    error,
    fetchTournaments,
    fetchTournament,
    createTournament,
    deleteTournament,
    completeTournament,
    recordResult,
  }
})
