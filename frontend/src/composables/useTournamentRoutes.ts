import { computed } from 'vue'
import { useRoute } from 'vue-router'

export function useTournamentRoutes() {
  const route = useRoute()
  const isAc = computed(() => route.meta?.tournamentType === 'asean_cup')

  const scheduleRoute = computed(() => isAc.value ? { name: 'ac-schedule' } : { name: 'wc-schedule' })
  const loginRoute = computed(() => isAc.value ? { name: 'ac-login' } : { name: 'wc-login' })
  const predictRoute = computed(() => isAc.value ? { name: 'ac-predict' } : { name: 'wc-predict' })
  const betRoute = computed(() => isAc.value ? { name: 'ac-bet' } : { name: 'wc-bet' })
  const adminRoute = computed(() => isAc.value ? { name: 'ac-admin' } : { name: 'wc-admin' })
  const profilePath = computed(() => isAc.value ? '/asean-cup/profile' : '/world-cup/profile')
  const loginPath = computed(() => isAc.value ? '/asean-cup/login' : '/world-cup/login')
  const predictPath = computed(() => isAc.value ? '/asean-cup/predict' : '/world-cup/predict')
  const tournamentTitle = computed(() => isAc.value ? '🏆 ASEAN Cup 2026' : '🏆 World Cup 2026')
  // Bare name for places that already show a trophy icon of their own (e.g. login card).
  const tournamentName = computed(() => isAc.value ? 'ASEAN Cup 2026' : 'World Cup 2026')

  return {
    isAc,
    scheduleRoute,
    loginRoute,
    predictRoute,
    betRoute,
    adminRoute,
    profilePath,
    loginPath,
    predictPath,
    tournamentTitle,
    tournamentName,
  }
}
