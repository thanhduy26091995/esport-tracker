export interface User {
  id: string
  name: string
  current_score: number
  created_at: string
  updated_at: string
  is_active: boolean
}

export interface CreateUserRequest {
  name: string
}

export interface UpdateUserRequest {
  name: string
}
