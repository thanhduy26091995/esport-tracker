export type TransactionType = 'deposit' | 'withdrawal'

export interface FundTransaction {
  id: string
  amount: number
  transaction_type: TransactionType
  description: string
  related_settlement_id?: string
  transaction_date: string
  created_at: string
}

export interface CreateDepositRequest {
  amount: number
  description: string
  date?: string
}

export interface CreateWithdrawalRequest {
  amount: number
  description: string
  date?: string
}

export interface FundBalance {
  balance: number
}

export interface FundStats {
  total_deposits: number
  total_withdrawals: number
  settlement_deposits: number
  balance: number
}
