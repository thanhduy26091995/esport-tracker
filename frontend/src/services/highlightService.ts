import { api } from './api'
import type { HighlightsResponse } from '@/types/highlight'

export const highlightService = {
  async getHighlights(): Promise<HighlightsResponse> {
    const response = await api.get<HighlightsResponse>('/highlights')
    return response.data
  },
}
