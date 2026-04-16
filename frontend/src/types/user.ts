export interface User {
  id: string
  name: string
  current_score: number
  created_at: string
  updated_at: string
  is_active: boolean
  tier: string
  handicap_rate: number
}

export interface CreateUserRequest {
  name: string
  tier?: string
  handicap_rate?: number
}

export interface UpdateUserRequest {
  name: string
  tier?: string
  handicap_rate?: number
}
