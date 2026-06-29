<template>
  <div v-if="!siteAccessStore.checked && !isAdminRoute" class="app-loading" />
  <SiteAccessGate v-else-if="!siteAccessStore.isGranted && !isAdminRoute" />
  <MainLayout v-else />
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import MainLayout from './layouts/MainLayout.vue'
import SiteAccessGate from './components/SiteAccessGate.vue'
import { useUserStore } from './stores/userStore'
import { useGlobalTheme } from './composables/useGlobalTheme'
import { useSiteAccessStore } from './stores/siteAccessStore'

const userStore = useUserStore()
const siteAccessStore = useSiteAccessStore()
const route = useRoute()
const isSocSite = import.meta.env.VITE_SITE === 'soc'

const isAdminRoute = computed(() =>
  route.path.startsWith('/world-cup/admin') || route.path === '/world-cup/login'
)

onMounted(() => siteAccessStore.init())

// On soc build: skip the FC25 user fetch and global club theme (no FC25 players on soc)
if (!isSocSite) {
  useGlobalTheme(computed(() => userStore.users))
  onMounted(() => userStore.fetchUsers())
}
</script>

<style scoped>
.app-loading {
  position: fixed;
  inset: 0;
  background: #fff;
}
</style>
