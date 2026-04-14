import { api } from './api'
import type { User, CreateUserRequest, UpdateUserRequest } from '@/types/user'

export const userService = {
  async getAll(): Promise<User[]> {
    const response = await api.get<User[]>('/users')
    return response.data
  },

  async getById(id: string): Promise<User> {
    const response = await api.get<User>(`/users/${id}`)
    return response.data
  },

  async create(data: CreateUserRequest): Promise<User> {
    const response = await api.post<User>('/users', data)
    return response.data
  },

  async update(id: string, data: UpdateUserRequest): Promise<User> {
    const response = await api.put<User>(`/users/${id}`, data)
    return response.data
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/users/${id}`)
  },

  async getLeaderboard(limit?: number): Promise<User[]> {
    const response = await api.get<User[]>('/users/leaderboard', {
      params: { limit },
    })
    return response.data
  },

  async getStats(): Promise<{ total: number; active: number }> {
    const response = await api.get<{ total: number; active: number }>('/users/stats')
    return response.data
  },
}
