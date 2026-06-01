import { api } from './api'
import type { WcLoginResponse } from '@/types/wc'

export const wcAuthService = {
  async register(name: string, password: string): Promise<WcLoginResponse> {
    const response = await api.post<WcLoginResponse>('/wc/auth/register', { name, password })
    return response.data
  },

  async login(name: string, password: string): Promise<WcLoginResponse> {
    const response = await api.post<WcLoginResponse>('/wc/auth/login', { name, password })
    return response.data
  },

  async resetPassword(name: string): Promise<void> {
    await api.post('/wc/auth/reset-password', { name })
  },
}
