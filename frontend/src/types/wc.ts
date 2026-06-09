export type WcMatchStatus = 'scheduled' | 'live' | 'completed' | 'cancelled'
export type WcStage = 'group' | 'r32' | 'r16' | 'qf' | 'sf' | 'final' | 'third_place'
export type WcPredictionType = 'handicap' | 'exact_score'
export type WcPredictionResult = 'correct' | 'incorrect' | 'void'
export type WcSettlementDirection = 'pay' | 'collect' | 'even'
export type WcSettlementStatus = 'pending' | 'done'
export type WcBetType = 'handicap' | 'exact_score'
export type WcBetResult = 'win' | 'lose' | 'push'

export interface WcUser {
  id: string
  name: string
  is_admin: boolean
  created_at: string
  updated_at: string
}

export interface WcConfig {
  id: number
  is_enabled: boolean
  updated_at: string
  updated_by?: string
}

export interface WcMatch {
  id: string
  external_id?: number
  home_team: string
  away_team: string
  home_team_code?: string
  away_team_code?: string
  match_date: string
  status: WcMatchStatus
  stage: WcStage
  group_name?: string
  home_score?: number
  away_score?: number
  handicap_value?: number
  handicap_team?: string
  odds_handicap_home?: number
  odds_handicap_away?: number
  predictions_open: boolean
  predictions_locked_at?: string
  bets_locked_at?: string
  settled_at?: string
  created_at: string
  updated_at: string
}

export interface WcScoreMultiplier {
  id: string
  match_id: string
  home_score: number
  away_score: number
  multiplier: number
  created_at: string
  updated_at: string
}

export interface WcScoreOdds {
  id: string
  home_score: number
  away_score: number
  odds: number
}

export interface WcMatchWithOdds extends WcMatch {
  score_multipliers: WcScoreMultiplier[]
  score_odds?: WcScoreOdds[]
}

export interface WcWallet {
  id: string
  wc_user_id: string
  balance: number
  updated_at: string
}

export interface WcWalletWithUser extends WcWallet {
  user_name: string
}

export interface WcWalletLog {
  id: string
  wc_user_id: string
  admin_id?: string
  delta: number
  balance_before: number
  balance_after: number
  note?: string
  created_at: string
}

export interface WcPrediction {
  id: string
  wc_user_id: string
  match_id: string
  prediction_type: WcPredictionType
  prediction_choice?: string
  points: number
  multiplier_snapshot: number
  handicap_snapshot?: number
  handicap_team_snapshot?: string
  predicted_home_score?: number
  predicted_away_score?: number
  result?: WcPredictionResult
  points_earned?: number
  created_at: string
  updated_at: string
}

export interface WcPredictionWithMatch extends WcPrediction {
  home_team: string
  away_team: string
  match_date: string
  match_status: WcMatchStatus
  predictions_open: boolean
  predictions_locked_at?: string
}

export interface WcPredictionPublic {
  id: string
  wc_user_id: string
  name: string
  prediction_type: WcPredictionType
  prediction_choice?: string
  points: number
  multiplier_snapshot: number
  predicted_home_score?: number
  predicted_away_score?: number
  result?: WcPredictionResult
  points_earned?: number
  created_at: string
}

export interface WcLeaderboardEntry {
  wc_user_id: string
  name: string
  net_points: number
  total_predictions: number
  correct: number
}

export interface WcSettlement {
  id: string
  name: string
  admin_id: string
  point_rate: number
  note?: string
  created_at: string
}

export interface WcSettlementDetail {
  id: string
  settlement_id: string
  wc_user_id: string
  balance_snapshot: number
  direction: WcSettlementDirection
  amount: number
  status: WcSettlementStatus
  done_note?: string
  updated_at: string
}

export interface WcSettlementDetailWithUser extends WcSettlementDetail {
  user_name: string
}

export interface WcSettlementWithDetails extends WcSettlement {
  details: WcSettlementDetailWithUser[]
}

export interface WcSettlementPreviewRow {
  wc_user_id: string
  user_name: string
  balance: number
  direction: WcSettlementDirection
  amount: number
}

export interface WcBet {
  id: string
  wc_user_id: string
  match_id: string
  bet_type: WcBetType
  bet_choice?: string
  stake: number
  odds_snapshot: number
  predicted_home_score?: number
  predicted_away_score?: number
  result?: WcBetResult
  payout?: number
  created_at: string
  updated_at: string
}

export interface WcBetWithMatch extends WcBet {
  home_team: string
  away_team: string
  match_date: string
  match_status: WcMatchStatus
  betting_open: boolean  // computed server-side from bets_locked_at
  bets_locked_at?: string
}

export interface WcBetPublic {
  id: string
  wc_user_id: string
  name: string
  bet_type: WcBetType
  bet_choice?: string
  stake: number
  odds_snapshot: number
  predicted_home_score?: number
  predicted_away_score?: number
  result?: WcBetResult
  payout?: number
  created_at: string
}

export interface WcPlaceBetRequest {
  match_id: string
  bet_type: WcBetType
  bet_choice?: string
  stake: number
  predicted_home_score?: number
  predicted_away_score?: number
}

export interface WcAuthUser {
  id: string
  name: string
  isAdmin: boolean
}

export interface WcLoginResponse {
  token: string
  user_id: string
  name: string
  is_admin: boolean
}

export interface WcSubmitPredictionRequest {
  match_id: string
  prediction_type: WcPredictionType
  prediction_choice?: string
  predicted_home_score?: number
  predicted_away_score?: number
  points: number
}

export interface WcMatchFilter {
  status?: WcMatchStatus
  stage?: WcStage
  group?: string
  date?: string
}
