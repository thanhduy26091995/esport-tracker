import { api } from './api'
import type { Config, ConfigKey, UpdateConfigRequest } from '@/types/config'

export const configService = {
  async getAll(): Promise<Config[]> {
    const response = await api.get<Config[]>('/config')
    return response.data
  },

  async getByKey(key: ConfigKey): Promise<Config> {
    const response = await api.get<Config>(`/config/${key}`)
    return response.data
  },

  async update(key: ConfigKey, data: UpdateConfigRequest): Promise<Config> {
    const response = await api.put<Config>(`/config/${key}`, data)
    return response.data
  },
}
