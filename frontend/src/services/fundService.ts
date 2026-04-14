import { api } from './api'
import type {
  FundTransaction,
  CreateDepositRequest,
  CreateWithdrawalRequest,
  FundBalance,
  FundStats,
} from '@/types/fund'
import type { PaginationParams } from '@/types/api'

export const fundService = {
  async getBalance(): Promise<FundBalance> {
    const response = await api.get<FundBalance>('/fund/balance')
    return response.data
  },

  async getStats(): Promise<FundStats> {
    const response = await api.get<FundStats>('/fund/stats')
    return response.data
  },

  async getTransactions(params?: PaginationParams & { type?: string }): Promise<FundTransaction[]> {
    const response = await api.get<FundTransaction[]>('/fund/transactions', { params })
    return response.data
  },

  async deposit(data: CreateDepositRequest): Promise<FundTransaction> {
    const response = await api.post<FundTransaction>('/fund/deposit', data)
    return response.data
  },

  async withdraw(data: CreateWithdrawalRequest): Promise<FundTransaction> {
    const response = await api.post<FundTransaction>('/fund/withdraw', data)
    return response.data
  },
}
