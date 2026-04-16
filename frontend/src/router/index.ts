import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'dashboard',
      component: () => import('../views/DashboardView.vue')
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
    }
  ]
})

export default router
