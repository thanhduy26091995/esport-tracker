import { wcApi } from './wcApi'
import type { WcLoginResponse } from '@/types/wc'

export const wcAuthService = {
  async login(name: string, password: string): Promise<WcLoginResponse> {
    const response = await wcApi.post<WcLoginResponse>('/auth/login', { name, password })
    return response.data
  },

  async googleLogin(idToken: string): Promise<WcLoginResponse> {
    const response = await wcApi.post<WcLoginResponse>('/auth/google', { id_token: idToken })
    return response.data
  },

  async googleLink(idToken: string): Promise<{ google_linked: boolean; avatar_url: string | null }> {
    const response = await wcApi.post<{ google_linked: boolean; avatar_url: string | null }>(
      '/auth/google/link',
      { id_token: idToken }
    )
    return response.data
  },

}
