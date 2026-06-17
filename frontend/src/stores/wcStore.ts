import { defineStore } from 'pinia'
import { ref } from 'vue'
import { wcService } from '@/services/wcService'
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

export const useWcStore = defineStore('wc', () => {
  const config = ref<WcConfig | null>(null)
  const isEnabled = ref<boolean>(true)
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

  async function fetchPublicConfig() {
    try {
      const res = await wcService.getPublicConfig()
      isEnabled.value = res.is_enabled
    } catch { /* silently */ }
  }

  async function fetchConfig() {
    try {
      config.value = await wcService.getConfig()
    } catch { /* admin-only, silently fail */ }
  }

  async function updateConfig(isEnabled: boolean) {
    await wcService.updateConfig(isEnabled)
    if (config.value) config.value.is_enabled = isEnabled
    ElMessage.success(isEnabled ? 'Đã bật tính năng World Cup' : 'Đã tắt tính năng World Cup')
  }

  async function fetchMatches(filter: WcMatchFilter = {}) {
    loading.value = true
    try {
      matches.value = (await wcService.listMatches(filter)) ?? []
    } finally {
      loading.value = false
    }
  }

  async function fetchMatch(id: string) {
    loading.value = true
    try {
      currentMatch.value = await wcService.getMatch(id)
    } finally {
      loading.value = false
    }
  }

  async function syncMatches() {
    loading.value = true
    try {
      const res = await wcService.syncMatches()
      ElMessage.success(`Đồng bộ thành công: ${res.synced} trận`)
      await fetchMatches()
    } finally {
      loading.value = false
    }
  }

  async function openMatch(id: string) {
    await wcService.openMatch(id)
    ElMessage.success('Đã mở cược trận đấu')
    await fetchMatches()
  }

  async function closeMatch(id: string) {
    await wcService.closeMatch(id)
    ElMessage.success('Đã đóng cược trận đấu')
    await fetchMatches()
  }

  async function finalizeMatch(id: string) {
    const res = await wcService.finalizeMatch(id)
    ElMessage.success(`Đã tính kết quả: ${res.predictions_processed} dự đoán, tổng điểm ${res.total_points_awarded}`)
    await fetchMatches()
  }

  async function finalizeAll() {
    const res = await wcService.finalizeAllMatches()
    ElMessage.success(`Tính điểm toàn bộ: ${res.processed} trận, bỏ qua ${res.skipped}, tổng ${res.total_points_awarded} điểm`)
    await fetchMatches()
  }

  async function fetchBets() {
    try {
      bets.value = (await wcService.listBets()) ?? []
    } catch { /* silently */ }
  }

  async function fetchMatchBets(matchId: string) {
    try {
      matchBets.value = (await wcService.getMatchBets(matchId)) ?? []
    } catch { /* silently */ }
  }

  async function updateBetStake(id: string, stake: number) {
    await wcService.updateBetStake(id, stake)
    await fetchBets()
  }

  async function deleteBet(id: string) {
    await wcService.deleteBet(id)
    bets.value = bets.value.filter(b => b.id !== id)
  }

  async function fetchWallet() {
    try {
      wallet.value = await wcService.getWallet()
    } catch { /* silently */ }
  }

  async function fetchPredictions() {
    loading.value = true
    try {
      predictions.value = (await wcService.listPredictions()) ?? []
    } finally {
      loading.value = false
    }
  }

  async function deletePrediction(id: string) {
    await wcService.deletePrediction(id)
    predictions.value = predictions.value.filter(p => p.id !== id)
    ElMessage.success('Đã xoá dự đoán')
  }

  async function updatePredictionPoints(id: string, points: number) {
    await wcService.updatePredictionPoints(id, points)
    const prediction = predictions.value.find(p => p.id === id)
    if (prediction) prediction.points = points
    ElMessage.success('Đã cập nhật điểm dự đoán')
  }

  async function fetchMatchPredictions(matchId: string) {
    matchPredictions.value = (await wcService.getMatchPredictions(matchId)) ?? []
  }

  async function fetchLeaderboard() {
    loading.value = true
    try {
      leaderboard.value = (await wcService.getLeaderboard()) ?? []
    } finally {
      loading.value = false
    }
  }

  async function fetchAllWallets() {
    allWallets.value = await wcService.getAllWallets()
  }

  async function fetchAllUsers() {
    allUsers.value = await wcService.listUsers()
  }

  async function topUp(wcUserId: string, delta: number, note?: string) {
    await wcService.topUp(wcUserId, delta, note)
    ElMessage.success('Đã cập nhật số dư')
    await fetchAllWallets()
  }

  async function setUserRole(wcUserId: string, isAdmin: boolean) {
    await wcService.setUserRole(wcUserId, isAdmin)
    ElMessage.success(isAdmin ? 'Đã cấp quyền Admin' : 'Đã thu hồi quyền Admin')
    await fetchAllUsers()
  }

  async function fetchSettlements() {
    settlements.value = await wcService.listSettlements()
  }

  async function fetchSettlement(id: string) {
    currentSettlement.value = await wcService.getSettlement(id)
  }

  async function previewSettlement(pointRate: number) {
    settlementPreview.value = await wcService.previewSettlement(pointRate)
  }

  async function createSettlement(name: string, pointRate: number, note?: string) {
    await wcService.createSettlement(name, pointRate, note)
    ElMessage.success('Đã tạo tất toán thành công')
    await fetchSettlements()
    await fetchAllWallets()
  }

  async function markSettlementDone(settlementId: string, wcUserId: string, doneNote?: string) {
    await wcService.markSettlementDone(settlementId, wcUserId, doneNote)
    ElMessage.success('Đã đánh dấu hoàn thành')
    if (currentSettlement.value?.id === settlementId) {
      await fetchSettlement(settlementId)
    }
  }

  return {
    config,
    isEnabled,
    matches,
    currentMatch,
    wallet,
    predictions,
    matchPredictions,
    bets,
    matchBets,
    leaderboard,
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
    fetchSettlements,
    fetchSettlement,
    previewSettlement,
    createSettlement,
    markSettlementDone,
  }
})
