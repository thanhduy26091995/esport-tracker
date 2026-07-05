<template>
  <div class="wc-leaderboard">
    <div v-if="visibleEntries.length === 0" class="empty-state">
      <div class="empty-state-title">{{ t('wc.noLeaderboard') }}</div>
    </div>
    <div v-else class="wc-lb-list">
      <div
        v-for="(entry, index) in visibleEntries"
        :key="entry.wc_user_id"
        class="wc-lb-row"
        :class="[{ 'wc-lb-row--top3': index < 3 }, rankClass(index, 'row')]"
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
        <div class="wc-lb-name" :class="rankClass(index, 'name')">
          <span v-if="index === 0" class="wc-lb-crown" aria-hidden="true">👑</span>
          <span class="wc-lb-name-text">{{ entry.name }}</span>
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
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { WcLeaderboardEntry } from '@/types/wc'

const DEFAULT_AVATAR = 'data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 32 32"%3E%3Ccircle cx="16" cy="16" r="16" fill="%23374151"/%3E%3Ccircle cx="16" cy="13" r="6" fill="%236b7280"/%3E%3Cellipse cx="16" cy="29" rx="9" ry="7" fill="%236b7280"/%3E%3C/svg%3E'

function fmtPts(v: number): string {
  return parseFloat(v.toFixed(2)).toString()
}

// Per-rank modifier class for the top 3 (bots already excluded, so rank = position).
function rankClass(index: number, part: 'row' | 'name'): string {
  if (index > 2) return ''
  return `wc-lb-${part}--rank${index + 1}`
}

const { t } = useI18n()
const props = defineProps<{ entries: WcLeaderboardEntry[] }>()

// Bots are hidden from the leaderboard; ranking is computed after filtering
// so the visible top 3 are always real players.
const visibleEntries = computed(() => props.entries.filter((e) => !e.is_bot))
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

/* ── Per-rank row treatment (gold / silver / bronze) ── */
.wc-lb-row--rank1,
.wc-lb-row--rank2,
.wc-lb-row--rank3 {
  position: relative;
  overflow: hidden;
  border-width: 1px 1px 1px 0; /* left edge is the accent bar (inset box-shadow) */
}

/* Diagonal light sweep across the whole podium block */
.wc-lb-row--rank1::before,
.wc-lb-row--rank2::before,
.wc-lb-row--rank3::before {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(105deg, transparent 32%, rgba(255, 255, 255, 0.16) 48%, rgba(255, 255, 255, 0.16) 52%, transparent 68%);
  transform: translateX(-120%);
  animation: wc-lb-sheen 4.5s ease-in-out infinite;
  pointer-events: none;
  z-index: 0;
}
/* Keep row content above the sweep */
.wc-lb-row > * { position: relative; z-index: 1; }

.wc-lb-row--rank1 {
  border-color: rgba(234, 179, 8, 0.85);
  background: linear-gradient(90deg, rgba(234, 179, 8, 0.24), rgba(234, 179, 8, 0.05) 55%, var(--surface-card));
  animation: wc-lb-glow-gold 2.6s ease-in-out infinite;
}
.wc-lb-row--rank1::before { animation-duration: 3.5s; }

.wc-lb-row--rank2 {
  border-color: rgba(226, 232, 240, 0.9);
  background: linear-gradient(90deg, rgba(226, 232, 240, 0.22), rgba(226, 232, 240, 0.05) 55%, var(--surface-card));
  animation: wc-lb-glow-silver 3s ease-in-out infinite;
}
.wc-lb-row--rank2::before { animation-duration: 4.3s; }

.wc-lb-row--rank3 {
  border-color: rgba(210, 140, 80, 0.85);
  background: linear-gradient(90deg, rgba(210, 140, 80, 0.2), rgba(210, 140, 80, 0.05) 55%, var(--surface-card));
  animation: wc-lb-glow-bronze 3.4s ease-in-out infinite;
}
.wc-lb-row--rank3::before { animation-duration: 5.2s; }

@keyframes wc-lb-sheen {
  0%       { transform: translateX(-120%); }
  55%,100% { transform: translateX(120%); }
}
@keyframes wc-lb-glow-gold {
  0%, 100% { box-shadow: inset 4px 0 0 #f4c025, 0 0 12px rgba(234, 179, 8, 0.3); }
  50%      { box-shadow: inset 4px 0 0 #f4c025, 0 0 28px rgba(234, 179, 8, 0.62); }
}
@keyframes wc-lb-glow-silver {
  0%, 100% { box-shadow: inset 4px 0 0 #e2e8f0, 0 0 12px rgba(226, 232, 240, 0.28); }
  50%      { box-shadow: inset 4px 0 0 #e2e8f0, 0 0 26px rgba(226, 232, 240, 0.58); }
}
@keyframes wc-lb-glow-bronze {
  0%, 100% { box-shadow: inset 4px 0 0 #d28c50, 0 0 10px rgba(210, 140, 80, 0.26); }
  50%      { box-shadow: inset 4px 0 0 #d28c50, 0 0 22px rgba(210, 140, 80, 0.52); }
}

/* Metallic avatar ring for the podium */
.wc-lb-row--rank1 .wc-lb-avatar { box-shadow: 0 0 0 2px #f4c025, 0 0 9px rgba(234, 179, 8, 0.65); }
.wc-lb-row--rank2 .wc-lb-avatar { box-shadow: 0 0 0 2px #e2e8f0, 0 0 9px rgba(226, 232, 240, 0.55); }
.wc-lb-row--rank3 .wc-lb-avatar { box-shadow: 0 0 0 2px #d28c50, 0 0 9px rgba(210, 140, 80, 0.55); }

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
  min-width: 0;
}

.wc-lb-name-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* ── Champion crown (rank 1 only) ── */
.wc-lb-crown {
  font-size: 15px;
  line-height: 1;
  flex-shrink: 0;
  animation: wc-lb-crown-bob 2s ease-in-out infinite;
  filter: drop-shadow(0 0 4px rgba(234, 179, 8, 0.6));
}

@keyframes wc-lb-crown-bob {
  0%, 100% { transform: translateY(0) rotate(-6deg); }
  50%      { transform: translateY(-2px) rotate(6deg); }
}

/* ── Metallic gradient name text, intensity fading rank1 → rank3 ── */
.wc-lb-name--rank1 .wc-lb-name-text,
.wc-lb-name--rank2 .wc-lb-name-text,
.wc-lb-name--rank3 .wc-lb-name-text {
  font-weight: 800;
  background-size: 200% auto;
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  color: transparent;
  animation: wc-lb-shimmer linear infinite;
}

.wc-lb-name--rank1 .wc-lb-name-text {
  background-image: linear-gradient(100deg, #fde68a 0%, #f5b301 25%, #fffbe6 50%, #f5b301 75%, #fde68a 100%);
  filter: drop-shadow(0 0 7px rgba(245, 179, 1, 0.7));
  animation-duration: 2.5s;
}
/* Rank 2 stays bright platinum end-to-end so it reads clearly on dark bg */
.wc-lb-name--rank2 .wc-lb-name-text {
  background-image: linear-gradient(100deg, #ffffff 0%, #dbe4ee 30%, #ffffff 50%, #dbe4ee 70%, #ffffff 100%);
  filter: drop-shadow(0 0 6px rgba(219, 228, 238, 0.75));
  animation-duration: 3.6s;
}
.wc-lb-name--rank3 .wc-lb-name-text {
  background-image: linear-gradient(100deg, #f2cfa8 0%, #d98b4a 35%, #ffe8d0 50%, #d98b4a 65%, #f2cfa8 100%);
  filter: drop-shadow(0 0 5px rgba(217, 139, 74, 0.6));
  animation-duration: 5.5s;
}

@keyframes wc-lb-shimmer {
  from { background-position: 200% center; }
  to   { background-position: -200% center; }
}

@media (prefers-reduced-motion: reduce) {
  .wc-lb-crown,
  .wc-lb-name--rank1 .wc-lb-name-text,
  .wc-lb-name--rank2 .wc-lb-name-text,
  .wc-lb-name--rank3 .wc-lb-name-text,
  .wc-lb-row--rank1,
  .wc-lb-row--rank2,
  .wc-lb-row--rank3 {
    animation: none;
  }
  .wc-lb-row--rank1::before,
  .wc-lb-row--rank2::before,
  .wc-lb-row--rank3::before {
    display: none;
  }
  /* Keep the accent bar + a steady glow so the podium stays prominent */
  .wc-lb-row--rank1 { box-shadow: inset 4px 0 0 #f4c025, 0 0 16px rgba(234, 179, 8, 0.42); }
  .wc-lb-row--rank2 { box-shadow: inset 4px 0 0 #e2e8f0, 0 0 16px rgba(226, 232, 240, 0.4); }
  .wc-lb-row--rank3 { box-shadow: inset 4px 0 0 #d28c50, 0 0 14px rgba(210, 140, 80, 0.36); }
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
