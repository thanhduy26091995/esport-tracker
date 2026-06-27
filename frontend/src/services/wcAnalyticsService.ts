import { wcApi } from './wcApi'
import type {
  MyAnalyticsResponse,
  CommunityAnalyticsResponse,
  CompareAnalyticsResponse,
  WcAnalyticsResponse,
} from '@/types/wc'

export const wcAnalyticsService = {
  getMyAnalytics(period = '30d', dateFrom?: string, dateTo?: string): Promise<MyAnalyticsResponse> {
    const params: Record<string, string> = { period }
    if (dateFrom) params.date_from = dateFrom
    if (dateTo) params.date_to = dateTo
    return wcApi.get<MyAnalyticsResponse>('/analytics/my', { params }).then(r => r.data)
  },

  getCommunityAnalytics(): Promise<CommunityAnalyticsResponse> {
    return wcApi.get<CommunityAnalyticsResponse>('/analytics/community').then(r => r.data)
  },

  getCompareAnalytics(): Promise<CompareAnalyticsResponse> {
    return wcApi.get<CompareAnalyticsResponse>('/analytics/compare').then(r => r.data)
  },

  getWC2026Analytics(): Promise<WcAnalyticsResponse> {
    return wcApi.get<WcAnalyticsResponse>('/analytics/world-cup-2026').then(r => r.data)
  },
}
