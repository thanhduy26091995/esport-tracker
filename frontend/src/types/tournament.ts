import type { User } from './user'
import type { MatchType } from './match'

export type TournamentStatus = 'active' | 'completed'

export interface TournamentParticipant {
  id: string
  tournament_id: string
  user_id: string
  tier_snapshot: string
  handicap_rate_snapshot: number
  user: User
}

export interface TournamentMatch {
  id: string
  tournament_id: string
  round: number
  match_order: number
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
  status: TournamentStatus
  affects_score: boolean
  entry_fee: number
  created_at: string
  updated_at: string
  participants: TournamentParticipant[]
  matches: TournamentMatch[]
}

export interface CreateTournamentRequest {
  name: string
  match_type: MatchType
  player_ids: string[]
  affects_score: boolean
  entry_fee: number
}

export interface RecordMatchResultRequest {
  actual_score1: number
  actual_score2: number
  recorded_by: string
}

export interface TournamentStanding {
  user: User
  wins: number
  draws: number
  losses: number
  goals_for: number
  goals_against: number
  points: number // 3 per win, 1 per draw
}
