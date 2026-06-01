import { wcApi } from './wcApi'
import type {
  WcConfig,
  WcMatch,
  WcMatchWithOdds,
  WcScoreOdds,
  WcWallet,
  WcWalletWithUser,
  WcWalletLog,
  WcBet,
  WcBetWithMatch,
  WcBetPublic,
  WcLeaderboardEntry,
  WcSettlement,
  WcSettlementWithDetails,
  WcSettlementPreviewRow,
  WcUser,
  WcMatchFilter,
  WcPlaceBetRequest,
} from '@/types/wc'

export const wcService = {
  // --- Config ---
  async getConfig(): Promise<WcConfig> {
    const r = await wcApi.get<WcConfig>('/admin/config')
    return r.data
  },
  async updateConfig(isEnabled: boolean): Promise<void> {
    await wcApi.put('/admin/config', { is_enabled: isEnabled })
  },

  // --- Matches ---
  async listMatches(filter: WcMatchFilter = {}): Promise<WcMatch[]> {
    const r = await wcApi.get<WcMatch[]>('/matches', { params: filter })
    return r.data
  },
  async getMatch(id: string): Promise<WcMatchWithOdds> {
    const r = await wcApi.get<WcMatchWithOdds>(`/matches/${id}`)
    return r.data
  },
  async syncMatches(): Promise<{ synced: number }> {
    const r = await wcApi.post<{ synced: number }>('/admin/sync')
    return r.data
  },
  async updateMatch(id: string, fields: Record<string, unknown>): Promise<void> {
    await wcApi.put(`/admin/matches/${id}`, fields)
  },
  async openMatch(id: string): Promise<void> {
    await wcApi.post(`/admin/matches/${id}/open`)
  },
  async closeMatch(id: string): Promise<void> {
    await wcApi.post(`/admin/matches/${id}/close`)
  },
  async settleMatch(id: string): Promise<{ bets_processed: number; total_paid_out: number }> {
    const r = await wcApi.post(`/admin/matches/${id}/settle`)
    return r.data
  },

  // --- Score Odds ---
  async getScoreOdds(matchId: string): Promise<WcScoreOdds[]> {
    const r = await wcApi.get<WcScoreOdds[]>(`/matches/${matchId}/score-odds`)
    return r.data
  },
  async addScoreOdds(matchId: string, homeScore: number, awayScore: number, odds: number): Promise<WcScoreOdds> {
    const r = await wcApi.post<WcScoreOdds>(`/admin/matches/${matchId}/score-odds`, {
      home_score: homeScore,
      away_score: awayScore,
      odds,
    })
    return r.data
  },
  async updateScoreOdds(id: string, odds: number): Promise<void> {
    await wcApi.put(`/admin/score-odds/${id}`, { odds })
  },
  async deleteScoreOdds(id: string): Promise<void> {
    await wcApi.delete(`/admin/score-odds/${id}`)
  },

  // --- Wallet ---
  async getWallet(): Promise<WcWallet> {
    const r = await wcApi.get<WcWallet>('/wallet')
    return r.data
  },
  async getAllWallets(): Promise<WcWalletWithUser[]> {
    const r = await wcApi.get<WcWalletWithUser[]>('/admin/wallets')
    return r.data
  },
  async topUp(wcUserId: string, delta: number, note?: string): Promise<void> {
    await wcApi.put(`/admin/wallets/${wcUserId}`, { delta, note })
  },
  async getWalletLogs(wcUserId: string): Promise<WcWalletLog[]> {
    const r = await wcApi.get<WcWalletLog[]>(`/admin/wallets/${wcUserId}/logs`)
    return r.data
  },

  // --- Bets ---
  async placeBet(req: WcPlaceBetRequest): Promise<WcBet> {
    const r = await wcApi.post<WcBet>('/bets', req)
    return r.data
  },
  async listBets(): Promise<WcBetWithMatch[]> {
    const r = await wcApi.get<WcBetWithMatch[]>('/bets')
    return r.data
  },
  async deleteBet(id: string): Promise<void> {
    await wcApi.delete(`/bets/${id}`)
  },
  async updateBetStake(id: string, stake: number): Promise<void> {
    await wcApi.put(`/bets/${id}`, { stake })
  },
  async getMatchBets(matchId: string): Promise<WcBetPublic[]> {
    const r = await wcApi.get<WcBetPublic[]>(`/matches/${matchId}/bets`)
    return r.data
  },

  // --- Leaderboard ---
  async getLeaderboard(): Promise<WcLeaderboardEntry[]> {
    const r = await wcApi.get<WcLeaderboardEntry[]>('/leaderboard')
    return r.data
  },

  // --- Users (admin) ---
  async listUsers(): Promise<WcUser[]> {
    const r = await wcApi.get<WcUser[]>('/admin/users')
    return r.data
  },
  async setUserRole(wcUserId: string, isAdmin: boolean): Promise<void> {
    await wcApi.put(`/admin/users/${wcUserId}/role`, { is_admin: isAdmin })
  },

  // --- Settlements (admin) ---
  async previewSettlement(pointRate: number): Promise<WcSettlementPreviewRow[]> {
    const r = await wcApi.get<WcSettlementPreviewRow[]>('/admin/settlements/preview', {
      params: { point_rate: pointRate },
    })
    return r.data
  },
  async createSettlement(name: string, pointRate: number, note?: string): Promise<WcSettlement> {
    const r = await wcApi.post<WcSettlement>('/admin/settlements', { name, point_rate: pointRate, note })
    return r.data
  },
  async listSettlements(): Promise<WcSettlement[]> {
    const r = await wcApi.get<WcSettlement[]>('/admin/settlements')
    return r.data
  },
  async getSettlement(id: string): Promise<WcSettlementWithDetails> {
    const r = await wcApi.get<WcSettlementWithDetails>(`/admin/settlements/${id}`)
    return r.data
  },
  async markSettlementDone(settlementId: string, wcUserId: string, doneNote?: string): Promise<void> {
    await wcApi.put(`/admin/settlements/${settlementId}/details/${wcUserId}`, {
      status: 'done',
      done_note: doneNote || '',
    })
  },
}
