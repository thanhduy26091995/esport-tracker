import { api } from './api'
import type { ScoreBonus, CreateScoreBonusRequest } from '@/types/scoreBonus'

export const scoreBonusService = {
  async create(data: CreateScoreBonusRequest): Promise<ScoreBonus> {
    const response = await api.post<ScoreBonus>('/score-bonuses', data)
    return response.data
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/score-bonuses/${id}`)
  },
}
