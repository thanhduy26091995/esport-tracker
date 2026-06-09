import type { User } from './user'

export interface ScoreBonus {
  id: string
  user_id: string
  points: number
  description: string
  bonus_date: string
  created_at: string
  user?: User
}

export interface CreateScoreBonusRequest {
  user_id: string
  points: number
  description: string
  bonus_date?: string
}
