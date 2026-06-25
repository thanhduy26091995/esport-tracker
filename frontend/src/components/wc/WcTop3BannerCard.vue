<template>
  <div
    class="banner-card"
    :class="[`banner-card--rank${rank}`, { 'banner-card--me': isCurrentUser }]"
    :aria-label="`${t('wc.top3Banner.rank')} ${rank}: ${entry.name}, ${sign}${fmtPts(entry.net_points)} ${t('wc.top3Banner.pts')}`"
  >
    <span class="banner-card-medal">{{ MEDALS[rank] }}</span>
    <img
      :src="entry.avatar_url || DEFAULT_AVATAR"
      class="banner-card-avatar"
      :alt="entry.name"
      @error="(e: Event) => ((e.target as HTMLImageElement).src = DEFAULT_AVATAR)"
    />
    <span class="banner-card-name">{{ displayName }}</span>
    <span class="banner-card-pts" :class="entry.net_points >= 0 ? 'pts--pos' : 'pts--neg'">
      {{ sign }}{{ fmtPts(entry.net_points) }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { WcLeaderboardEntry } from '@/types/wc'

const { t } = useI18n()

const props = defineProps<{
  entry: WcLeaderboardEntry
  rank: 1 | 2 | 3
  isCurrentUser: boolean
}>()

const MEDALS: Record<number, string> = { 1: '🥇', 2: '🥈', 3: '🥉' }

const DEFAULT_AVATAR = 'data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 32 32"%3E%3Ccircle cx="16" cy="16" r="16" fill="%23374151"/%3E%3Ccircle cx="16" cy="13" r="6" fill="%236b7280"/%3E%3Cellipse cx="16" cy="29" rx="9" ry="7" fill="%236b7280"/%3E%3C/svg%3E'

const displayName = computed(() =>
  props.entry.name.length > 16 ? props.entry.name.slice(0, 16) + '…' : props.entry.name
)

const sign = computed(() => props.entry.net_points >= 0 ? '+' : '')

function fmtPts(v: number): string {
  return parseFloat(v.toFixed(2)).toString()
}
</script>

<style scoped>
.banner-card {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  padding: 6px 12px 6px 8px;
  background: var(--surface-card);
  border: 1px solid rgba(217, 119, 6, 0.25);
  border-radius: 20px;
  white-space: nowrap;
  flex-shrink: 0;
  transition: border-color 0.15s;
}

.banner-card--rank1 { border-color: rgba(234, 179, 8, 0.45); }
.banner-card--rank2 { border-color: rgba(148, 163, 184, 0.45); }
.banner-card--rank3 { border-color: rgba(180, 120, 60, 0.4); }

.banner-card--me {
  box-shadow: 0 0 0 2px #d97706;
}

.banner-card-medal {
  font-size: 16px;
  line-height: 1;
  flex-shrink: 0;
}

.banner-card-avatar {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  object-fit: cover;
  flex-shrink: 0;
  background: #374151;
}

.banner-card-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  max-width: 110px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.banner-card-pts {
  font-size: 13px;
  font-weight: 800;
  font-variant-numeric: tabular-nums;
  flex-shrink: 0;
}

.pts--pos { color: #16a34a; }
.pts--neg { color: #ef4444; }

@media (max-width: 640px) {
  .banner-card {
    padding: 5px 10px 5px 7px;
    gap: 5px;
  }
  .banner-card-name {
    font-size: 11px;
    max-width: 80px;
  }
  .banner-card-pts {
    font-size: 11px;
  }
  .banner-card-avatar {
    width: 20px;
    height: 20px;
  }
}
</style>
