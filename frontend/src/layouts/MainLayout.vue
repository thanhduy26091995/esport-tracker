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
            <div class="logo-title">FC25 Tracker</div>
            <div class="logo-sub">Esport Score System</div>
          </div>
        </div>
        <nav class="sidebar-nav">
          <router-link
            v-for="item in navigation"
            :key="item.name"
            :to="item.href"
            @click="mobileMenuOpen = false"
            class="nav-item"
            :class="{ 'nav-item--active': isActiveRoute(item.href) }"
          >
            <el-icon :size="18" class="nav-icon"><component :is="item.icon" /></el-icon>
            <span>{{ item.name }}</span>
          </router-link>
        </nav>
        <div class="sidebar-footer">v1.0.0 · FC25 Tracker</div>
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
            <div class="logo-title">FC25 Tracker</div>
            <div class="logo-sub">Esport Score System</div>
          </div>
        </div>
        <nav class="sidebar-nav">
          <router-link
            v-for="item in navigation"
            :key="item.name"
            :to="item.href"
            class="nav-item"
            :class="{ 'nav-item--active': isActiveRoute(item.href) }"
          >
            <el-icon :size="18" class="nav-icon"><component :is="item.icon" /></el-icon>
            <span>{{ item.name }}</span>
          </router-link>
        </nav>
        <div class="sidebar-footer">v1.0.0 · © 2024 FC25 Tracker</div>
      </div>
    </aside>

    <!-- Main content -->
    <div class="main-area">
      <!-- Mobile topbar -->
      <header class="mobile-topbar">
        <div class="flex items-center gap-3">
          <button class="mobile-menu-btn" @click="mobileMenuOpen = true">
            <el-icon :size="20"><Menu /></el-icon>
          </button>
          <span class="mobile-page-title">{{ currentPageName }}</span>
        </div>
        <div class="logo-icon logo-icon--sm">
          <el-icon color="white" :size="16"><Trophy /></el-icon>
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
import { Menu, Trophy, HomeFilled, UserFilled, TrendCharts, DocumentCopy, Wallet, Setting, Grid } from '@element-plus/icons-vue'

const route = useRoute()
const mobileMenuOpen = ref(false)

const navigation = [
  { name: 'Dashboard', href: '/', icon: HomeFilled },
  { name: 'Players', href: '/users', icon: UserFilled },
  { name: 'Matches', href: '/matches', icon: TrendCharts },
  { name: 'Tournaments', href: '/tournaments', icon: Grid },
  { name: 'Settlements', href: '/settlements', icon: DocumentCopy },
  { name: 'Fund', href: '/fund', icon: Wallet },
  { name: 'Settings', href: '/settings', icon: Setting },
]

const isActiveRoute = (href: string): boolean =>
  href === '/' ? route.path === '/' : route.path.startsWith(href)

const currentPageName = computed(() =>
  navigation.find(item => isActiveRoute(item.href))?.name || 'FC25 Tracker'
)
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
  background: var(--sidebar-hover);
  color: #e2e8f0;
}

.nav-item--active {
  background: var(--sidebar-active);
  color: var(--sidebar-text-active);
  box-shadow: 0 2px 8px rgba(29, 78, 216, 0.3);
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

/* ── Main area ── */
.main-area {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;
  overflow: hidden;
}

/* ── Mobile topbar ── */
.mobile-topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: var(--surface-card);
  border-bottom: 1px solid var(--border-subtle);
  box-shadow: 0 1px 3px rgba(0,0,0,0.06);
  flex-shrink: 0;
}

@media (min-width: 1024px) {
  .mobile-topbar {
    display: none;
  }
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
}

.mobile-menu-btn:hover {
  background: var(--surface-subtle);
}

.mobile-page-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

/* ── Scrollable page ── */
.page-scroll {
  flex: 1;
  overflow-y: auto;
}
</style>
