<template>
  <div>
    <div v-if="entries.length === 0" class="empty-state">
      <div class="empty-state-title">Chưa có kèo phụ nào.</div>
    </div>
    <div v-else class="wc-cb-history-list">
      <div v-for="entry in entries" :key="entry.id" class="wc-cb-history-row">
        <div class="wc-cb-history-main">
          <div class="wc-cb-history-match">
            {{ entry.home_team }} vs {{ entry.away_team }}
            <span class="wc-cb-history-date">{{ formatDate(entry.match_date) }}</span>
          </div>
          <div class="wc-cb-history-bet">
            <span class="wc-cb-history-type">Kèo phụ</span>
            <span class="wc-cb-history-title">
              {{ entry.bet_title }}
              <span v-if="entry.bet_line != null" class="wc-cb-history-line">@{{ entry.bet_line }}</span>
            </span>
          </div>
          <div class="wc-cb-history-detail">
            <span class="wc-cb-history-option">{{ entry.option_label }}</span>
            <span class="wc-cb-history-stake">{{ entry.stake }} × {{ entry.odds_snapshot.toFixed(2) }}</span>
          </div>
        </div>
        <div class="wc-cb-history-result">
          <span v-if="entry.status === 'pending'" class="wc-result-badge wc-result--pending">Chờ kết quả</span>
          <span v-else-if="entry.status === 'won'" class="wc-result-badge wc-result--win">
            +{{ ((entry.payout ?? 0) - entry.stake).toFixed(2) }} Thắng
          </span>
          <span v-else-if="entry.status === 'lost'" class="wc-result-badge wc-result--lose">
            -{{ entry.stake }} Thua
          </span>
          <span v-else class="wc-result-badge wc-result--push">±0 Huỷ</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { WcCustomBetEntryHistory } from '@/types/wc'

defineProps<{ entries: WcCustomBetEntryHistory[] }>()

function formatDate(d: string) {
  return new Date(d).toLocaleDateString('vi-VN', { day: '2-digit', month: '2-digit' })
}
</script>

<style scoped>
.wc-cb-history-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.wc-cb-history-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 14px;
  background: var(--surface-card);
  border: 1px solid var(--border-default);
  border-radius: 10px;
}

.wc-cb-history-main {
  display: flex;
  flex-direction: column;
  gap: 3px;
  flex: 1;
  min-width: 0;
}

.wc-cb-history-match {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  display: flex;
  align-items: center;
  gap: 8px;
}

.wc-cb-history-date {
  font-size: 11px;
  color: var(--text-muted);
  font-weight: 400;
}

.wc-cb-history-bet {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
}

.wc-cb-history-type {
  background: var(--el-color-warning-light-8);
  color: var(--el-color-warning);
  border-radius: 4px;
  padding: 1px 6px;
  font-weight: 600;
  font-size: 11px;
  white-space: nowrap;
}

.wc-cb-history-title {
  color: var(--text-secondary);
  display: flex;
  align-items: center;
  gap: 4px;
}

.wc-cb-history-line {
  font-size: 11px;
  font-weight: 700;
  color: var(--el-color-warning);
  background: rgba(var(--el-color-warning-rgb, 230, 162, 60), 0.1);
  padding: 1px 5px;
  border-radius: 4px;
}

.wc-cb-history-detail {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--text-muted);
}

.wc-cb-history-option {
  font-weight: 500;
  color: var(--text-secondary);
}

.wc-cb-history-stake {
  color: var(--text-muted);
}

.wc-cb-history-result {
  flex-shrink: 0;
}

.wc-result-badge {
  font-size: 12px;
  font-weight: 700;
  padding: 3px 10px;
  border-radius: 20px;
  white-space: nowrap;
}

.wc-result--pending {
  background: var(--el-color-info-light-9);
  color: var(--el-color-info);
}

.wc-result--win {
  background: rgba(22, 163, 74, 0.1);
  color: #16a34a;
}

.wc-result--lose {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.wc-result--push {
  background: var(--el-color-warning-light-9);
  color: var(--el-color-warning);
}
</style>
