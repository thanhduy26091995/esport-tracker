<template>
  <MainLayout />
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import MainLayout from './layouts/MainLayout.vue'
import { useUserStore } from './stores/userStore'
import { useGlobalTheme } from './composables/useGlobalTheme'

const userStore = useUserStore()
useGlobalTheme(computed(() => userStore.users))

// Fetch users immediately so the global theme reflects rank #1's club from the first page load.
onMounted(() => userStore.fetchUsers())
</script>
