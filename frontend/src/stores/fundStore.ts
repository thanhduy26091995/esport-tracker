import { defineStore } from 'pinia'
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import type {
  FundTransaction,
  CreateDepositRequest,
  CreateWithdrawalRequest,
  FundStats,
} from '@/types/fund'
import { fundService } from '@/services/fundService'

export const useFundStore = defineStore('fund', () => {
  const transactions = ref<FundTransaction[]>([])
  const balance = ref(0)
  const stats = ref<FundStats>({
    total_deposits: 0,
    total_withdrawals: 0,
    settlement_deposits: 0,
    balance: 0,
  })
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Actions
  async function fetchBalance() {
    try {
      const data = await fundService.getBalance()
      balance.value = data.balance
    } catch (err: any) {
      console.error('Failed to fetch fund balance:', err)
    }
  }

  async function fetchStats() {
    try {
      stats.value = await fundService.getStats()
      balance.value = stats.value.balance
    } catch (err: any) {
      console.error('Failed to fetch fund stats:', err)
    }
  }

  async function fetchTransactions(params?: { page?: number; limit?: number; type?: string }) {
    loading.value = true
    error.value = null
    try {
      transactions.value = await fundService.getTransactions(params)
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch transactions'
      if (error.value) ElMessage.error(error.value)
    } finally {
      loading.value = false
    }
  }

  async function deposit(data: CreateDepositRequest) {
    loading.value = true
    error.value = null
    try {
      const transaction = await fundService.deposit(data)
      transactions.value.unshift(transaction)
      balance.value += transaction.amount
      stats.value.total_deposits += transaction.amount
      stats.value.balance = balance.value
      ElMessage.success(`Deposited ${transaction.amount.toLocaleString('vi-VN')} VND`)
      return transaction
    } catch (err: any) {
      const errorMsg =
        err.response?.data?.error?.message || err.message || 'Failed to deposit'
      error.value = errorMsg
      ElMessage.error(errorMsg)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function withdraw(data: CreateWithdrawalRequest) {
    loading.value = true
    error.value = null
    try {
      const transaction = await fundService.withdraw(data)
      transactions.value.unshift(transaction)
      balance.value -= transaction.amount
      stats.value.total_withdrawals += transaction.amount
      stats.value.balance = balance.value
      ElMessage.success(`Withdrew ${transaction.amount.toLocaleString('vi-VN')} VND`)
      return transaction
    } catch (err: any) {
      const errorMsg =
        err.response?.data?.error?.message || err.message || 'Failed to withdraw'
      error.value = errorMsg
      ElMessage.error(errorMsg)
      throw err
    } finally {
      loading.value = false
    }
  }

  return {
    transactions,
    balance,
    stats,
    loading,
    error,
    fetchBalance,
    fetchStats,
    fetchTransactions,
    deposit,
    withdraw,
  }
})
