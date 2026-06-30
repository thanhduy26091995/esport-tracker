<template>
  <div>
    <div v-if="items.length === 0" class="empty-state">
      <div class="empty-state-title">{{ t('wc.noBets') }}</div>
    </div>
    <div v-else class="wc-history-list">
      <div v-for="item in items" :key="item.id" class="wc-history-row" :class="{ 'wc-history-row--cancelled': !!item.cancelled_at }">
        <div class="wc-history-main">
          <div class="wc-history-match-info">
            <span class="wc-history-teams">{{ item.home_team }} vs {{ item.away_team }}</span>
            <span class="wc-history-date">{{ formatDate(item.created_at) }}</span>
          </div>
          <div class="wc-history-details">
            <template v-if="item.kind === 'regular'">
              <span class="wc-history-type">{{ betTypeLabel(item.bet_type ?? '') }}</span>
              <span class="wc-history-choice">
                <template v-if="item.bet_type === 'exact_score'">
                  {{ item.predicted_home_score }}–{{ item.predicted_away_score }}
                </template>
                <template v-else>{{ item.bet_choice }}</template>
              </span>
            </template>
            <template v-else>
              <span class="wc-history-type">{{ t('wc.betTypeCustom') }}</span>
              <span class="wc-history-choice">{{ item.bet_title }}</span>
              <span class="wc-history-option">{{ item.option_label }}</span>
            </template>
            <span class="wc-history-stake">{{ item.stake }} × {{ item.odds_snapshot.toFixed(2) }}</span>
          </div>
        </div>

        <div class="wc-history-result">
          <template v-if="item.cancelled_at">
            <div class="wc-result-badge wc-result--cancelled">{{ t('wc.betCancelledBadge') }}</div>
            <div v-if="item.cancel_penalty && item.cancel_penalty > 0" class="wc-history-penalty">
              -{{ item.cancel_penalty }} {{ t('wc.cancelPenaltyLabel') }}
            </div>
          </template>
          <template v-else-if="!item.result">
            <span class="wc-result-badge wc-result--pending">{{ t('wc.resultPending') }}</span>
          </template>
          <template v-else-if="item.result === 'win' || item.result === 'correct'">
            <span class="wc-result-badge wc-result--win">+{{ ((item.payout ?? 0) - item.stake).toFixed(2) }} {{ t('wc.resultWin') }}</span>
          </template>
          <template v-else-if="item.result === 'win_half'">
            <span class="wc-result-badge wc-result--win-half">+{{ ((item.payout ?? 0) - item.stake).toFixed(2) }} {{ t('wc.resultWinHalf') }}</span>
          </template>
          <template v-else-if="item.result === 'lose_half'">
            <span class="wc-result-badge wc-result--lose-half">-{{ (item.stake - (item.payout ?? 0)).toFixed(2) }} {{ t('wc.resultLoseHalf') }}</span>
          </template>
          <template v-else-if="item.result === 'lose' || item.result === 'incorrect' || item.result === 'lost'">
            <span class="wc-result-badge wc-result--lose">-{{ item.stake }} {{ t('wc.resultLose') }}</span>
          </template>
          <template v-else>
            <span class="wc-result-badge wc-result--push">±0 {{ t('wc.resultPush') }}</span>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useWcBetTypeLabel } from '@/utils/wcBetType'
import type { BetHistoryItem } from '@/types/wc'

const { t } = useI18n()
const betTypeLabel = useWcBetTypeLabel()

defineProps<{ items: BetHistoryItem[] }>()

function formatDate(s: string) {
  return new Date(s).toLocaleDateString('vi-VN', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' })
}
</script>

<style scoped>
.wc-history-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.wc-history-row {
  background: var(--surface-card);
  border: 1px solid var(--border-default);
  border-radius: 10px;
  padding: 12px 14px;
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.wc-history-row--cancelled {
  opacity: 0.65;
}

.wc-history-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.wc-history-match-info {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.wc-history-teams {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.wc-history-date {
  font-size: 11px;
  color: var(--text-muted);
}

.wc-history-details {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.wc-history-type {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  background: var(--surface-page);
  padding: 2px 7px;
  border-radius: 4px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.wc-history-choice,
.wc-history-option {
  font-size: 13px;
  font-weight: 700;
  color: var(--text-primary);
}

.wc-history-stake {
  font-size: 12px;
  color: var(--text-secondary);
}

.wc-history-result {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
}

.wc-history-penalty {
  font-size: 11px;
  color: #ef4444;
  font-weight: 600;
}

.wc-result-badge {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 700;
}

.wc-result--pending {
  background: rgba(100, 116, 139, 0.1);
  color: #64748b;
}

.wc-result--win {
  background: rgba(22, 163, 74, 0.12);
  color: #16a34a;
}

.wc-result--lose {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.wc-result--win-half {
  background: rgba(22, 163, 74, 0.08);
  color: #16a34a;
}

.wc-result--lose-half {
  background: rgba(239, 68, 68, 0.07);
  color: #ef4444;
}

.wc-result--push {
  background: rgba(217, 119, 6, 0.1);
  color: #d97706;
}

.wc-result--cancelled {
  background: rgba(100, 116, 139, 0.15);
  color: #64748b;
  text-decoration: line-through;
}
</style>
