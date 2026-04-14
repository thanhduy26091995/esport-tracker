import { api } from './api'
import type {
  DebtSettlement,
  TriggerSettlementRequest,
  SettlementStats,
} from '@/types/settlement'
import type { PaginationParams } from '@/types/api'

export const settlementService = {
  async getAll(params?: PaginationParams): Promise<DebtSettlement[]> {
    const response = await api.get<DebtSettlement[]>('/settlements', { params })
    return response.data
  },

  async getById(id: string): Promise<DebtSettlement> {
    const response = await api.get<DebtSettlement>(`/settlements/${id}`)
    return response.data
  },

  async getByDebtorId(debtorId: string, params?: PaginationParams): Promise<DebtSettlement[]> {
    const response = await api.get<DebtSettlement[]>(`/settlements/debtor/${debtorId}`, {
      params,
    })
    return response.data
  },

  async trigger(data: TriggerSettlementRequest): Promise<DebtSettlement> {
    const response = await api.post<DebtSettlement>('/settlements/trigger', data)
    return response.data
  },

  async getStats(): Promise<SettlementStats> {
    const response = await api.get<SettlementStats>('/settlements/stats')
    return response.data
  },
}
