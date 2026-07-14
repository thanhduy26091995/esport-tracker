<template>
  <div
    v-if="bottom3.length >= 1 || wcStore.leaderboardLoading"
    class="wc-top3-banner"
    role="region"
    :aria-label="t('wc.top3Banner.ariaLabel')"
  >
    <span class="banner-label">{{ t('wc.top3Banner.label') }}</span>

    <!-- Skeleton while first load -->
    <div v-if="wcStore.leaderboardLoading && bottom3.length === 0" class="banner-skeleton">
      <div v-for="n in 3" :key="n" class="banner-skeleton-card" />
    </div>

    <!-- Live marquee -->
    <div
      v-else
      class="banner-track-wrapper"
      @mouseenter="paused = true"
      @mouseleave="paused = false"
    >
      <div class="banner-track" :class="{ 'banner-track--paused': paused }">
        <WcTop3BannerCard
          v-for="(entry, i) in displayEntries"
          :key="`${entry.wc_user_id}-${i}`"
          :entry="entry"
          :rank="((i % bottom3.length) + 1) as 1 | 2 | 3"
          :isCurrentUser="entry.wc_user_id === currentUserId"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useWcStore } from '@/stores/wcStore'
import { useWcAuthStore } from '@/stores/wcAuthStore'
import WcTop3BannerCard from './WcTop3BannerCard.vue'

const { t } = useI18n()
const wcStore = useWcStore()
const wcAuthStore = useWcAuthStore()

const paused = ref(false)

// Header shows the REVERSE of the leaderboard: lowest / most-negative points first.
// (The leaderboard page itself keeps the normal top-first order.)
const bottom3 = computed(() =>
  [...wcStore.leaderboard.filter(e => !e.is_bot)]
    .sort((a, b) => a.net_points - b.net_points)
    .slice(0, 3),
)
const displayEntries = computed(() => [...bottom3.value, ...bottom3.value])
const currentUserId = computed(() => wcAuthStore.user?.id ?? null)

let refreshTimer: ReturnType<typeof setInterval> | null = null

onMounted(() => {
  if (wcStore.leaderboard.length === 0) wcStore.fetchLeaderboard()
  refreshTimer = setInterval(() => wcStore.fetchLeaderboard(), 5 * 60 * 1000)
})

onUnmounted(() => {
  if (refreshTimer !== null) clearInterval(refreshTimer)
})
</script>

<style scoped>
.wc-top3-banner {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  height: 48px;
  padding: 0 16px;
  background: linear-gradient(90deg, rgba(217, 119, 6, 0.08) 0%, var(--surface-card) 100%);
  border-bottom: 1px solid rgba(217, 119, 6, 0.2);
  overflow: hidden;
  flex-shrink: 0;
  box-sizing: border-box;
}

.banner-label {
  font-size: 11px;
  font-weight: 700;
  color: #d97706;
  white-space: nowrap;
  flex-shrink: 0;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

/* ── Skeleton ── */
.banner-skeleton {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 10px;
  overflow: hidden;
}

.banner-skeleton-card {
  width: 160px;
  height: 30px;
  border-radius: 20px;
  background: linear-gradient(90deg, var(--surface-subtle, #1e293b) 25%, rgba(255,255,255,0.06) 50%, var(--surface-subtle, #1e293b) 75%);
  background-size: 200% 100%;
  animation: shimmer 1.4s ease-in-out infinite;
  flex-shrink: 0;
}

@keyframes shimmer {
  0%   { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

/* ── Live marquee ── */
.banner-track-wrapper {
  flex: 1;
  overflow: hidden;
  position: relative;
  min-width: 0;
  /* fade edges */
  -webkit-mask-image: linear-gradient(90deg, transparent 0%, black 4%, black 96%, transparent 100%);
  mask-image: linear-gradient(90deg, transparent 0%, black 4%, black 96%, transparent 100%);
}

.banner-track {
  display: inline-flex;
  gap: 10px;
  animation: marquee 24s linear infinite;
}

.banner-track--paused {
  animation-play-state: paused;
}

@keyframes marquee {
  from { transform: translateX(0); }
  to   { transform: translateX(-50%); }
}

@media (prefers-reduced-motion: reduce) {
  .banner-track {
    animation: none;
  }
  .banner-skeleton-card {
    animation: none;
  }
  .banner-track-wrapper {
    overflow-x: auto;
    -webkit-mask-image: none;
    mask-image: none;
  }
}

@media (max-width: 640px) {
  .wc-top3-banner {
    height: 44px;
    padding: 0 10px;
    gap: 6px;
  }
  .banner-label {
    font-size: 10px;
  }
  .banner-skeleton-card {
    width: 120px;
    height: 26px;
  }
}
</style>
