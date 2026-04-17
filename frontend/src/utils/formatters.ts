import { getIntlLocale } from '@/utils/intl'

/**
 * Format a number as Vietnamese Dong (VND) using the active locale.
 */
export function formatVND(amount: number): string {
  return new Intl.NumberFormat(getIntlLocale(), {
    style: 'currency',
    currency: 'VND',
    minimumFractionDigits: 0,
  }).format(amount)
}

/**
 * Format a number with locale-aware thousand separators.
 */
export function formatNumber(value: number): string {
  return new Intl.NumberFormat(getIntlLocale()).format(value)
}

/**
 * Convert points to VND based on conversion rate
 * Example: 5 points with 22000 rate => 110000
 */
export function pointsToVND(points: number, conversionRate: number = 22000): number {
  return points * conversionRate
}

/**
 * Format points as VND value
 * Example: 5 points => "110.000 ₫" (at 22000/point)
 */
export function formatPointsAsVND(points: number, conversionRate: number = 22000): string {
  return formatVND(pointsToVND(points, conversionRate))
}
