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
        <div class="wc-lb-name">{{ entry.name }}</div>
        <div class="wc-lb-stats">
          <span class="wc-lb-stat">{{ entry.correct }}W / {{ entry.total_predictions }}T</span>
        </div>
        <div class="wc-lb-profit" :class="entry.net_points >= 0 ? 'wc-profit--pos' : 'wc-profit--neg'">
          {{ entry.net_points >= 0 ? '+' : '' }}{{ entry.net_points }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { WcLeaderboardEntry } from '@/types/wc'

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
}

.wc-lb-stats {
  font-size: 12px;
  color: var(--text-muted);
  flex-shrink: 0;
}

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
