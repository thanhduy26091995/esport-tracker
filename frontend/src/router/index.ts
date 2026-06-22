import { createRouter, createWebHistory } from 'vue-router'

const WC_API_BASE = (import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1') + '/wc'

async function isWcFeatureEnabled(): Promise<boolean> {
  try {
    const res = await fetch(`${WC_API_BASE}/config`)
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
          component: () => import('../views/WcScheduleView.vue')
        },
        {
          path: '/world-cup/login',
          name: 'wc-login',
          component: () => import('../views/WcLoginView.vue')
        },
        {
          path: '/world-cup/register',
          redirect: '/world-cup/login'
        },
        {
          path: '/world-cup/link-google',
          name: 'wc-link-google',
          component: () => import('../views/WcLinkGoogleView.vue'),
          meta: { requiresWcAuth: true, skipGoogleLinkCheck: true }
        },
        {
          path: '/world-cup/profile',
          name: 'wc-profile',
          component: () => import('../views/WcProfileView.vue'),
          meta: { requiresWcAuth: true, requiresGoogleLink: true }
        },
        {
          path: '/world-cup/predict',
          name: 'wc-predict',
          component: () => import('../views/WcPredictView.vue'),
          meta: { requiresWcAuth: true, requiresGoogleLink: true, requiresWcFeature: true }
        },
        {
          path: '/world-cup/bet',
          name: 'wc-bet',
          component: () => import('../views/WcBettingView.vue'),
          meta: { requiresWcAuth: true, requiresGoogleLink: true, requiresWcFeature: true }
        },
        {
          path: '/world-cup/admin',
          name: 'wc-admin',
          component: () => import('../views/WcAdminView.vue'),
          meta: { requiresWcAuth: true, requiresWcAdmin: true }
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
        {
          path: '/world-cup',
          name: 'wc-schedule',
          component: () => import('../views/WcScheduleView.vue')
        },
        {
          path: '/world-cup/login',
          name: 'wc-login',
          component: () => import('../views/WcLoginView.vue')
        },
        // /world-cup/register removed — redirect to login
        {
          path: '/world-cup/register',
          redirect: '/world-cup/login'
        },
        {
          path: '/world-cup/link-google',
          name: 'wc-link-google',
          component: () => import('../views/WcLinkGoogleView.vue'),
          meta: { requiresWcAuth: true, skipGoogleLinkCheck: true }
        },
        {
          path: '/world-cup/profile',
          name: 'wc-profile',
          component: () => import('../views/WcProfileView.vue'),
          meta: { requiresWcAuth: true, requiresGoogleLink: true }
        },
        {
          path: '/world-cup/predict',
          name: 'wc-predict',
          component: () => import('../views/WcPredictView.vue'),
          meta: { requiresWcAuth: true, requiresGoogleLink: true, requiresWcFeature: true }
        },
        {
          path: '/world-cup/bet',
          name: 'wc-bet',
          component: () => import('../views/WcBettingView.vue'),
          meta: { requiresWcAuth: true, requiresGoogleLink: true, requiresWcFeature: true }
        },
        {
          path: '/world-cup/admin',
          name: 'wc-admin',
          component: () => import('../views/WcAdminView.vue'),
          meta: { requiresWcAuth: true, requiresWcAdmin: true }
        },
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
  if (to.meta.requiresWcFeature) {
    const enabled = await isWcFeatureEnabled()
    if (!enabled) return { name: 'not-found' }
  }

  if (to.meta.requiresWcAuth) {
    const token = localStorage.getItem('wc_token')
    if (!token) return { name: 'wc-login' }
  }

  if (to.meta.requiresWcAdmin) {
    try {
      const raw = localStorage.getItem('wc_user')
      const user = raw ? JSON.parse(raw) : null
      if (!user?.isAdmin) return { name: 'wc-schedule' }
    } catch {
      return { name: 'wc-schedule' }
    }
  }

  // Google link gate: block access if user hasn't linked Google yet
  if (to.meta.requiresGoogleLink) {
    try {
      const raw = localStorage.getItem('wc_user')
      const user = raw ? JSON.parse(raw) : null
      if (user && !user.googleLinked) {
        return { name: 'wc-link-google' }
      }
    } catch {
      return { name: 'wc-login' }
    }
  }
})

export default router
