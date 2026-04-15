import type { User } from './user'

export interface DebtSettlement {
  id: string
  debtor_id: string
  debt_amount: number
  money_amount: number
  fund_amount: number
  winner_distribution: number
  original_debt_points: number
  settlement_date: string
  created_at: string
  debtor: User
  winners: SettlementWinner[]
}

export interface SettlementWinner {
  id: string
  settlement_id: string
  winner_id: string
  money_amount: number
  points_deducted: number
  winner: User
}

export interface TriggerSettlementRequest {
  debtor_id: string
  winner_ids?: string[]
}

export interface SettlementStats {
  total: number
  today: number
}
