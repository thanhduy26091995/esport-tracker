// Head-to-head API response types (snake_case, mirroring the backend JSON).

export interface H2HPlayer {
  id: string
  name: string
  avatar_url?: string | null
  favorite_club?: string | null
  tier: string
  is_active: boolean
}

export interface H2HParticipant {
  user_id: string
  name: string
  avatar_url?: string | null
  team: number // 1 or 2
}

export interface H2HMatch {
  match_id: string
  match_type: string
  match_date: string
  winner_team: number
  player1_team: number
  player1_won: boolean
  participants: H2HParticipant[]
}

export interface H2HStreak {
  player_id?: string | null // whose current streak (null when no matches)
  count: number
}

export interface HeadToHeadResponse {
  player1: H2HPlayer
  player2: H2HPlayer
  total_matches: number
  player1_wins: number
  player2_wins: number
  player1_win_rate: number // 0..1
  player2_win_rate: number
  current_streak: H2HStreak
  form: string[] // "W"/"L", most-recent first, max 10
  recent_matches: H2HMatch[]
}
