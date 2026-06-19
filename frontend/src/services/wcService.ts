import { wcApi } from './wcApi'
import type {
  WcConfig,
  WcMatch,
  WcMatchWithOdds,
  WcScoreMultiplier,
  WcWallet,
  WcWalletWithUser,
  WcWalletLog,
  WcPrediction,
  WcPredictionWithMatch,
  WcPredictionPublic,
  WcLeaderboardEntry,
  WcSettlement,
  WcSettlementWithDetails,
  WcSettlementPreviewRow,
  WcUser,
  WcMatchFilter,
  WcSubmitPredictionRequest,
  WcBet,
  WcBetWithMatch,
  WcBetPublic,
  WcPlaceBetRequest,
  HousePnLResponse,
  WcSyncLog,
  MappingResult,
  ImportHandicapPreview,
  ImportOUPreview,
  GeneratePoissonPreview,
  WcChampionTeam,
  WcChampionConfig,
  WcChampionPredictionMine,
  WcChampionPredictionPublic,
  WcChampionSettleResult,
} from '@/types/wc'

export const wcService = {
  // --- Config ---
  async getPublicConfig(): Promise<{ is_enabled: boolean }> {
    const r = await wcApi.get<{ is_enabled: boolean }>('/config')
    return r.data
  },
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
    const match = r.data
    if (!match.score_odds) {
      match.score_odds = (match.score_multipliers ?? []).map(sm => ({
        id: sm.id,
        home_score: sm.home_score,
        away_score: sm.away_score,
        odds: sm.multiplier,
      }))
    }
    return match
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
  async finalizeMatch(id: string): Promise<{ predictions_processed: number; total_points_awarded: number }> {
    const r = await wcApi.post(`/admin/matches/${id}/finalize`)
    return r.data
  },
  async finalizeAllMatches(): Promise<{ processed: number; skipped: number; total_points_awarded: number }> {
    const r = await wcApi.post('/admin/matches/finalize-all')
    return r.data
  },
  async refinalizeAllMatches(): Promise<{ processed: number; skipped: number; total_points_awarded: number }> {
    const r = await wcApi.post('/admin/matches/refinalize-all')
    return r.data
  },

  // --- Score Multipliers ---
  async getScoreMultipliers(matchId: string): Promise<WcScoreMultiplier[]> {
    const r = await wcApi.get<WcScoreMultiplier[]>(`/matches/${matchId}/score-multipliers`)
    return r.data
  },
  async addScoreMultiplier(matchId: string, homeScore: number, awayScore: number, multiplier: number): Promise<WcScoreMultiplier> {
    const r = await wcApi.post<WcScoreMultiplier>(`/admin/matches/${matchId}/score-multipliers`, {
      home_score: homeScore,
      away_score: awayScore,
      multiplier,
    })
    return r.data
  },
  async updateScoreMultiplier(id: string, multiplier: number): Promise<void> {
    await wcApi.put(`/admin/score-multipliers/${id}`, { multiplier })
  },
  async deleteScoreMultiplier(id: string): Promise<void> {
    await wcApi.delete(`/admin/score-multipliers/${id}`)
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

  // --- Predictions ---
  async submitPrediction(req: WcSubmitPredictionRequest): Promise<WcPrediction> {
    const r = await wcApi.post<WcPrediction>('/predictions', req)
    return r.data
  },
  async listPredictions(): Promise<WcPredictionWithMatch[]> {
    const r = await wcApi.get<WcPredictionWithMatch[]>('/predictions')
    return r.data
  },
  async deletePrediction(id: string): Promise<void> {
    await wcApi.delete(`/predictions/${id}`)
  },
  async updatePredictionPoints(id: string, points: number): Promise<void> {
    await wcApi.put(`/predictions/${id}`, { points })
  },
  async getMatchPredictions(matchId: string): Promise<WcPredictionPublic[]> {
    const r = await wcApi.get<WcPredictionPublic[]>(`/matches/${matchId}/predictions`)
    return r.data
  },

  // --- Bets ---
  async listBets(): Promise<WcBetWithMatch[]> {
    const r = await wcApi.get<WcBetWithMatch[]>('/bets')
    return r.data
  },
  async getMatchBets(matchId: string): Promise<WcBetPublic[]> {
    const r = await wcApi.get<WcBetPublic[]>(`/matches/${matchId}/bets`)
    return r.data
  },
  async placeBet(req: WcPlaceBetRequest): Promise<WcBet> {
    const r = await wcApi.post<WcBet>('/bets', req)
    return r.data
  },
  async updateBetStake(id: string, stake: number): Promise<void> {
    await wcApi.put(`/bets/${id}`, { stake })
  },
  async deleteBet(id: string): Promise<void> {
    await wcApi.delete(`/bets/${id}`)
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
  async getHousePnL(): Promise<HousePnLResponse> {
    const r = await wcApi.get<HousePnLResponse>('/admin/house-pnl')
    return r.data
  },

  // --- StatsAPI Sync ---
  async setupMapping(previewOnly: boolean): Promise<MappingResult> {
    const r = await wcApi.post<MappingResult>('/admin/setup-statsapi-mapping', { preview_only: previewOnly })
    return r.data
  },
  async importHandicap(matchId: string, previewOnly: boolean): Promise<ImportHandicapPreview | { ok: boolean }> {
    const r = await wcApi.post<ImportHandicapPreview | { ok: boolean }>(`/admin/matches/${matchId}/import-handicap`, { preview_only: previewOnly })
    return r.data
  },
  async importOU(matchId: string, previewOnly: boolean): Promise<ImportOUPreview | { ok: boolean }> {
    const r = await wcApi.post<ImportOUPreview | { ok: boolean }>(`/admin/matches/${matchId}/import-ou`, { preview_only: previewOnly })
    return r.data
  },
  async generatePoisson(matchId: string, params: { home_lambda: number; away_lambda: number; house_margin?: number; min_prob?: number }, previewOnly: boolean): Promise<GeneratePoissonPreview | { ok: boolean; count: number }> {
    const r = await wcApi.post<GeneratePoissonPreview | { ok: boolean; count: number }>(`/admin/matches/${matchId}/generate-poisson`, { ...params, preview_only: previewOnly })
    return r.data
  },
  async getSyncLogs(): Promise<WcSyncLog[]> {
    const r = await wcApi.get<WcSyncLog[]>('/admin/sync-logs')
    return r.data
  },

  // --- Champion Prediction ---
  async getChampionConfig(): Promise<WcChampionConfig> {
    const r = await wcApi.get<WcChampionConfig>('/champion/config')
    return r.data
  },
  async getChampionTeams(): Promise<WcChampionTeam[]> {
    const r = await wcApi.get<WcChampionTeam[]>('/champion/teams')
    return r.data
  },
  async getChampionPredictions(): Promise<WcChampionPredictionPublic[]> {
    const r = await wcApi.get<WcChampionPredictionPublic[]>('/champion/predictions')
    return r.data
  },
  async getMyChampionPrediction(): Promise<WcChampionPredictionMine | null> {
    const r = await wcApi.get<WcChampionPredictionMine | null>('/champion/my-prediction')
    return r.data ?? null
  },
  async placeChampionPrediction(teamId: string, points: number): Promise<WcChampionPredictionMine> {
    const r = await wcApi.post<WcChampionPredictionMine>('/champion/predict', { team_id: teamId, points })
    return r.data
  },
  async deleteChampionPrediction(): Promise<void> {
    await wcApi.delete('/champion/predict')
  },
  async adminUpdateChampionConfig(isOpen: boolean): Promise<void> {
    await wcApi.put('/admin/champion/config', { is_open: isOpen })
  },
  async adminCreateChampionTeam(data: { name: string; code: string; flag_emoji: string; odds: number }): Promise<WcChampionTeam> {
    const r = await wcApi.post<WcChampionTeam>('/admin/champion/teams', data)
    return r.data
  },
  async adminUpdateChampionTeamOdds(teamId: string, odds: number): Promise<void> {
    await wcApi.put(`/admin/champion/teams/${teamId}`, { odds })
  },
  async adminSettleChampion(winnerTeamId: string): Promise<WcChampionSettleResult> {
    const r = await wcApi.post<WcChampionSettleResult>('/admin/champion/settle', { winner_team_id: winnerTeamId })
    return r.data
  },
}
