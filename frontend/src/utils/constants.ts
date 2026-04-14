// API Configuration
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'

// Business Rules
export const DEBT_THRESHOLD_DEFAULT = -6
export const POINT_TO_VND_DEFAULT = 22000
export const FUND_SPLIT_PERCENT_DEFAULT = 50

// Pagination
export const DEFAULT_PAGE_SIZE = 20
export const LEADERBOARD_DEFAULT_SIZE = 10

// Match Types
export const MATCH_TYPES = ['1v1', '2v2'] as const

// Transaction Types
export const TRANSACTION_TYPES = ['deposit', 'withdrawal'] as const

// Config Keys
export const CONFIG_KEYS = ['debt_threshold', 'point_to_vnd', 'fund_split_percent'] as const

// UI Constants
export const DEBOUNCE_DELAY = 300
export const TOAST_DURATION = 3000

// Color Codes
export const SCORE_COLORS = {
  positive: '#67C23A', // Success green
  negative: '#F56C6C', // Danger red
  neutral: '#909399', // Info gray
  warning: '#E6A23C', // Warning orange
}

export const TEAM_COLORS = {
  team1: '#409EFF', // Primary blue
  team2: '#E6A23C', // Orange
}
