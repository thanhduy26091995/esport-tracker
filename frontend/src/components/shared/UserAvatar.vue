<template>
  <div class="user-avatar" :class="[`user-avatar--${size}`, { 'user-avatar--has-img': resolvedUrl }]">
    <img v-if="resolvedUrl" :src="resolvedUrl" :alt="name" class="user-avatar__img" />
    <span v-else class="user-avatar__initials">{{ initials }}</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  avatarUrl?: string | null
  name: string
  size?: 'xs' | 'sm' | 'md' | 'lg'
}>(), { size: 'md' })

const initials = computed(() =>
  props.name
    .split(' ')
    .map(w => w[0])
    .join('')
    .toUpperCase()
    .slice(0, 2)
)

// Resolve relative paths (e.g. /uploads/avatars/x.jpg) to the backend origin.
// Absolute URLs (http/https) are passed through unchanged.
const resolvedUrl = computed(() => {
  const url = props.avatarUrl
  if (!url) return null
  if (url.startsWith('http')) return url
  const base = (import.meta.env.VITE_API_BASE_URL as string) || 'http://localhost:8080/api/v1'
  const origin = base.replace(/\/api\/v1\/?$/, '')
  return origin + url
})
</script>

<style scoped>
.user-avatar {
  position: relative;
  border-radius: 50%;
  overflow: hidden;
  flex-shrink: 0;
  /* Centers initials text; img uses absolute positioning so flex doesn't affect it */
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  /* Neutral initials style — same as old .player-avatar */
  background: linear-gradient(135deg, #dbeafe, #bfdbfe);
  color: #1d4ed8;
  /* Hardware-accelerate the circle clip to prevent subpixel blur */
  transform: translateZ(0);
  -webkit-transform: translateZ(0);
}

.user-avatar--xs { width: 24px; height: 24px; font-size: 10px; }
.user-avatar--sm { width: 32px; height: 32px; font-size: 12px; }
.user-avatar--md { width: 40px; height: 40px; font-size: 14px; }
.user-avatar--lg { width: 56px; height: 56px; font-size: 18px; }

.user-avatar__img {
  /* Absolute fill — avoids flex-child height:100% quirks */
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.user-avatar__initials {
  line-height: 1;
  user-select: none;
}
</style>
