<template>
  <div class="card hfp">
    <div class="hfp-header">
      <div class="hfp-title">
        <span class="hfp-bolt">⚡</span>
        <span>Highlights</span>
      </div>
      <span v-if="store.generatedAt" class="hfp-ts">{{ formatTime(store.generatedAt) }}</span>
    </div>

    <div v-if="store.loading" class="hfp-grid">
      <div v-for="i in 4" :key="i" class="hfp-section">
        <el-skeleton :rows="2" animated />
      </div>
    </div>

    <div v-else class="hfp-grid">
      <div v-for="s in sections" :key="s.key" class="hfp-section">
        <div class="hfp-sec-head">
          <div class="hfp-sec-label">
            <span class="hfp-sec-dot" :style="{ background: s.color }"></span>
            <span class="hfp-sec-emoji">{{ s.emoji }}</span>
            <span class="hfp-sec-name">{{ s.label }}</span>
          </div>
          <span v-if="s.items.length" class="hfp-sec-badge" :style="{ background: s.color }">
            {{ s.items.length }}
          </span>
        </div>
        <HighlightList
          :items="s.items"
          :color="s.color"
          :color-bg="s.colorBg"
          :empty-text="s.emptyText"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useHighlightStore } from '@/stores/highlightStore'
import HighlightList from './HighlightList.vue'

const store = useHighlightStore()

const sections = computed(() => [
  {
    key: 'trending',
    label: 'Trending',
    emoji: '🔥',
    color: '#f97316',
    colorBg: 'rgba(249,115,22,0.07)',
    items: store.trending,
    emptyText: 'Chưa có highlight nào',
  },
  {
    key: 'daily_recap',
    label: 'Hôm nay',
    emoji: '📋',
    color: '#0ea5e9',
    colorBg: 'rgba(14,165,233,0.07)',
    items: store.dailyRecap,
    emptyText: 'Chưa có highlight hôm nay',
  },
  {
    key: 'competitive',
    label: 'BXH',
    emoji: '⚔️',
    color: '#8b5cf6',
    colorBg: 'rgba(139,92,246,0.07)',
    items: store.competitive,
    emptyText: 'Chưa có thay đổi BXH',
  },
  {
    key: 'social',
    label: 'Social',
    emoji: '😄',
    color: '#10b981',
    colorBg: 'rgba(16,185,129,0.07)',
    items: store.social,
    emptyText: 'Chưa có gì thú vị',
  },
])

function formatTime(iso: string): string {
  return new Date(iso).toLocaleTimeString('vi-VN', { hour: '2-digit', minute: '2-digit' })
}
</script>

<style scoped>
/* Clip grid gap at rounded corners */
.hfp { overflow: hidden; }

.hfp-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 11px 16px 10px;
  border-bottom: 1px solid var(--border-default);
}

.hfp-title {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 14px;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: -0.15px;
}

.hfp-bolt { font-size: 15px; }

.hfp-ts {
  font-size: 11px;
  color: var(--text-muted);
}

/* The 1px gap between sections acts as the divider line */
.hfp-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1px;
  background: var(--border-default);
}

@media (min-width: 640px) {
  .hfp-grid { grid-template-columns: repeat(2, 1fr); }
}

.hfp-section {
  background: var(--surface-card);
  padding: 12px 14px 14px;
}

.hfp-sec-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.hfp-sec-label {
  display: flex;
  align-items: center;
  gap: 5px;
}

.hfp-sec-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}

.hfp-sec-emoji { font-size: 12px; }

.hfp-sec-name {
  font-size: 10.5px;
  font-weight: 700;
  letter-spacing: 0.65px;
  text-transform: uppercase;
  color: var(--text-muted);
}

.hfp-sec-badge {
  font-size: 10px;
  font-weight: 700;
  color: #fff;
  padding: 1px 7px;
  border-radius: 99px;
  line-height: 1.6;
}
</style>
