import { api } from './api'
import type {
  Match,
  CreateMatchRequest,
  MatchStats,
} from '@/types/match'
import type { PaginationParams } from '@/types/api'

export const matchService = {
  async getAll(params?: PaginationParams): Promise<Match[]> {
    const response = await api.get<Match[]>('/matches', { params })
    return response.data
  },

  async getById(id: string): Promise<Match> {
    const response = await api.get<Match>(`/matches/${id}`)
    return response.data
  },

  async create(data: CreateMatchRequest): Promise<Match> {
    const response = await api.post<Match>('/matches', data)
    return response.data
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/matches/${id}`)
  },

  async getByUser(userId: string, params?: PaginationParams): Promise<Match[]> {
    const response = await api.get<Match[]>(`/matches/user/${userId}`, { params })
    return response.data
  },

  async getStats(): Promise<MatchStats> {
    const response = await api.get<MatchStats>('/matches/stats')
    return response.data
  },
}
