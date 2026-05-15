import type { User } from '@/types/user'

export type PlayerSortStrategy = 'debt-first' | 'default' | 'winners-first'

// Handles strategies 'default' and 'winners-first' only.
// 'debt-first' uses backend-sorted paymentRankingUsers — do not pass it here.
export function sortByStrategy<T extends User>(users: T[], strategy: PlayerSortStrategy): T[] {
  const byName = (a: T, b: T) => a.name.localeCompare(b.name)
  const neg = [...users.filter(u => u.current_score < 0)].sort((a, b) => a.current_score - b.current_score || byName(a, b))
  const pos = [...users.filter(u => u.current_score > 0)].sort((a, b) => b.current_score - a.current_score || byName(a, b))
  const zer = [...users.filter(u => u.current_score === 0)].sort(byName)

  if (strategy === 'winners-first') return [...pos, ...zer, ...neg]
  return [...neg, ...pos, ...zer] // 'default' + any unknown fallback
}
