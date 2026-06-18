import { wcApi } from './wcApi'
import type { WcProfile } from '@/types/wc'

export const wcProfileService = {
  async getProfile(): Promise<WcProfile> {
    const response = await wcApi.get<WcProfile>('/profile')
    return response.data
  },

  async updateProfile(data: { name?: string; avatar_url?: string }): Promise<WcProfile> {
    const response = await wcApi.put<WcProfile>('/profile', data)
    return response.data
  },
}
