<template>
  <div class="app-shell">
    <!-- Mobile sidebar drawer -->
    <el-drawer
      v-model="mobileMenuOpen"
      direction="ltr"
      :show-close="false"
      size="272px"
      :with-header="false"
    >
      <div class="sidebar-inner">
        <div class="sidebar-logo">
          <div class="logo-icon">
            <el-icon color="white" :size="20"><Trophy /></el-icon>
          </div>
          <div>
            <div class="logo-title">{{ appTitle }}</div>
            <div class="logo-sub">{{ appSubtitle }}</div>
          </div>
        </div>
        <nav class="sidebar-nav">
          <router-link
            v-for="item in navigation"
            :key="item.navKey"
            :to="item.href"
            @click="mobileMenuOpen = false"
            class="nav-item"
            :class="{ 'nav-item--active': isActiveRoute(item), 'nav-item--wc': item.highlight === 'wc', 'nav-item--ac': item.highlight === 'ac' }"
          >
            <el-icon :size="18" class="nav-icon"><component :is="item.icon" /></el-icon>
            <span>{{ t(item.navKey) }}</span>
            <span v-if="item.highlight === 'wc' || item.highlight === 'ac'" class="nav-item-badge" :class="{ 'nav-item-badge--ac': item.highlight === 'ac' }">2026</span>
          </router-link>
        </nav>
        <div class="sidebar-lang">
          <LanguageSwitcher />
        </div>
        <div class="sidebar-footer">{{ t('layout.version') }} v1.0.0 · {{ t('common.appName') }}</div>
      </div>
    </el-drawer>

    <!-- Desktop sidebar -->
    <aside class="sidebar-desktop">
      <div class="sidebar-inner">
        <div class="sidebar-logo">
          <div class="logo-icon">
            <el-icon color="white" :size="20"><Trophy /></el-icon>
          </div>
          <div>
            <div class="logo-title">{{ appTitle }}</div>
            <div class="logo-sub">{{ appSubtitle }}</div>
          </div>
        </div>
        <nav class="sidebar-nav">
          <router-link
            v-for="item in navigation"
            :key="item.navKey"
            :to="item.href"
            class="nav-item"
            :class="{ 'nav-item--active': isActiveRoute(item), 'nav-item--wc': item.highlight === 'wc', 'nav-item--ac': item.highlight === 'ac' }"
          >
            <el-icon :size="18" class="nav-icon"><component :is="item.icon" /></el-icon>
            <span>{{ t(item.navKey) }}</span>
            <span v-if="item.highlight === 'wc' || item.highlight === 'ac'" class="nav-item-badge" :class="{ 'nav-item-badge--ac': item.highlight === 'ac' }">2026</span>
          </router-link>
        </nav>
        <div class="sidebar-lang">
          <LanguageSwitcher />
        </div>
        <div class="sidebar-footer">{{ t('layout.version') }} v1.0.0 · © 2024 {{ t('common.appName') }}</div>
      </div>
    </aside>

    <!-- Main content -->
    <div class="main-area">
      <!-- Topbar (mobile + desktop) -->
      <header class="topbar">
        <div class="topbar-left">
          <button class="mobile-menu-btn" @click="mobileMenuOpen = true">
            <el-icon :size="20"><Menu /></el-icon>
          </button>
          <div class="topbar-page-info">
            <el-icon :size="15" class="topbar-page-icon"><component :is="currentPageIcon" /></el-icon>
            <span class="topbar-page-title">{{ currentPageName }}</span>
          </div>
        </div>
        <div class="topbar-right">
          <div v-if="reigningClub" class="topbar-champion">
            <span class="topbar-champion-icon">🏆</span>
            <span class="topbar-champion-info">
              <span class="topbar-champion-club">{{ reigningClub.name }}</span>
              <span class="topbar-champion-player">{{ reigningPlayerName }}</span>
            </span>
          </div>
          <span class="topbar-date">{{ todayLabel }}</span>
          <div class="topbar-lang">
            <LanguageSwitcher />
          </div>
        </div>
      </header>

      <!-- WC Top-3 Honor Banner -->
      <WcTop3Banner v-if="isWcRoute && wcAuth.isLoggedIn" />

      <!-- Scrollable page -->
      <main class="page-scroll">
        <router-view />
      </main>
    </div>

    <!-- WC Live Chat — floating button + panel (WC/soc builds only) -->
    <template v-if="isSocSite || isWcRoute">
      <WcChatButton />
      <WcChatPanel />
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Menu, Trophy, HomeFilled, UserFilled, TrendCharts, DocumentCopy, Wallet, Setting, Grid, Promotion } from '@element-plus/icons-vue'
import LanguageSwitcher from '@/components/common/LanguageSwitcher.vue'
import { useUserStore } from '@/stores/userStore'
import { CLUBS } from '@/config/clubs'
import { useActivityFeed } from '@/composables/useActivityFeed'
import { useChatStore } from '@/stores/chatStore'
import { useWcAuthStore } from '@/stores/wcAuthStore'
import WcChatButton from '@/components/wc/WcChatButton.vue'
import WcChatPanel from '@/components/wc/WcChatPanel.vue'
import WcTop3Banner from '@/components/wc/WcTop3Banner.vue'

useActivityFeed()

const wcAuth = useWcAuthStore()
const chatStore = useChatStore()
if (wcAuth.isLoggedIn) {
  chatStore.fetchUnreadMentionCount()
}

const { t } = useI18n()
const route = useRoute()

const isSocSite = import.meta.env.VITE_SITE === 'soc'
const isWcRoute = computed(() => route.path.startsWith('/world-cup') || route.path.startsWith('/asean-cup'))
const appTitle = isSocSite ? t('common.appNameSoc') : t('common.appName')
const appSubtitle = isSocSite ? t('layout.sidebarSubtitleSoc') : t('layout.sidebarSubtitle')
const mobileMenuOpen = ref(false)
const userStore = useUserStore()

const reigningClub = computed(() => {
  const slug = userStore.users[0]?.favorite_club
  if (!slug || slug === 'none') return null
  return CLUBS.find(c => c.slug === slug) ?? null
})

const reigningPlayerName = computed(() => userStore.users[0]?.name ?? '')

const navigation = isSocSite
  ? [
      { navKey: 'nav.aseanCup', href: '/asean-cup/predict', activePrefix: '/asean-cup', icon: Promotion, highlight: 'ac' }
    ]
  : [
      { navKey: 'nav.dashboard', href: '/', icon: HomeFilled },
      { navKey: 'nav.players', href: '/users', icon: UserFilled },
      { navKey: 'nav.matches', href: '/matches', icon: TrendCharts },
      { navKey: 'nav.tournaments', href: '/tournaments', icon: Grid },
      { navKey: 'nav.settlements', href: '/settlements', icon: DocumentCopy },
      { navKey: 'nav.fund', href: '/fund', icon: Wallet },
      { navKey: 'nav.aseanCup', href: '/asean-cup/predict', activePrefix: '/asean-cup', icon: Promotion, highlight: 'ac' },
      { navKey: 'nav.settings', href: '/settings', icon: Setting },
    ]

const isActiveRoute = (item: { href: string; activePrefix?: string }): boolean => {
  const prefix = item.activePrefix ?? item.href
  return prefix === '/' ? route.path === '/' : route.path.startsWith(prefix)
}

const currentNavItem = computed(() => navigation.find(item => isActiveRoute(item)))
const currentPageName = computed(() => currentNavItem.value ? t(currentNavItem.value.navKey) : t('common.appName'))
const currentPageIcon = computed(() => currentNavItem.value?.icon ?? HomeFilled)

const todayLabel = computed(() => {
  const now = new Date()
  return now.toLocaleDateString(undefined, { weekday: 'short', month: 'short', day: 'numeric' })
})
</script>

<style scoped>
.app-shell {
  display: flex;
  height: 100vh;
  overflow: hidden;
  background: var(--surface-page);
}

/* ── Sidebar ── */
.sidebar-desktop {
  display: none;
  flex-direction: column;
  width: 240px;
  flex-shrink: 0;
}

@media (min-width: 1024px) {
  .sidebar-desktop {
    display: flex;
  }
}

.sidebar-inner {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--sidebar-bg);
}

.sidebar-logo {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 20px 20px 16px;
  border-bottom: 1px solid var(--sidebar-border);
}

.logo-icon {
  width: 36px;
  height: 36px;
  background: var(--theme-gradient);
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 4px 12px var(--theme-glow);
  transition: background 0.5s ease, box-shadow 0.5s ease;
}

.logo-icon--sm {
  width: 32px;
  height: 32px;
  border-radius: 8px;
}

.logo-title {
  font-size: 14px;
  font-weight: 700;
  color: #ffffff;
  line-height: 1.2;
}

.logo-sub {
  font-size: 11px;
  color: var(--sidebar-text);
  line-height: 1.2;
  margin-top: 1px;
}

.sidebar-nav {
  flex: 1;
  padding: 12px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 12px;
  border-radius: 8px;
  text-decoration: none;
  font-size: 13px;
  font-weight: 500;
  color: var(--sidebar-text);
  transition: background 0.15s ease, color 0.15s ease;
}

.nav-item:hover {
  background: rgba(255, 255, 255, 0.05);
  color: #cbd5e1;
}

.nav-item--active {
  background: var(--theme-bg);
  color: var(--theme-accent);
  box-shadow: inset 3px 0 0 var(--theme-primary);
  transition: background 0.5s ease, color 0.5s ease, box-shadow 0.5s ease;
}

.nav-icon {
  flex-shrink: 0;
  opacity: 0.8;
}

.nav-item--active .nav-icon {
  opacity: 1;
}

.nav-item--wc {
  background: linear-gradient(90deg, rgba(22, 163, 74, 0.15) 0%, transparent 100%);
  border: 1px solid rgba(22, 163, 74, 0.2);
  color: #16a34a;
}

.nav-item--wc .nav-icon {
  opacity: 1;
  color: #16a34a;
}

.nav-item--ac {
  background: linear-gradient(90deg, rgba(234, 88, 12, 0.15) 0%, transparent 100%);
  border: 1px solid rgba(234, 88, 12, 0.2);
  color: #ea580c;
}

.nav-item--ac .nav-icon {
  opacity: 1;
  color: #ea580c;
}

.nav-item-badge {
  margin-left: auto;
  font-size: 10px;
  font-weight: 700;
  background: rgba(22, 163, 74, 0.18);
  color: #16a34a;
  padding: 1px 6px;
  border-radius: 10px;
  letter-spacing: 0.02em;
  flex-shrink: 0;
}

.nav-item-badge--ac {
  background: rgba(234, 88, 12, 0.18);
  color: #ea580c;
}

.sidebar-footer {
  padding: 16px 20px;
  border-top: 1px solid var(--sidebar-border);
  font-size: 11px;
  color: #334155;
}

.sidebar-lang {
  padding: 10px 16px 8px;
  border-top: 1px solid var(--sidebar-border);
}

/* ── Main area ── */
.main-area {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;
  overflow: hidden;
}

/* ── Topbar (all screen sizes) ── */
.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 0 20px;
  height: 52px;
  background: var(--surface-card);
  border-bottom: 3px solid var(--theme-primary);
  box-shadow: 0 1px 3px rgba(0,0,0,0.04);
  flex-shrink: 0;
  transition: border-color 0.5s ease;
}

.topbar-left {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.topbar-page-info {
  display: flex;
  align-items: center;
  gap: 7px;
  min-width: 0;
}

.topbar-page-icon {
  color: var(--theme-primary);
  flex-shrink: 0;
  opacity: 0.9;
  transition: color 0.5s ease;
}

.topbar-page-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  letter-spacing: -0.01em;
}

.topbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}

.topbar-date {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-muted);
  white-space: nowrap;
}

.topbar-lang :deep(.language-label) {
  display: none;
}

/* ── Champion pill ── */
.topbar-champion {
  display: flex;
  align-items: center;
  gap: 7px;
  padding: 5px 12px 5px 8px;
  border-radius: 20px;
  background: var(--theme-gradient);
  box-shadow: 0 2px 10px var(--theme-glow);
  transition: background 0.5s ease, box-shadow 0.5s ease;
  cursor: default;
  white-space: nowrap;
}

.topbar-champion-icon {
  font-size: 15px;
  line-height: 1;
  flex-shrink: 0;
}

.topbar-champion-info {
  display: flex;
  flex-direction: column;
  line-height: 1.15;
}

.topbar-champion-club {
  font-size: 11px;
  font-weight: 700;
  color: var(--theme-text-on-primary);
  letter-spacing: 0.03em;
  text-transform: uppercase;
}

.topbar-champion-player {
  font-size: 10px;
  color: var(--theme-text-on-primary);
  opacity: 0.8;
}

.mobile-menu-btn {
  padding: 6px;
  border: none;
  background: none;
  border-radius: 8px;
  color: var(--text-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  transition: background 0.15s;
  flex-shrink: 0;
}

.mobile-menu-btn:hover {
  background: var(--surface-subtle);
}

@media (min-width: 1024px) {
  .mobile-menu-btn {
    display: none;
  }

  .topbar-date {
    display: block;
  }
}

@media (max-width: 1023px) {
  .topbar-date {
    display: none;
  }
}

@media (max-width: 480px) {
  .topbar-champion-player {
    display: none;
  }
}

/* ── Scrollable page ── */
.page-scroll {
  flex: 1;
  overflow-y: auto;
}
</style>
