import type { User } from './user'

export type MatchType = '1v1' | '2v2'
export type TeamNumber = 1 | 2

export interface Match {
  id: string
  match_type: MatchType
  winner_team: TeamNumber
  match_date: string
  recorded_by: string
  created_at: string
  is_locked: boolean
  participants: MatchParticipant[]
}

export interface MatchParticipant {
  id: string
  match_id: string
  user_id: string
  team_number: TeamNumber
  point_change: number
  user: User
}

export interface CreateMatchRequest {
  match_type: MatchType
  team1: string[] // User IDs
  team2: string[] // User IDs
  winner_team: TeamNumber
  match_date?: string
}

export interface MatchStats {
  total: number
  today: number
}
