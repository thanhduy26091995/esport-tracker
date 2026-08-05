import axios from 'axios'
import type { WcMatch, WcMatchFilter, WcStandingsResponse } from '@/types/wc'

const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'

// Bare instances with no error interceptors — used by Dashboard for silent background fetches.
// wcApi/acApi (with full error handling + toasts) remain unchanged for tournament pages.
function publicApiFor(prefix: string) {
  return axios.create({
    baseURL: `${API_BASE}/${prefix}`,
    timeout: 10000,
    headers: { 'Content-Type': 'application/json' },
  })
}

export async function listMatchesPublic(filter: WcMatchFilter, prefix: string): Promise<WcMatch[]> {
  const r = await publicApiFor(prefix).get<WcMatch[]>('/matches', { params: filter })
  return r.data
}

export async function getStandings(prefix = 'wc'): Promise<WcStandingsResponse> {
  const r = await publicApiFor(prefix).get<WcStandingsResponse>('/standings')
  return r.data
}

export async function getTournamentConfig(prefix: string): Promise<{ is_enabled: boolean }> {
  const r = await publicApiFor(prefix).get<{ is_enabled: boolean }>('/config')
  return r.data
}
