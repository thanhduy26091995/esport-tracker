import { getIntlLocale } from '@/utils/intl'

/**
 * Format a date string using the active locale.
 */
export function formatDate(dateString: string): string {
  try {
    const date = new Date(dateString)
    return date.toLocaleDateString(getIntlLocale(), {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit'
    })
  } catch {
    return dateString
  }
}

/**
 * Format a date string with date and time using the active locale.
 */
export function formatDateTime(dateString: string): string {
  try {
    const date = new Date(dateString)
    return date.toLocaleString(getIntlLocale(), {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    })
  } catch {
    return dateString
  }
}

/**
 * Format a date string as relative time using the active locale.
 */
export function formatRelativeTime(dateString: string): string {
  try {
    const date = new Date(dateString)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffMins = Math.floor(diffMs / 60000)
    const diffHours = Math.floor(diffMins / 60)
    const diffDays = Math.floor(diffHours / 24)
    const relativeTime = new Intl.RelativeTimeFormat(getIntlLocale(), { numeric: 'auto' })

    if (diffMins < 1) return relativeTime.format(0, 'second')
    if (diffMins < 60) return relativeTime.format(-diffMins, 'minute')
    if (diffHours < 24) return relativeTime.format(-diffHours, 'hour')
    if (diffDays < 30) return relativeTime.format(-diffDays, 'day')
    return formatDate(dateString)
  } catch {
    return dateString
  }
}

/**
 * Format a date for input fields (YYYY-MM-DD)
 */
export function formatDateForInput(date: Date = new Date()): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

/**
 * Check if a date is today
 */
export function isToday(dateString: string): boolean {
  try {
    const date = new Date(dateString)
    const today = new Date()
    return date.toDateString() === today.toDateString()
  } catch {
    return false
  }
}
