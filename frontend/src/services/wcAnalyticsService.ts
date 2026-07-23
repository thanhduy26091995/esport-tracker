import { wcService } from './wcService'
import type {
  MyAnalyticsResponse,
  CommunityAnalyticsResponse,
  CompareAnalyticsResponse,
  WcAnalyticsResponse,
} from '@/types/wc'

// Delegates to wcService proxy — automatically routes to WC or AC based on active tournament.
export const wcAnalyticsService = {
  getMyAnalytics(period = '30d', dateFrom?: string, dateTo?: string): Promise<MyAnalyticsResponse> {
    return (wcService as any).getMyAnalytics(period, dateFrom, dateTo)
  },
  getCommunityAnalytics(): Promise<CommunityAnalyticsResponse> {
    return (wcService as any).getCommunityAnalytics()
  },
  getCompareAnalytics(): Promise<CompareAnalyticsResponse> {
    return (wcService as any).getCompareAnalytics()
  },
  getWC2026Analytics(): Promise<WcAnalyticsResponse> {
    return (wcService as any).getTournamentAnalytics()
  },
}
