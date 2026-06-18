import type { User } from './user'
import type { MatchType } from './match'

export type TournamentStatus = 'active' | 'completed'
export type TournamentFormat = 'classic' | 'round_robin_top4'
export type TournamentMatchStage = 'group' | 'semi' | 'final' | 'third_place'

export interface TournamentParticipant {
  id: string
  tournament_id: string
  user_id: string
  tier_snapshot: string
  handicap_rate_snapshot: number
  user: User
}

export interface TournamentTeam {
  id: string
  tournament_id: string
  player1_id: string
  player2_id: string
  player1_handicap_snapshot: number
  player2_handicap_snapshot: number
  player1: User
  player2: User
}

export interface TournamentMatch {
  id: string
  tournament_id: string
  round: number
  match_order: number
  stage: TournamentMatchStage
  team1_team_id?: string
  team2_team_id?: string
  team1_player1_id: string
  team1_player2_id?: string
  team2_player1_id: string
  team2_player2_id?: string
  handicap_team1: number
  handicap_team2: number
  status: 'pending' | 'completed'
  actual_score1?: number
  actual_score2?: number
  effective_winner: number // 0=draw, 1=team1, 2=team2
  match_id?: string
}

export interface Tournament {
  id: string
  name: string
  match_type: MatchType
  format: TournamentFormat
  status: TournamentStatus
  affects_score: boolean
  entry_fee: number
  knockout_size: number // 2 = final only; 4 = semis+final+3rd
  champion_team_id?: string
  champion_team?: TournamentTeam
  created_at: string
  updated_at: string
  participants: TournamentParticipant[]
  teams: TournamentTeam[]
  matches: TournamentMatch[]
  standings?: TeamStanding[]
}

export interface TeamInputEntry {
  player1_id: string
  player2_id: string
}

export interface CreateTournamentRequest {
  name: string
  match_type?: MatchType
  format?: TournamentFormat
  player_ids?: string[]
  teams?: TeamInputEntry[]
  affects_score: boolean
  entry_fee: number
  knockout_size?: number // 2 or 4; only relevant for round_robin_top4
}

export interface RecordMatchResultRequest {
  actual_score1: number
  actual_score2: number
  recorded_by: string
}

// Per-player standing (classic format, computed on frontend)
export interface TournamentStanding {
  user: User
  wins: number
  draws: number
  losses: number
  goals_for: number
  goals_against: number
  points: number
}

// Per-team standing (round_robin_top4, computed on backend)
export interface TeamStanding {
  team_id: string
  player1: User
  player2: User
  played: number
  won: number
  drawn: number
  lost: number
  gf: number
  ga: number
  gd: number
  points: number
  seed: number // 1-4 = qualified; 0 = eliminated
}
