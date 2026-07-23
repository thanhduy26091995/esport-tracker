import { createRouter, createWebHistory } from 'vue-router'
import { setActiveTournamentService } from '@/services/wcService'

const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'

async function isTournamentEnabled(prefix: string): Promise<boolean> {
  try {
    const res = await fetch(`${API_BASE}/${prefix}/config`)
    const data = await res.json()
    return !!data.is_enabled
  } catch {
    return false
  }
}

const isSocSite = import.meta.env.VITE_SITE === 'soc'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: isSocSite
    ? [
        { path: '/', redirect: '/world-cup/login' },
        {
          path: '/world-cup',
          name: 'wc-schedule',
          component: () => import('../views/WcScheduleView.vue'),
          meta: { tournamentType: 'world_cup' }
        },
        {
          path: '/world-cup/login',
          name: 'wc-login',
          component: () => import('../views/WcLoginView.vue'),
          meta: { tournamentType: 'world_cup' }
        },
        {
          path: '/world-cup/register',
          redirect: '/world-cup/login'
        },
        {
          path: '/world-cup/link-google',
          name: 'wc-link-google',
          component: () => import('../views/WcLinkGoogleView.vue'),
          meta: { tournamentType: 'world_cup', requiresWcAuth: true, skipGoogleLinkCheck: true }
        },
        {
          path: '/world-cup/profile',
          name: 'wc-profile',
          component: () => import('../views/WcProfileView.vue'),
          meta: { tournamentType: 'world_cup', requiresWcAuth: true, requiresGoogleLink: true }
        },
        {
          path: '/world-cup/predict',
          name: 'wc-predict',
          component: () => import('../views/WcPredictView.vue'),
          meta: { tournamentType: 'world_cup', requiresWcAuth: true, requiresGoogleLink: true, requiresTournamentFeature: true }
        },
        {
          path: '/world-cup/bet',
          name: 'wc-bet',
          component: () => import('../views/WcBettingView.vue'),
          meta: { tournamentType: 'world_cup', requiresWcAuth: true, requiresGoogleLink: true, requiresTournamentFeature: true }
        },
        {
          path: '/world-cup/admin',
          name: 'wc-admin',
          component: () => import('../views/WcAdminView.vue'),
          meta: { tournamentType: 'world_cup', requiresWcAuth: true, requiresWcAdmin: true }
        },
        // --- ASEAN Cup (soc build) ---
        {
          path: '/asean-cup',
          name: 'ac-schedule',
          component: () => import('../views/WcScheduleView.vue'),
          meta: { tournamentType: 'asean_cup' }
        },
        {
          path: '/asean-cup/login',
          name: 'ac-login',
          component: () => import('../views/WcLoginView.vue'),
          meta: { tournamentType: 'asean_cup' }
        },
        {
          path: '/asean-cup/link-google',
          name: 'ac-link-google',
          component: () => import('../views/WcLinkGoogleView.vue'),
          meta: { tournamentType: 'asean_cup', requiresWcAuth: true, skipGoogleLinkCheck: true }
        },
        {
          path: '/asean-cup/profile',
          name: 'ac-profile',
          component: () => import('../views/WcProfileView.vue'),
          meta: { tournamentType: 'asean_cup', requiresWcAuth: true, requiresGoogleLink: true }
        },
        {
          path: '/asean-cup/predict',
          name: 'ac-predict',
          component: () => import('../views/WcPredictView.vue'),
          meta: { tournamentType: 'asean_cup', requiresWcAuth: true, requiresGoogleLink: true, requiresTournamentFeature: true }
        },
        {
          path: '/asean-cup/bet',
          name: 'ac-bet',
          component: () => import('../views/WcBettingView.vue'),
          meta: { tournamentType: 'asean_cup', requiresWcAuth: true, requiresGoogleLink: true, requiresTournamentFeature: true }
        },
        {
          path: '/asean-cup/admin',
          name: 'ac-admin',
          component: () => import('../views/WcAdminView.vue'),
          meta: { tournamentType: 'asean_cup', requiresWcAuth: true, requiresWcAdmin: true }
        },
        {
          path: '/:pathMatch(.*)*',
          name: 'not-found',
          component: () => import('../views/NotFoundView.vue')
        }
      ]
    : [
        {
          path: '/',
          name: 'dashboard',
          component: () => import('../views/DashboardView.vue')
        },
        // --- World Cup ---
        {
          path: '/world-cup',
          name: 'wc-schedule',
          component: () => import('../views/WcScheduleView.vue'),
          meta: { tournamentType: 'world_cup' }
        },
        {
          path: '/world-cup/login',
          name: 'wc-login',
          component: () => import('../views/WcLoginView.vue'),
          meta: { tournamentType: 'world_cup' }
        },
        {
          path: '/world-cup/register',
          redirect: '/world-cup/login'
        },
        {
          path: '/world-cup/link-google',
          name: 'wc-link-google',
          component: () => import('../views/WcLinkGoogleView.vue'),
          meta: { tournamentType: 'world_cup', requiresWcAuth: true, skipGoogleLinkCheck: true }
        },
        {
          path: '/world-cup/profile',
          name: 'wc-profile',
          component: () => import('../views/WcProfileView.vue'),
          meta: { tournamentType: 'world_cup', requiresWcAuth: true, requiresGoogleLink: true }
        },
        {
          path: '/world-cup/predict',
          name: 'wc-predict',
          component: () => import('../views/WcPredictView.vue'),
          meta: { tournamentType: 'world_cup', requiresWcAuth: true, requiresGoogleLink: true, requiresTournamentFeature: true }
        },
        {
          path: '/world-cup/bet',
          name: 'wc-bet',
          component: () => import('../views/WcBettingView.vue'),
          meta: { tournamentType: 'world_cup', requiresWcAuth: true, requiresGoogleLink: true, requiresTournamentFeature: true }
        },
        {
          path: '/world-cup/admin',
          name: 'wc-admin',
          component: () => import('../views/WcAdminView.vue'),
          meta: { tournamentType: 'world_cup', requiresWcAuth: true, requiresWcAdmin: true }
        },
        // --- ASEAN Cup ---
        {
          path: '/asean-cup',
          name: 'ac-schedule',
          component: () => import('../views/WcScheduleView.vue'),
          meta: { tournamentType: 'asean_cup' }
        },
        {
          path: '/asean-cup/login',
          name: 'ac-login',
          component: () => import('../views/WcLoginView.vue'),
          meta: { tournamentType: 'asean_cup' }
        },
        {
          path: '/asean-cup/link-google',
          name: 'ac-link-google',
          component: () => import('../views/WcLinkGoogleView.vue'),
          meta: { tournamentType: 'asean_cup', requiresWcAuth: true, skipGoogleLinkCheck: true }
        },
        {
          path: '/asean-cup/profile',
          name: 'ac-profile',
          component: () => import('../views/WcProfileView.vue'),
          meta: { tournamentType: 'asean_cup', requiresWcAuth: true, requiresGoogleLink: true }
        },
        {
          path: '/asean-cup/predict',
          name: 'ac-predict',
          component: () => import('../views/WcPredictView.vue'),
          meta: { tournamentType: 'asean_cup', requiresWcAuth: true, requiresGoogleLink: true, requiresTournamentFeature: true }
        },
        {
          path: '/asean-cup/bet',
          name: 'ac-bet',
          component: () => import('../views/WcBettingView.vue'),
          meta: { tournamentType: 'asean_cup', requiresWcAuth: true, requiresGoogleLink: true, requiresTournamentFeature: true }
        },
        {
          path: '/asean-cup/admin',
          name: 'ac-admin',
          component: () => import('../views/WcAdminView.vue'),
          meta: { tournamentType: 'asean_cup', requiresWcAuth: true, requiresWcAdmin: true }
        },
        // --- Core app ---
        {
          path: '/users',
          name: 'users',
          component: () => import('../views/UsersView.vue')
        },
        {
          path: '/matches',
          name: 'matches',
          component: () => import('../views/MatchesView.vue')
        },
        {
          path: '/settlements',
          name: 'settlements',
          component: () => import('../views/SettlementsView.vue')
        },
        {
          path: '/fund',
          name: 'fund',
          component: () => import('../views/FundView.vue')
        },
        {
          path: '/settings',
          name: 'settings',
          component: () => import('../views/ConfigView.vue')
        },
        {
          path: '/tournaments',
          name: 'tournaments',
          component: () => import('../views/TournamentsView.vue')
        },
        {
          path: '/tournaments/create',
          name: 'tournament-create',
          component: () => import('../views/CreateTournamentView.vue')
        },
        {
          path: '/tournaments/:id',
          name: 'tournament-detail',
          component: () => import('../views/TournamentDetailView.vue')
        },
        {
          path: '/:pathMatch(.*)*',
          name: 'not-found',
          component: () => import('../views/NotFoundView.vue')
        }
      ]
})

router.beforeEach(async (to) => {
  const tt = (to.meta.tournamentType as string | undefined) ?? 'world_cup'
  const prefix = tt === 'asean_cup' ? 'ac' : 'wc'
  const loginRouteName = tt === 'asean_cup' ? 'ac-login' : 'wc-login'
  const scheduleRouteName = tt === 'asean_cup' ? 'ac-schedule' : 'wc-schedule'
  const linkGoogleRouteName = tt === 'asean_cup' ? 'ac-link-google' : 'wc-link-google'

  // Sync active service so wcService proxy and stores resolve correctly
  setActiveTournamentService(tt as 'world_cup' | 'asean_cup')

  if (to.meta.requiresTournamentFeature) {
    const enabled = await isTournamentEnabled(prefix)
    if (!enabled) return { name: 'not-found' }
  }

  if (to.meta.requiresWcAuth) {
    const token = localStorage.getItem('wc_token')
    if (!token) return { name: loginRouteName }
  }

  if (to.meta.requiresWcAdmin) {
    try {
      const raw = localStorage.getItem('wc_user')
      const user = raw ? JSON.parse(raw) : null
      if (!user?.isAdmin) return { name: scheduleRouteName }
    } catch {
      return { name: scheduleRouteName }
    }
  }

  if (to.meta.requiresGoogleLink) {
    try {
      const raw = localStorage.getItem('wc_user')
      const user = raw ? JSON.parse(raw) : null
      if (user && !user.googleLinked) {
        return { name: linkGoogleRouteName }
      }
    } catch {
      return { name: loginRouteName }
    }
  }
})

export default router
