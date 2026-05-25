<template>
  <div class="hl-list">
    <div v-if="items.length === 0" class="hl-empty">{{ emptyText }}</div>
    <div v-else class="hl-items">
      <div
        v-for="(item, i) in items"
        :key="i"
        class="hl-card"
        :style="{ '--c': color, '--cbg': colorBg, animationDelay: i * 55 + 'ms' }"
      >
        <span class="hl-emoji">{{ item.emoji }}</span>
        <span class="hl-msg">{{ item.message }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Highlight } from '@/types/highlight'

defineProps<{
  items: Highlight[]
  color: string
  colorBg: string
  emptyText: string
}>()
</script>

<style scoped>
.hl-list { min-height: 48px; }

.hl-empty {
  padding: 14px 0 4px;
  text-align: center;
  font-size: 12px;
  color: var(--text-muted);
}

.hl-items { display: flex; flex-direction: column; gap: 3px; }

.hl-card {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 7px 9px;
  border-radius: 8px;
  border-left: 3px solid var(--c);
  background: var(--cbg);
  animation: hl-in 0.22s ease both;
}

@keyframes hl-in {
  from { opacity: 0; transform: translateY(4px); }
  to   { opacity: 1; transform: translateY(0); }
}

.hl-emoji {
  font-size: 1rem;
  flex-shrink: 0;
  margin-top: 1px;
  line-height: 1.4;
}

.hl-msg {
  font-size: 12.5px;
  line-height: 1.45;
  color: var(--text-primary);
}
</style>
