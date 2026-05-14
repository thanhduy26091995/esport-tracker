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
            <div class="logo-title">{{ t('common.appName') }}</div>
            <div class="logo-sub">{{ t('layout.sidebarSubtitle') }}</div>
          </div>
        </div>
        <nav class="sidebar-nav">
          <router-link
            v-for="item in navigation"
            :key="item.navKey"
            :to="item.href"
            @click="mobileMenuOpen = false"
            class="nav-item"
            :class="{ 'nav-item--active': isActiveRoute(item.href) }"
          >
            <el-icon :size="18" class="nav-icon"><component :is="item.icon" /></el-icon>
            <span>{{ t(item.navKey) }}</span>
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
            <div class="logo-title">{{ t('common.appName') }}</div>
            <div class="logo-sub">{{ t('layout.sidebarSubtitle') }}</div>
          </div>
        </div>
        <nav class="sidebar-nav">
          <router-link
            v-for="item in navigation"
            :key="item.navKey"
            :to="item.href"
            class="nav-item"
            :class="{ 'nav-item--active': isActiveRoute(item.href) }"
          >
            <el-icon :size="18" class="nav-icon"><component :is="item.icon" /></el-icon>
            <span>{{ t(item.navKey) }}</span>
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
          <span class="topbar-date">{{ todayLabel }}</span>
          <div class="topbar-lang">
            <LanguageSwitcher />
          </div>
        </div>
      </header>

      <!-- Scrollable page -->
      <main class="page-scroll">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Menu, Trophy, HomeFilled, UserFilled, TrendCharts, DocumentCopy, Wallet, Setting, Grid } from '@element-plus/icons-vue'
import LanguageSwitcher from '@/components/common/LanguageSwitcher.vue'

const { t } = useI18n()
const route = useRoute()
const mobileMenuOpen = ref(false)

const navigation = [
  { navKey: 'nav.dashboard', href: '/', icon: HomeFilled },
  { navKey: 'nav.players', href: '/users', icon: UserFilled },
  { navKey: 'nav.matches', href: '/matches', icon: TrendCharts },
  { navKey: 'nav.tournaments', href: '/tournaments', icon: Grid },
  { navKey: 'nav.settlements', href: '/settlements', icon: DocumentCopy },
  { navKey: 'nav.fund', href: '/fund', icon: Wallet },
  { navKey: 'nav.settings', href: '/settings', icon: Setting },
]

const isActiveRoute = (href: string): boolean =>
  href === '/' ? route.path === '/' : route.path.startsWith(href)

const currentNavItem = computed(() => navigation.find(item => isActiveRoute(item.href)))
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
  background: linear-gradient(135deg, #3b82f6, #1d4ed8);
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.4);
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
  background: rgba(59, 130, 246, 0.15);
  color: #60a5fa;
  box-shadow: inset 3px 0 0 #3b82f6;
}

.nav-icon {
  flex-shrink: 0;
  opacity: 0.8;
}

.nav-item--active .nav-icon {
  opacity: 1;
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
  border-bottom: 1px solid var(--border-default);
  box-shadow: 0 1px 3px rgba(0,0,0,0.04);
  flex-shrink: 0;
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
  color: var(--color-primary);
  flex-shrink: 0;
  opacity: 0.8;
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

/* ── Scrollable page ── */
.page-scroll {
  flex: 1;
  overflow-y: auto;
}
</style>
