import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import { wcService } from '@/services/wcService'
import { acService } from '@/services/acService'
import type {
  WcMatch,
  WcMatchWithOdds,
  WcWallet,
  WcWalletWithUser,
  WcPredictionWithMatch,
  WcPredictionPublic,
  WcLeaderboardEntry,
  WcConfig,
  WcUser,
  WcSettlement,
  WcSettlementWithDetails,
  WcSettlementPreviewRow,
  WcMatchFilter,
  WcBetWithMatch,
  WcBetPublic,
} from '@/types/wc'
import { ElMessage } from 'element-plus'

type TournamentService = typeof wcService

export function createTournamentStore(id: string, service: TournamentService) {
  return defineStore(id, () => {
    const config = ref<WcConfig | null>(null)
    const isEnabled = ref<boolean>(true)
    const minPoints = ref<number>(1)
    const maxPoints = ref<number>(5)
    const cancelPenaltyEnabled = ref<boolean>(false)
    const cancelPenaltyPercent = ref<number>(20)
    const betReduceMaxPercent = ref<number>(50)
    const betReducePenaltyPercent = ref<number>(20)
    const matches = ref<WcMatch[]>([])
    const currentMatch = ref<WcMatchWithOdds | null>(null)
    const wallet = ref<WcWallet | null>(null)
    const predictions = ref<WcPredictionWithMatch[]>([])
    const matchPredictions = ref<WcPredictionPublic[]>([])
    const leaderboard = ref<WcLeaderboardEntry[]>([])
    const allWallets = ref<WcWalletWithUser[]>([])
    const allUsers = ref<WcUser[]>([])
    const settlements = ref<WcSettlement[]>([])
    const currentSettlement = ref<WcSettlementWithDetails | null>(null)
    const settlementPreview = ref<WcSettlementPreviewRow[]>([])

    const bets = ref<WcBetWithMatch[]>([])
    const matchBets = ref<WcBetPublic[]>([])

    const loading = ref(false)
    const leaderboardLoading = ref(false)

    async function fetchPublicConfig() {
      try {
        const res = await service.getPublicConfig()
        isEnabled.value = res.is_enabled
        minPoints.value = res.min_points ?? 1
        maxPoints.value = res.max_points ?? 5
        cancelPenaltyEnabled.value = res.cancel_penalty_enabled ?? false
        cancelPenaltyPercent.value = res.cancel_penalty_percent ?? 20
        betReduceMaxPercent.value = res.bet_reduce_max_percent ?? 50
        betReducePenaltyPercent.value = res.bet_reduce_penalty_percent ?? 20
      } catch { /* silently */ }
    }

    async function fetchConfig() {
      try {
        config.value = await service.getConfig()
      } catch { /* admin-only, silently fail */ }
    }

    async function updateConfig(isEnabledVal: boolean) {
      await service.updateConfig(isEnabledVal)
      if (config.value) config.value.is_enabled = isEnabledVal
      ElMessage.success(isEnabledVal ? 'Đã bật tính năng' : 'Đã tắt tính năng')
    }

    async function fetchMatches(filter: WcMatchFilter = {}) {
      loading.value = true
      try {
        matches.value = (await service.listMatches(filter)) ?? []
      } finally {
        loading.value = false
      }
    }

    async function fetchMatch(matchId: string) {
      loading.value = true
      try {
        currentMatch.value = await service.getMatch(matchId)
      } finally {
        loading.value = false
      }
    }

    async function syncMatches() {
      loading.value = true
      try {
        const res = await service.syncMatches()
        ElMessage.success(`Đồng bộ thành công: ${res.synced} trận`)
        await fetchMatches()
      } finally {
        loading.value = false
      }
    }

    async function openMatch(matchId: string) {
      await service.openMatch(matchId)
      ElMessage.success('Đã mở cược trận đấu')
      await fetchMatches()
    }

    async function closeMatch(matchId: string) {
      await service.closeMatch(matchId)
      ElMessage.success('Đã đóng cược trận đấu')
      await fetchMatches()
    }

    async function finalizeMatch(matchId: string) {
      const res = await service.finalizeMatch(matchId)
      ElMessage.success(`Đã tính kết quả: ${res.predictions_processed} dự đoán, tổng điểm ${res.total_points_awarded}`)
      if (res.unsettled_custom_bets_count > 0) {
        ElMessage.warning(`Trận này còn ${res.unsettled_custom_bets_count} kèo phụ chưa tất toán`)
      }
      await fetchMatches()
    }

    async function finalizeAll() {
      const res = await service.finalizeAllMatches()
      ElMessage.success(`Tính điểm toàn bộ: ${res.processed} trận, bỏ qua ${res.skipped}, tổng ${res.total_points_awarded} điểm`)
      if (res.matches_with_unsettled_custom_bets > 0) {
        ElMessage.warning(`Còn ${res.matches_with_unsettled_custom_bets} trận có kèo phụ chưa tất toán`)
      }
      await fetchMatches()
    }

    async function refinalizeAll() {
      const res = await service.refinalizeAllMatches()
      ElMessage.success(`Tính lại toàn bộ: ${res.processed} trận, bỏ qua ${res.skipped}, tổng ${res.total_points_awarded} điểm`)
      await fetchMatches()
    }

    async function fetchBets() {
      try {
        bets.value = (await service.listBets()) ?? []
      } catch { /* silently */ }
    }

    async function fetchMatchBets(matchId: string) {
      try {
        matchBets.value = (await service.getMatchBets(matchId)) ?? []
      } catch { /* silently */ }
    }

    async function updateBetStake(betId: string, stake: number) {
      await service.updateBetStake(betId, stake)
      await fetchBets()
    }

    async function deleteBet(betId: string) {
      await service.deleteBet(betId)
      await Promise.all([fetchBets(), fetchWallet()])
    }

    async function fetchWallet() {
      try {
        wallet.value = await service.getWallet()
      } catch { /* silently */ }
    }

    async function fetchPredictions() {
      loading.value = true
      try {
        predictions.value = (await service.listPredictions()) ?? []
      } finally {
        loading.value = false
      }
    }

    async function deletePrediction(predId: string) {
      await service.deletePrediction(predId)
      await Promise.all([fetchPredictions(), fetchWallet()])
      ElMessage.success('Đã huỷ dự đoán')
    }

    async function updatePredictionPoints(predId: string, points: number) {
      await service.updatePredictionPoints(predId, points)
      await Promise.all([fetchPredictions(), fetchWallet()])
      ElMessage.success('Đã cập nhật điểm dự đoán')
    }

    async function fetchMatchPredictions(matchId: string) {
      matchPredictions.value = (await service.getMatchPredictions(matchId)) ?? []
    }

    async function fetchLeaderboard() {
      leaderboardLoading.value = true
      try {
        leaderboard.value = (await service.getLeaderboard()) ?? []
      } finally {
        leaderboardLoading.value = false
      }
    }

    async function fetchAllWallets() {
      allWallets.value = await service.getAllWallets()
    }

    async function fetchAllUsers() {
      allUsers.value = await service.listUsers()
    }

    async function topUp(wcUserId: string, delta: number, note?: string) {
      await service.topUp(wcUserId, delta, note)
      ElMessage.success('Đã cập nhật số dư')
      await fetchAllWallets()
    }

    async function setUserRole(wcUserId: string, isAdmin: boolean) {
      await service.setUserRole(wcUserId, isAdmin)
      ElMessage.success(isAdmin ? 'Đã cấp quyền Admin' : 'Đã thu hồi quyền Admin')
      await fetchAllUsers()
    }

    async function setUserBot(wcUserId: string, isBot: boolean) {
      await service.setUserBot(wcUserId, isBot)
      ElMessage.success(isBot ? 'Đã đánh dấu là Bot' : 'Đã bỏ đánh dấu Bot')
      await fetchAllUsers()
    }

    async function fetchSettlements() {
      settlements.value = await service.listSettlements()
    }

    async function fetchSettlement(settlementId: string) {
      currentSettlement.value = await service.getSettlement(settlementId)
    }

    async function previewSettlement(pointRate: number) {
      settlementPreview.value = await service.previewSettlement(pointRate)
    }

    async function createSettlement(name: string, pointRate: number, note?: string) {
      await service.createSettlement(name, pointRate, note)
      ElMessage.success('Đã tạo tất toán thành công')
      await fetchSettlements()
      await fetchAllWallets()
    }

    async function markSettlementDone(settlementId: string, wcUserId: string, doneNote?: string) {
      await service.markSettlementDone(settlementId, wcUserId, doneNote)
      ElMessage.success('Đã đánh dấu hoàn thành')
      if (currentSettlement.value?.id === settlementId) {
        await fetchSettlement(settlementId)
      }
    }

    return {
      config,
      isEnabled,
      minPoints,
      maxPoints,
      cancelPenaltyEnabled,
      cancelPenaltyPercent,
      betReduceMaxPercent,
      betReducePenaltyPercent,
      matches,
      currentMatch,
      wallet,
      predictions,
      matchPredictions,
      bets,
      matchBets,
      leaderboard,
      leaderboardLoading,
      allWallets,
      allUsers,
      settlements,
      currentSettlement,
      settlementPreview,
      loading,
      fetchPublicConfig,
      fetchConfig,
      updateConfig,
      fetchMatches,
      fetchMatch,
      syncMatches,
      openMatch,
      closeMatch,
      finalizeMatch,
      finalizeAll,
      refinalizeAll,
      fetchBets,
      fetchMatchBets,
      updateBetStake,
      deleteBet,
      fetchWallet,
      fetchPredictions,
      deletePrediction,
      updatePredictionPoints,
      fetchMatchPredictions,
      fetchLeaderboard,
      fetchAllWallets,
      fetchAllUsers,
      topUp,
      setUserRole,
      setUserBot,
      fetchSettlements,
      fetchSettlement,
      previewSettlement,
      createSettlement,
      markSettlementDone,
    }
  })
}

const _useWcStoreImpl = createTournamentStore('wc', wcService)
const _useAcStoreImpl = createTournamentStore('ac', acService as unknown as TournamentService)

export function useWcStore() {
  const route = useRoute()
  return route.meta?.tournamentType === 'asean_cup' ? _useAcStoreImpl() : _useWcStoreImpl()
}
