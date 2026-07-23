<template>
  <div class="page-wrapper">
    <div class="page-container">
      <div class="page-header">
        <div class="page-header-left">
          <h1 class="page-title">{{ tournamentTitle }}</h1>
          <p class="page-subtitle">{{ t('wc.adminPanel') }}</p>
        </div>
        <div class="wc-user-header">
          <div class="wc-user-info">
            <span class="wc-user-name">{{ authStore.userName }}</span>
            <span class="wc-admin-badge">Admin</span>
          </div>
          <el-button size="small" text @click="handleLogout">{{ t('wc.logout') }}</el-button>
        </div>
      </div>

      <WcAdminPanel />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useWcAuthStore } from '@/stores/wcAuthStore'
import WcAdminPanel from '@/components/wc/WcAdminPanel.vue'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const authStore = useWcAuthStore()

const isAc = computed(() => route.meta?.tournamentType === 'asean_cup')
const tournamentTitle = computed(() => isAc.value ? '🏆 ASEAN Cup 2026' : '🏆 World Cup 2026')

function handleLogout() {
  authStore.logout()
  router.push(isAc.value ? '/asean-cup/login' : '/world-cup/login')
}
</script>

<style scoped>
.wc-user-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

@media (max-width: 540px) {
  .wc-user-header {
    justify-content: space-between;
    width: 100%;
  }
}

.wc-user-info {
  display: flex;
  align-items: center;
  gap: 6px;
}

.wc-user-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.wc-admin-badge {
  font-size: 11px;
  font-weight: 700;
  background: rgba(239, 68, 68, 0.12);
  color: #ef4444;
  padding: 2px 8px;
  border-radius: 6px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}
</style>
