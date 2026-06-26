export type WcMatchStatus = 'scheduled' | 'live' | 'completed' | 'cancelled'
export type WcStage = 'group' | 'r32' | 'r16' | 'qf' | 'sf' | 'final' | 'third_place'
export type WcPredictionType = 'handicap' | 'exact_score' | 'over_under'
export type WcPredictionResult = 'correct' | 'incorrect' | 'void' | 'win_half' | 'lose_half'
export type WcSettlementDirection = 'pay' | 'collect' | 'even'
export type WcSettlementStatus = 'pending' | 'done'
export type WcBetType = 'handicap' | 'exact_score'
export type WcBetResult = 'win' | 'lose' | 'push' | 'win_half' | 'lose_half'

export interface WcUser {
  id: string
  name: string
  is_admin: boolean
  is_blocked: boolean
  created_at: string
  updated_at: string
}

export interface WcConfig {
  id: number
  is_enabled: boolean
  min_points: number
  max_points: number
  updated_at: string
  updated_by?: string
}

export interface WcMatch {
  id: string
  external_id?: number
  statsapi_fixture_id?: string
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
  odds_synced_at?: string
  ou_line?: number
  odds_over?: number
  odds_under?: number
  ou_synced_at?: string
  poisson_synced_at?: string
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
  prediction_type: WcPredictionType | 'custom'
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
  bet_title?: string
}

export interface WcPredictionPublic {
  id: string
  wc_user_id: string
  name: string
  avatar_url: string | null
  prediction_type: WcPredictionType | 'custom'
  prediction_choice?: string
  points: number
  multiplier_snapshot: number
  predicted_home_score?: number
  predicted_away_score?: number
  result?: WcPredictionResult
  points_earned?: number
  created_at: string
  bet_title?: string
}

export interface WcLeaderboardEntry {
  wc_user_id: string
  name: string
  avatar_url: string | null
  net_points: number
  total_predictions: number
  correct: number
  win_half: number
  lose_half: number
  incorrect: number
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
  avatar_url: string | null
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
  avatarUrl: string | null
  googleLinked: boolean
}

export interface WcLoginResponse {
  token: string
  user_id: string
  name: string
  is_admin: boolean
  avatar_url: string | null
  google_linked: boolean
}

export interface WcProfile {
  id: string
  name: string
  avatar_url: string | null
  is_admin: boolean
  google_linked: boolean
  created_at: string
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
  date_from?: string // ISO8601 UTC — match_date >= date_from
  date_to?: string   // ISO8601 UTC — match_date <= date_to
}

export interface WcSyncLog {
  id: string
  trigger: string
  sync_type: string
  triggered_by?: string
  matches_updated: number
  matches_failed: number
  error_detail?: string
  created_at: string
}

export interface StatsApiFixtureRef {
  id: string
  home_team: string
  away_team: string
  match_date: string
}

// --- Analytics ---

export interface AnalyticsTimelinePoint {
  period: string
  wins: number
  losses: number
  accuracy: number
}

export interface AnalyticsCompareMetrics {
  home_bias: number | null
  avg_goals_predicted: number | null
  exact_score_rate: number
  underdog_rate: number | null
  avg_stake: number
  over_preference_rate: number | null
  exact_score_hit_rate: number | null
  bet_frequency: number
  last_minute_rate: number
}

export interface TeamCountEntry {
  team: string
  bet_count: number
}

export interface ScorelineCountEntry {
  scoreline: string
  count: number
}

export interface TopPredictorEntry {
  user_id: string
  name: string
  avatar_url: string | null
  accuracy: number
  settled_matches: number
}

export interface MyAnalyticsResponse {
  accuracy: number
  settled_matches: number
  wins: number
  losses: number
  pending_bets: number
  profile_label: string | null
  current_win_streak: number
  current_lose_streak: number
  longest_win_streak: number
  bet_type_distribution: { handicap: number; exact_score: number; over_under: number; custom: number }
  favorite_teams: TeamCountEntry[]
  favorite_scorelines: ScorelineCountEntry[]
  accuracy_timeline: AnalyticsTimelinePoint[]
  compare_metrics: AnalyticsCompareMetrics
}

export interface CommunityAnalyticsResponse {
  total_bets_placed: number
  active_users: number
  avg_accuracy: number
  prediction_distribution: { home: number; away: number; other: number }
  trending_teams: TeamCountEntry[]
  trending_scorelines: ScorelineCountEntry[]
  community_compare_metrics: AnalyticsCompareMetrics
  top_predictors: TopPredictorEntry[]
}

export interface CompareAnalyticsResponse {
  me: AnalyticsCompareMetrics
  community: AnalyticsCompareMetrics
  my_accuracy: number
  community_accuracy: number
}

export interface MappedMatch {
  wc_match_id: string
  home_team: string
  away_team: string
  statsapi_fixture_id: string
  confidence: string
}

export interface MappingResult {
  matched: MappedMatch[]
  unmatched_local: Array<{ id: string; home_team: string; away_team: string }>
  unmatched_api: StatsApiFixtureRef[]
  total_api_fixtures: number
}

export interface HandicapOddsSnapshot {
  handicap_team?: string
  handicap_value?: number
  odds_handicap_home?: number
  odds_handicap_away?: number
}

export interface OUOddsSnapshot {
  ou_line?: number
  odds_over?: number
  odds_under?: number
}

export interface ImportHandicapPreview {
  match_id: string
  statsapi_fixture_id: string
  current: HandicapOddsSnapshot
  proposed: HandicapOddsSnapshot
  source: string
  fetched_at: string
}

export interface ImportOUPreview {
  match_id: string
  current: OUOddsSnapshot
  proposed: OUOddsSnapshot
  source: string
  fetched_at: string
}

export interface PoissonScoreline {
  home_score: number
  away_score: number
  probability: number
  odds: number
}

export interface GeneratePoissonPreview {
  match_id: string
  score_odds: PoissonScoreline[]
  count: number
  house_margin: number
}

export interface HousePnLMatch {
  match_id: string
  home_team: string
  away_team: string
  match_date: string
  stage: string
  stake: number
  payout: number
  profit: number
  bet_count: number
}

export interface HousePnLResponse {
  total_stake_settled: number
  total_payout_settled: number
  house_profit: number
  total_stake_void: number
  total_stake_pending: number
  pending_bet_count: number
  settled_bet_count: number
  match_breakdown: HousePnLMatch[]
  generated_at: string
}

// --- Champion Prediction ---

export interface WcChampionTeam {
  id: string
  name: string
  code: string
  flag_emoji: string
  odds: number
}

export interface WcChampionConfig {
  is_open: boolean
  settled_at?: string
  winner_team?: WcChampionTeam
}

export interface WcChampionPredictionMine {
  id: string
  team_id: string
  team_name: string
  team_code: string
  flag_emoji: string
  points: number
  odds_snapshot: number
  payout_if_correct: number
  result?: string
  points_earned?: number
  created_at: string
  updated_at: string
}

export interface WcChampionPredictionPublic {
  user_name: string
  wc_user_id: string
  team_name: string
  team_code: string
  flag_emoji: string
  points: number
  odds_snapshot: number
  payout_if_correct: number
  result?: string
}

export interface WcChampionSettleResult {
  winner: string
  settled_count: number
  settled_user_count: number
  correct_count: number
  total_points_awarded: number
}

// --- Finalize Preview ---

export interface FinalizePreviewRow {
  wc_user_id: string
  user_name: string
  prediction_type: string  // handicap | exact_score | over_under
  points: number
  multiplier: number
  new_result: string       // correct | incorrect | void | win_half | lose_half
  new_points_earned: number
  net_delta: number        // new_points_earned - points
}

export interface FinalizePreviewMatch {
  match_id: string
  home_team: string
  away_team: string
  home_score: number
  away_score: number
  stage: string
  already_settled: boolean
  predictions: FinalizePreviewRow[]
}

export interface FinalizePreviewHouse {
  total_staked: number
  total_paid_out: number
  house_net: number
  prediction_count: number
  match_count: number
}

export interface FinalizePreviewResult {
  matches: FinalizePreviewMatch[]
  house_summary: FinalizePreviewHouse
}

export interface WcTeamStanding {
  team_name: string
  team_code: string
  played: number
  won: number
  drawn: number
  lost: number
  goals_for: number
  goals_against: number
  goal_difference: number
  points: number
  form: string[]
}

export interface WcGroupStanding {
  group_name: string
  teams: WcTeamStanding[]
}

export interface WcStandingsResponse {
  groups: WcGroupStanding[]
}

// --- Custom Bets (Kèo phụ) ---

export type WcCustomBetStatus = 'open' | 'closed' | 'settled' | 'void'
export type WcCustomBetEntryStatus = 'pending' | 'won' | 'lost' | 'void'

export interface WcCustomBetOption {
  id: string
  custom_bet_id: string
  label: string
  odds: number
  is_winner: boolean
  display_order: number
}

export interface WcCustomBetEntry {
  id: string
  custom_bet_id: string
  option_id: string
  wc_user_id: string
  stake: number
  odds_snapshot: number
  payout?: number
  status: WcCustomBetEntryStatus
  created_at: string
}

export interface WcCustomBet {
  id: string
  match_id: string
  title: string
  line?: number
  status: WcCustomBetStatus
  created_by?: string
  created_at: string
  settled_at?: string
  settled_by?: string
}

export interface WcCustomBetEntryPublic {
  id: string
  wc_user_id: string
  option_id: string
  option_label: string
  name: string
  avatar_url: string | null
  stake: number
  odds_snapshot: number
  status: WcCustomBetEntryStatus
  payout?: number
  created_at: string
}

export interface WcCustomBetWithOptions extends WcCustomBet {
  options: WcCustomBetOption[]
  my_entry?: WcCustomBetEntry
  entry_count: number
  entries: WcCustomBetEntryPublic[]
}

export interface CreateCustomBetOption {
  label: string
  odds: number
  display_order: number
}

export interface WcCustomBetEntryHistory {
  id: string
  custom_bet_id: string
  option_id: string
  wc_user_id: string
  stake: number
  odds_snapshot: number
  payout?: number
  status: WcCustomBetEntryStatus
  created_at: string
  bet_title: string
  bet_line?: number
  option_label: string
  home_team: string
  away_team: string
  match_date: string
}
