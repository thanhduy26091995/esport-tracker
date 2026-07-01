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
  FinalizePreviewResult,
  WcCustomBetWithOptions,
  WcCustomBetEntryHistory,
  CreateCustomBetOption,
  BetHistoryItem,
  ReduceStakePreview,
} from '@/types/wc'

export const wcService = {
  // --- Config ---
  async getPublicConfig(): Promise<Pick<WcConfig, 'is_enabled' | 'min_points' | 'max_points' | 'cancel_penalty_enabled' | 'cancel_penalty_percent' | 'bet_reduce_max_percent' | 'bet_reduce_penalty_percent'>> {
    const r = await wcApi.get('/config')
    return r.data
  },
  async getConfig(): Promise<WcConfig> {
    const r = await wcApi.get<WcConfig>('/admin/config')
    return r.data
  },
  async updateConfig(isEnabled: boolean): Promise<void> {
    await wcApi.put('/admin/config', { is_enabled: isEnabled })
  },
  async updateBetLimits(minPoints: number, maxPoints: number): Promise<WcConfig> {
    const r = await wcApi.put<WcConfig>('/admin/config', { min_points: minPoints, max_points: maxPoints })
    return r.data
  },
  async updatePenaltyConfig(params: {
    cancel_penalty_enabled?: boolean
    cancel_penalty_percent?: number
    bet_reduce_max_percent?: number
    bet_reduce_penalty_percent?: number
  }): Promise<WcConfig> {
    const r = await wcApi.put<WcConfig>('/admin/config', params)
    return r.data
  },

  async backfillOriginalPoints(): Promise<{ ok: boolean; rows_updated: number }> {
    const r = await wcApi.post<{ ok: boolean; rows_updated: number }>('/admin/backfill-original-points')
    return r.data
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
  async finalizeMatch(id: string): Promise<{ predictions_processed: number; total_points_awarded: number; unsettled_custom_bets_count: number }> {
    const r = await wcApi.post(`/admin/matches/${id}/finalize`)
    return r.data
  },
  async finalizeAllMatches(): Promise<{ processed: number; skipped: number; total_points_awarded: number; matches_with_unsettled_custom_bets: number }> {
    const r = await wcApi.post('/admin/matches/finalize-all')
    return r.data
  },
  async refinalizeAllMatches(): Promise<{ processed: number; skipped: number; total_points_awarded: number }> {
    const r = await wcApi.post('/admin/matches/refinalize-all')
    return r.data
  },
  async previewFinalizeMatch(matchId: string): Promise<FinalizePreviewResult> {
    const r = await wcApi.get<FinalizePreviewResult>(`/admin/matches/${matchId}/finalize-preview`)
    return r.data
  },
  async previewFinalizeAll(): Promise<FinalizePreviewResult> {
    const r = await wcApi.get<FinalizePreviewResult>('/admin/matches/finalize-all-preview')
    return r.data
  },
  async previewRefinalizeAll(): Promise<FinalizePreviewResult> {
    const r = await wcApi.get<FinalizePreviewResult>('/admin/matches/refinalize-all-preview')
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
  async previewReducePredictionPoints(predId: string, newPoints: number): Promise<ReduceStakePreview> {
    const r = await wcApi.get<ReduceStakePreview>(`/predictions/${predId}/reduce-preview`, { params: { new_points: newPoints } })
    return r.data
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
  async getBetHistory(): Promise<BetHistoryItem[]> {
    const r = await wcApi.get<BetHistoryItem[]>('/bets/history')
    return r.data
  },
  async previewReduceStake(betId: string, newStake: number): Promise<ReduceStakePreview> {
    const r = await wcApi.get<ReduceStakePreview>(`/bets/${betId}/reduce-preview`, { params: { new_stake: newStake } })
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
  async blockUser(userId: string): Promise<{ ok: boolean; voided_bets: number }> {
    const r = await wcApi.put<{ ok: boolean; voided_bets: number }>(`/admin/users/${userId}/block`)
    return r.data
  },
  async unblockUser(userId: string): Promise<void> {
    await wcApi.put(`/admin/users/${userId}/unblock`)
  },
  async setUserBot(userId: string, isBot: boolean): Promise<void> {
    await wcApi.put(`/admin/users/${userId}/bot`, { is_bot: isBot })
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
  async getMyChampionPredictions(): Promise<WcChampionPredictionMine[]> {
    const r = await wcApi.get<WcChampionPredictionMine[]>('/champion/my-prediction')
    return r.data ?? []
  },
  async placeChampionPrediction(teamId: string, points: number): Promise<WcChampionPredictionMine> {
    const r = await wcApi.post<WcChampionPredictionMine>('/champion/predict', { team_id: teamId, points })
    return r.data
  },
  async deleteChampionPrediction(predId: string): Promise<void> {
    await wcApi.delete(`/champion/predict/${predId}`)
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

  // --- Custom Bets (Kèo phụ) ---
  async adminListCustomBets(matchId: string): Promise<WcCustomBetWithOptions[]> {
    const r = await wcApi.get<WcCustomBetWithOptions[]>(`/admin/matches/${matchId}/custom-bets`)
    return r.data ?? []
  },
  async adminCreateCustomBet(matchId: string, title: string, line: number | null, options: CreateCustomBetOption[]): Promise<WcCustomBetWithOptions> {
    const r = await wcApi.post<WcCustomBetWithOptions>(`/admin/matches/${matchId}/custom-bets`, { title, line: line ?? undefined, options })
    return r.data
  },
  async adminUpdateCustomBet(betId: string, data: { title?: string; line?: number | null; status?: string }): Promise<void> {
    await wcApi.put(`/admin/custom-bets/${betId}`, data)
  },
  async adminSettleCustomBet(betId: string, winningOptionId: string): Promise<void> {
    await wcApi.post(`/admin/custom-bets/${betId}/settle`, { winning_option_id: winningOptionId })
  },
  async adminVoidCustomBet(betId: string): Promise<void> {
    await wcApi.put(`/admin/custom-bets/${betId}/void`)
  },
  async listCustomBets(matchId: string): Promise<WcCustomBetWithOptions[]> {
    const r = await wcApi.get<WcCustomBetWithOptions[]>(`/matches/${matchId}/custom-bets`)
    return r.data ?? []
  },
  async placeCustomBetEntry(betId: string, optionId: string, stake: number): Promise<void> {
    await wcApi.post(`/custom-bets/${betId}/entry`, { option_id: optionId, stake })
  },
  async cancelCustomBetEntry(entryId: string): Promise<void> {
    await wcApi.delete(`/custom-bet-entries/${entryId}`)
  },
  async getMyCustomBetEntries(): Promise<WcCustomBetEntryHistory[]> {
    const r = await wcApi.get<WcCustomBetEntryHistory[]>('/custom-bet-entries')
    return r.data ?? []
  },
  async getWcUsersForMention(): Promise<{ users: { id: string; name: string; avatar_url: string | null }[] }> {
    const r = await wcApi.get<{ users: { id: string; name: string; avatar_url: string | null }[] }>('/users')
    return r.data
  },
  async getUnreadMentionCount(): Promise<{ count: number }> {
    const r = await wcApi.get<{ count: number }>('/chat/mentions/unread-count')
    return r.data
  },
  async markMentionsRead(): Promise<void> {
    await wcApi.post('/chat/mentions/read')
  },
}
