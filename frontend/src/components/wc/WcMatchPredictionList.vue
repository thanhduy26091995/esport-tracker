<template>
  <div class="wc-match-bets">
    <div class="wc-mb-header">
      <span class="wc-mb-title">{{ t('wc.predictorsOnMatch') }}</span>
      <span class="wc-mb-count">{{ predictions.length }}</span>
    </div>
    <div v-if="predictions.length === 0" class="wc-mb-empty">{{ t('wc.noPredictions') }}</div>
    <div v-else class="wc-mb-list">
      <div v-for="pred in predictions" :key="pred.id" class="wc-mb-row">
        <span class="wc-mb-user">{{ pred.name }}</span>
        <span class="wc-mb-type">
          {{ pred.prediction_type === 'handicap' ? t('wc.predictionTypeHandicap') : t('wc.predictionTypeExactScore') }}
        </span>
        <span class="wc-mb-choice">
          <template v-if="pred.prediction_type === 'exact_score'">
            {{ pred.predicted_home_score }}–{{ pred.predicted_away_score }}
          </template>
          <template v-else>
            {{ pred.prediction_choice }}
          </template>
        </span>
        <span class="wc-mb-stake">{{ pred.points }}</span>
        <span v-if="pred.result" class="wc-mb-result" :class="`wc-mb-result--${pred.result}`">
          {{ pred.result === 'correct' ? '+' + pred.points_earned : pred.result === 'incorrect' ? '-' + pred.points : '±0' }}
        </span>
        <span v-else class="wc-mb-result wc-mb-result--pending">?</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { WcPredictionPublic } from '@/types/wc'

const { t } = useI18n()
defineProps<{ predictions: WcPredictionPublic[] }>()
</script>

<style scoped>
.wc-match-bets {
  background: var(--surface-card);
  border: 1px solid var(--border-default);
  border-radius: 12px;
  overflow: hidden;
}

.wc-mb-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  border-bottom: 1px solid var(--border-subtle);
  background: var(--surface-page);
}

.wc-mb-title {
  font-size: 12px;
  font-weight: 700;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.wc-mb-count {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-muted);
  background: var(--border-default);
  padding: 1px 8px;
  border-radius: 10px;
}

.wc-mb-empty {
  padding: 16px;
  text-align: center;
  color: var(--text-muted);
  font-size: 13px;
}

.wc-mb-list {
  display: flex;
  flex-direction: column;
}

.wc-mb-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 14px;
  border-bottom: 1px solid var(--border-subtle);
  font-size: 12px;
  flex-wrap: wrap;
}

.wc-mb-row:last-child {
  border-bottom: none;
}

.wc-mb-user {
  font-weight: 600;
  color: var(--text-primary);
  min-width: 80px;
}

.wc-mb-type {
  color: var(--text-muted);
  font-size: 11px;
  background: var(--surface-page);
  padding: 1px 6px;
  border-radius: 4px;
}

.wc-mb-choice {
  font-weight: 600;
  color: var(--text-secondary);
}

.wc-mb-stake {
  color: var(--text-muted);
  margin-left: auto;
}

.wc-mb-result {
  font-weight: 700;
  tabular-nums: true;
  padding: 1px 7px;
  border-radius: 6px;
  font-size: 12px;
}

.wc-mb-result--correct {
  background: rgba(22, 163, 74, 0.12);
  color: #16a34a;
}

.wc-mb-result--incorrect {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.wc-mb-result--void {
  background: rgba(217, 119, 6, 0.1);
  color: #d97706;
}

.wc-mb-result--pending {
  background: rgba(100, 116, 139, 0.1);
  color: #94a3b8;
}
</style>
