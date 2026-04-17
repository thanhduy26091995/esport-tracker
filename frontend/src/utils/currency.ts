import { getIntlLocale } from '@/utils/intl'

export function formatVND(amount: number): string {
  return new Intl.NumberFormat(getIntlLocale(), {
    style: 'currency',
    currency: 'VND',
    minimumFractionDigits: 0,
  }).format(amount)
}

export function pointsToVND(points: number, conversionRate: number): number {
  return points * conversionRate
}
