import { ref, computed } from 'vue'
import type { Ref } from 'vue'
import type { WcMatch } from '@/types/wc'

export type MatchFilterKey = 'incoming' | 'open' | 'live' | 'locked' | 'completed' | 'all'

export interface MatchFilterOption {
  key: MatchFilterKey
  label: string
  count?: number
}

export function useMatchFilter(matches: Ref<WcMatch[]>, defaultFilter: MatchFilterKey = 'incoming') {
  const search = ref('')
  const activeFilter = ref<MatchFilterKey>(defaultFilter)

  function isLocked(m: WcMatch): boolean {
    return !!m.bets_locked_at && new Date(m.bets_locked_at) <= new Date()
  }

  const counts = computed(() => {
    const result: Record<MatchFilterKey, number> = {
      incoming: 0, open: 0, live: 0, locked: 0, completed: 0, all: matches.value.length,
    }
    for (const m of matches.value) {
      if (m.status === 'scheduled') result.incoming++
      if (m.status === 'live') result.live++
      if (m.status === 'completed') result.completed++
      if (isLocked(m) && m.status !== 'completed') result.locked++
      if (m.status !== 'completed' && m.status !== 'cancelled' && !isLocked(m)) result.open++
    }
    return result
  })

  const filtered = computed(() => {
    let list = matches.value

    switch (activeFilter.value) {
      case 'incoming':
        list = list.filter(m => m.status === 'scheduled' && !isLocked(m))
        break
      case 'open':
        list = list.filter(m =>
          m.status !== 'completed' &&
          m.status !== 'cancelled' &&
          !isLocked(m),
        )
        break
      case 'live':
        list = list.filter(m => m.status === 'live')
        break
      case 'locked':
        list = list.filter(m => isLocked(m) && m.status !== 'completed')
        break
      case 'completed':
        list = list.filter(m => m.status === 'completed')
        break
    }

    if (search.value.trim()) {
      const q = search.value.trim().toLowerCase()
      list = list.filter(m =>
        m.home_team.toLowerCase().includes(q) ||
        m.away_team.toLowerCase().includes(q) ||
        (m.home_team_code?.toLowerCase().includes(q) ?? false) ||
        (m.away_team_code?.toLowerCase().includes(q) ?? false),
      )
    }

    return [...list].sort((a, b) => {
      const dir = activeFilter.value === 'completed' ? -1 : 1
      return dir * (new Date(a.match_date).getTime() - new Date(b.match_date).getTime())
    })
  })

  return { search, activeFilter, filtered, counts }
}
