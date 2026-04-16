import { api } from './api'
import type {
  Tournament,
  CreateTournamentRequest,
  RecordMatchResultRequest,
} from '@/types/tournament'

export const tournamentService = {
  async getAll(): Promise<Tournament[]> {
    const response = await api.get<Tournament[]>('/tournaments')
    return response.data
  },

  async getById(id: string): Promise<Tournament> {
    const response = await api.get<Tournament>(`/tournaments/${id}`)
    return response.data
  },

  async create(data: CreateTournamentRequest): Promise<Tournament> {
    const response = await api.post<Tournament>('/tournaments', data)
    return response.data
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/tournaments/${id}`)
  },

  async complete(id: string): Promise<Tournament> {
    const response = await api.put<Tournament>(`/tournaments/${id}/complete`)
    return response.data
  },

  async recordResult(
    tournamentId: string,
    matchId: string,
    data: RecordMatchResultRequest
  ): Promise<any> {
    const response = await api.post(
      `/tournaments/${tournamentId}/matches/${matchId}/result`,
      data
    )
    return response.data
  },
}
