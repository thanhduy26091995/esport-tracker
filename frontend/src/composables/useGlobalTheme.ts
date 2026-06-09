import { watch, type Ref } from 'vue'
import { CLUBS, DEFAULT_THEME } from '@/config/clubs'

export function useGlobalTheme(leaderboard: Ref<{ favorite_club?: string | null }[]>) {
  let lastClub = ''

  function applyTheme(club: string | null | undefined) {
    const slug = club ?? ''
    if (slug === lastClub) return
    lastClub = slug

    const theme = CLUBS.find(c => c.slug === slug) ?? DEFAULT_THEME
    const root = document.documentElement.style
    root.setProperty('--theme-primary',        theme.primary)
    root.setProperty('--theme-secondary',      theme.secondary)
    root.setProperty('--theme-accent',         theme.accent)
    root.setProperty('--theme-bg',             theme.bg)
    root.setProperty('--theme-gradient',       theme.gradient)
    root.setProperty('--theme-glow',           theme.glow)
    root.setProperty('--theme-text-on-primary', theme.text === 'dark' ? '#000' : '#fff')
  }

  watch(
    () => leaderboard.value[0]?.favorite_club,
    applyTheme,
    { immediate: true },
  )
}
