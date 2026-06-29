import axios from 'axios'
import type { WcMatch, WcMatchFilter, WcStandingsResponse } from '@/types/wc'

const BASE = (import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1') + '/wc'

// Bare instance with no error interceptors — used by Dashboard for silent background fetches.
// wcApi (with full error handling + toasts) remains unchanged for WC-specific pages.
export const wcPublicApi = axios.create({
  baseURL: BASE,
  timeout: 10000,
  headers: { 'Content-Type': 'application/json' },
})

wcPublicApi.interceptors.request.use((config) => {
  const siteToken = localStorage.getItem('site_access_token')
  if (siteToken) config.headers['X-Site-Token'] = siteToken
  return config
})

export async function listMatchesPublic(filter: WcMatchFilter = {}): Promise<WcMatch[]> {
  const r = await wcPublicApi.get<WcMatch[]>('/matches', { params: filter })
  return r.data
}

export async function getStandings(): Promise<WcStandingsResponse> {
  const r = await wcPublicApi.get<WcStandingsResponse>('/standings')
  return r.data
}
