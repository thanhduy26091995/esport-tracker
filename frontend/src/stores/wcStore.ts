import { defineStore } from 'pinia'
import { ref } from 'vue'
import { wcService } from '@/services/wcService'
import type {
  WcMatch,
  WcMatchWithOdds,
  WcWallet,
  WcWalletWithUser,
  WcBetWithMatch,
  WcBetPublic,
  WcLeaderboardEntry,
  WcConfig,
  WcUser,
  WcSettlement,
  WcSettlementWithDetails,
  WcSettlementPreviewRow,
  WcMatchFilter,
} from '@/types/wc'
import { ElMessage } from 'element-plus'

export const useWcStore = defineStore('wc', () => {
  const config = ref<WcConfig | null>(null)
  const matches = ref<WcMatch[]>([])
  const currentMatch = ref<WcMatchWithOdds | null>(null)
  const wallet = ref<WcWallet | null>(null)
  const bets = ref<WcBetWithMatch[]>([])
  const matchBets = ref<WcBetPublic[]>([])
  const leaderboard = ref<WcLeaderboardEntry[]>([])
  const allWallets = ref<WcWalletWithUser[]>([])
  const allUsers = ref<WcUser[]>([])
  const settlements = ref<WcSettlement[]>([])
  const currentSettlement = ref<WcSettlementWithDetails | null>(null)
  const settlementPreview = ref<WcSettlementPreviewRow[]>([])

  const loading = ref(false)

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

  async function lockMatch(id: string) {
    await wcService.lockMatch(id)
    ElMessage.success('Đã khoá cược trận đấu')
    await fetchMatches()
  }

  async function unlockMatch(id: string) {
    await wcService.unlockMatch(id)
    ElMessage.success('Đã mở cược trận đấu')
    await fetchMatches()
  }

  async function settleMatch(id: string) {
    const res = await wcService.settleMatch(id)
    ElMessage.success(`Đã tính kết quả: ${res.bets_processed} cược, tổng thưởng ${res.total_paid_out}`)
    await fetchMatches()
  }

  async function fetchWallet() {
    try {
      wallet.value = await wcService.getWallet()
    } catch { /* silently */ }
  }

  async function fetchBets() {
    loading.value = true
    try {
      bets.value = (await wcService.listBets()) ?? []
    } finally {
      loading.value = false
    }
  }

  async function deleteBet(id: string) {
    await wcService.deleteBet(id)
    bets.value = bets.value.filter(b => b.id !== id)
    ElMessage.success('Đã xoá cược')
  }

  async function updateBetStake(id: string, stake: number) {
    await wcService.updateBetStake(id, stake)
    const bet = bets.value.find(b => b.id === id)
    if (bet) bet.stake = stake
    ElMessage.success('Đã cập nhật số tiền cược')
  }

  async function fetchMatchBets(matchId: string) {
    matchBets.value = (await wcService.getMatchBets(matchId)) ?? []
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
    matches,
    currentMatch,
    wallet,
    bets,
    matchBets,
    leaderboard,
    allWallets,
    allUsers,
    settlements,
    currentSettlement,
    settlementPreview,
    loading,
    fetchConfig,
    updateConfig,
    fetchMatches,
    fetchMatch,
    syncMatches,
    lockMatch,
    unlockMatch,
    settleMatch,
    fetchWallet,
    fetchBets,
    deleteBet,
    updateBetStake,
    fetchMatchBets,
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
