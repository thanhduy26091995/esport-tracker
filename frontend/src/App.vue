<template>
  <MainLayout />
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import MainLayout from './layouts/MainLayout.vue'
import { useUserStore } from './stores/userStore'
import { useGlobalTheme } from './composables/useGlobalTheme'

const userStore = useUserStore()
const isSocSite = import.meta.env.VITE_SITE === 'soc'

// On soc build: skip the FC25 user fetch and global club theme (no FC25 players on soc)
if (!isSocSite) {
  useGlobalTheme(computed(() => userStore.users))
  onMounted(() => userStore.fetchUsers())
}
</script>
