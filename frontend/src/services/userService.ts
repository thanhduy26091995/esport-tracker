import { api } from './api'
import type { UserWithStats, UserWithPaymentTotal, CreateUserRequest, UpdateUserRequest } from '@/types/user'

export const userService = {
  async getAll(): Promise<UserWithStats[]> {
    const response = await api.get<UserWithStats[]>('/users')
    return response.data
  },

  async getById(id: string): Promise<UserWithStats> {
    const response = await api.get<UserWithStats>(`/users/${id}`)
    return response.data
  },

  async create(data: CreateUserRequest): Promise<UserWithStats> {
    const response = await api.post<UserWithStats>('/users', data)
    return response.data
  },

  async update(id: string, data: UpdateUserRequest): Promise<UserWithStats> {
    const response = await api.put<UserWithStats>(`/users/${id}`, data)
    return response.data
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/users/${id}`)
  },

  async getLeaderboard(limit?: number): Promise<UserWithStats[]> {
    const response = await api.get<UserWithStats[]>('/users/leaderboard', {
      params: { limit },
    })
    return response.data
  },

  async getStats(): Promise<{ total: number; active: number }> {
    const response = await api.get<{ total: number; active: number }>('/users/stats')
    return response.data
  },

  async getPaymentRanking(): Promise<UserWithPaymentTotal[]> {
    const response = await api.get<UserWithPaymentTotal[]>('/users/payment-ranking')
    return response.data
  },
}
