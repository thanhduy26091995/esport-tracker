import axios from 'axios'
import type { WcMatch, WcMatchFilter } from '@/types/wc'

const BASE = (import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1') + '/wc'

// Bare instance with no error interceptors — used by Dashboard for silent background fetches.
// wcApi (with full error handling + toasts) remains unchanged for WC-specific pages.
export const wcPublicApi = axios.create({
  baseURL: BASE,
  timeout: 10000,
  headers: { 'Content-Type': 'application/json' },
})

export async function listMatchesPublic(filter: WcMatchFilter = {}): Promise<WcMatch[]> {
  const r = await wcPublicApi.get<WcMatch[]>('/matches', { params: filter })
  return r.data
}
