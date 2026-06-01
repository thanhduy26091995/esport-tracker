import { wcApi } from './wcApi'
import type { WcLoginResponse } from '@/types/wc'

export const wcAuthService = {
  async register(name: string, password: string): Promise<WcLoginResponse> {
    const response = await wcApi.post<WcLoginResponse>('/auth/register', { name, password })
    return response.data
  },

  async login(name: string, password: string): Promise<WcLoginResponse> {
    const response = await wcApi.post<WcLoginResponse>('/auth/login', { name, password })
    return response.data
  },

  async resetPassword(name: string): Promise<void> {
    await wcApi.post('/auth/reset-password', { name })
  },
}
