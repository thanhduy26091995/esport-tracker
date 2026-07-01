<template>
  <div class="wc-leaderboard">
    <div v-if="entries.length === 0" class="empty-state">
      <div class="empty-state-title">{{ t('wc.noLeaderboard') }}</div>
    </div>
    <div v-else class="wc-lb-list">
      <div
        v-for="(entry, index) in entries"
        :key="entry.wc_user_id"
        class="wc-lb-row"
        :class="{ 'wc-lb-row--top3': index < 3 }"
      >
        <div class="wc-lb-rank">
          <span v-if="index === 0" class="wc-lb-medal wc-lb-medal--gold">🥇</span>
          <span v-else-if="index === 1" class="wc-lb-medal wc-lb-medal--silver">🥈</span>
          <span v-else-if="index === 2" class="wc-lb-medal wc-lb-medal--bronze">🥉</span>
          <span v-else class="wc-lb-rank-num">{{ index + 1 }}</span>
        </div>
        <img
          :src="entry.avatar_url || DEFAULT_AVATAR"
          class="wc-lb-avatar"
          :alt="entry.name"
          @error="(e: Event) => ((e.target as HTMLImageElement).src = DEFAULT_AVATAR)"
        />
        <div class="wc-lb-name">
          {{ entry.name }}
          <span v-if="entry.is_bot" class="wc-lb-bot-badge">Bot</span>
        </div>
        <div class="wc-lb-stats">
          <span class="wc-lb-stat wc-stat--win">{{ entry.correct }}W</span>
          <span v-if="entry.win_half > 0" class="wc-lb-stat wc-stat--winhalf">{{ entry.win_half }}½W</span>
          <span v-if="entry.lose_half > 0" class="wc-lb-stat wc-stat--losehalf">{{ entry.lose_half }}½L</span>
          <span class="wc-lb-stat wc-stat--loss">{{ entry.incorrect }}L</span>
          <span class="wc-lb-stat wc-stat--total">/ {{ entry.total_predictions }}</span>
        </div>
        <div class="wc-lb-profit" :class="entry.net_points >= 0 ? 'wc-profit--pos' : 'wc-profit--neg'">
          {{ entry.net_points >= 0 ? '+' : '' }}{{ fmtPts(entry.net_points) }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { WcLeaderboardEntry } from '@/types/wc'

const DEFAULT_AVATAR = 'data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 32 32"%3E%3Ccircle cx="16" cy="16" r="16" fill="%23374151"/%3E%3Ccircle cx="16" cy="13" r="6" fill="%236b7280"/%3E%3Cellipse cx="16" cy="29" rx="9" ry="7" fill="%236b7280"/%3E%3C/svg%3E'

function fmtPts(v: number): string {
  return parseFloat(v.toFixed(2)).toString()
}

const { t } = useI18n()
defineProps<{ entries: WcLeaderboardEntry[] }>()
</script>

<style scoped>
.wc-leaderboard {
  width: 100%;
}

.wc-lb-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.wc-lb-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  background: var(--surface-card);
  border: 1px solid var(--border-default);
  border-radius: 10px;
  transition: border-color 0.15s;
}

.wc-lb-row--top3 {
  border-color: rgba(217, 119, 6, 0.3);
  background: linear-gradient(90deg, rgba(217, 119, 6, 0.04), var(--surface-card));
}

.wc-lb-rank {
  width: 32px;
  flex-shrink: 0;
  text-align: center;
}

.wc-lb-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  object-fit: cover;
  flex-shrink: 0;
  background: #374151;
}

.wc-lb-medal {
  font-size: 20px;
  line-height: 1;
}

.wc-lb-rank-num {
  font-size: 13px;
  font-weight: 700;
  color: var(--text-muted);
}

.wc-lb-name {
  flex: 1;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  display: flex;
  align-items: center;
  gap: 6px;
}

.wc-lb-bot-badge {
  font-size: 10px;
  font-weight: 700;
  color: #6b7280;
  background: rgba(107, 114, 128, 0.12);
  border: 1px solid rgba(107, 114, 128, 0.3);
  border-radius: 4px;
  padding: 1px 5px;
  letter-spacing: 0.04em;
  flex-shrink: 0;
}

.wc-lb-stats {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  font-weight: 600;
  flex-shrink: 0;
}
.wc-lb-stat { font-variant-numeric: tabular-nums; }
.wc-stat--win      { color: #16a34a; }
.wc-stat--winhalf  { color: #65a30d; }
.wc-stat--losehalf { color: #ea580c; }
.wc-stat--loss     { color: #ef4444; }
.wc-stat--total    { color: var(--text-muted); font-weight: 400; }

.wc-lb-profit {
  font-size: 16px;
  font-weight: 800;
  tabular-nums: true;
  flex-shrink: 0;
  min-width: 60px;
  text-align: right;
}

.wc-profit--pos {
  color: #16a34a;
}

.wc-profit--neg {
  color: #ef4444;
}
</style>
