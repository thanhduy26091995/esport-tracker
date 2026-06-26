<template>
  <div class="compare-panel">
    <div v-if="loading" class="analytics-loading">
      <div class="analytics-loading-inner">{{ t('common.loading') }}</div>
    </div>

    <div v-else-if="!data" class="analytics-empty">
      <div class="analytics-empty-icon">⚖️</div>
      <div class="analytics-empty-title">{{ t('wc.analytics.compare.noData') }}</div>
    </div>

    <template v-else>
      <!-- Accuracy headline -->
      <div class="compare-accuracy-row">
        <div class="compare-accuracy-card compare-accuracy-card--me" :class="{ 'compare-accuracy-card--better': data.my_accuracy >= data.community_accuracy }">
          <div class="compare-accuracy-value">{{ pct(data.my_accuracy) }}%</div>
          <div class="compare-accuracy-label">{{ t('wc.analytics.compare.me') }}</div>
        </div>
        <div class="compare-vs">VS</div>
        <div class="compare-accuracy-card compare-accuracy-card--community" :class="{ 'compare-accuracy-card--better': data.community_accuracy > data.my_accuracy }">
          <div class="compare-accuracy-value">{{ pct(data.community_accuracy) }}%</div>
          <div class="compare-accuracy-label">{{ t('wc.analytics.compare.community') }}</div>
        </div>
      </div>

      <!-- Metric table -->
      <div class="compare-table">
        <div class="compare-table-header">
          <span class="compare-col compare-col--metric">{{ t('wc.analytics.compare.metric') }}</span>
          <span class="compare-col compare-col--me">{{ t('wc.analytics.compare.me') }}</span>
          <span class="compare-col compare-col--community">{{ t('wc.analytics.compare.community') }}</span>
        </div>
        <div v-for="row in metricRows" :key="row.key" class="compare-table-row">
          <span class="compare-col compare-col--metric">{{ row.label }}</span>
          <span class="compare-col compare-col--me" :class="cellClass(row, 'me')">
            {{ formatMetric(row, row.meValue) }}
          </span>
          <span class="compare-col compare-col--community" :class="cellClass(row, 'community')">
            {{ formatMetric(row, row.communityValue) }}
          </span>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { CompareAnalyticsResponse } from '@/types/wc'

const props = defineProps<{
  data: CompareAnalyticsResponse | null
  loading: boolean
}>()

const { t } = useI18n()

function pct(v: number): number {
  return Math.round(v * 100)
}

type MetricRow = {
  key: string
  label: string
  meValue: number | null
  communityValue: number | null
  format: 'pct' | 'decimal' | 'pts' | 'rate'
  higherIsBetter: boolean
}

const metricRows = computed<MetricRow[]>(() => {
  if (!props.data) return []
  const me = props.data.me
  const com = props.data.community
  return [
    {
      key: 'home_bias',
      label: t('wc.analytics.compare.metrics.homeBias'),
      meValue: me.home_bias,
      communityValue: com.home_bias,
      format: 'pct',
      higherIsBetter: false,
    },
    {
      key: 'avg_goals',
      label: t('wc.analytics.compare.metrics.avgGoals'),
      meValue: me.avg_goals_predicted,
      communityValue: com.avg_goals_predicted,
      format: 'decimal',
      higherIsBetter: false,
    },
    {
      key: 'exact_score_rate',
      label: t('wc.analytics.compare.metrics.exactScoreRate'),
      meValue: me.exact_score_rate,
      communityValue: com.exact_score_rate,
      format: 'pct',
      higherIsBetter: false,
    },
    {
      key: 'underdog_rate',
      label: t('wc.analytics.compare.metrics.underdogRate'),
      meValue: me.underdog_rate,
      communityValue: com.underdog_rate,
      format: 'pct',
      higherIsBetter: false,
    },
    {
      key: 'avg_stake',
      label: t('wc.analytics.compare.metrics.avgStake'),
      meValue: me.avg_stake,
      communityValue: com.avg_stake,
      format: 'pts',
      higherIsBetter: false,
    },
    {
      key: 'over_preference',
      label: t('wc.analytics.compare.metrics.overPreference'),
      meValue: me.over_preference_rate,
      communityValue: com.over_preference_rate,
      format: 'pct',
      higherIsBetter: false,
    },
    {
      key: 'exact_score_hit_rate',
      label: t('wc.analytics.compare.metrics.exactScoreHitRate'),
      meValue: me.exact_score_hit_rate,
      communityValue: com.exact_score_hit_rate,
      format: 'pct',
      higherIsBetter: true,
    },
    {
      key: 'bet_frequency',
      label: t('wc.analytics.compare.metrics.betFrequency'),
      meValue: me.bet_frequency,
      communityValue: com.bet_frequency,
      format: 'decimal',
      higherIsBetter: false,
    },
    {
      key: 'last_minute_rate',
      label: t('wc.analytics.compare.metrics.lastMinuteRate'),
      meValue: me.last_minute_rate,
      communityValue: com.last_minute_rate,
      format: 'pct',
      higherIsBetter: false,
    },
  ]
})

function formatMetric(row: MetricRow, value: number | null): string {
  if (value === null) return 'N/A'
  switch (row.format) {
    case 'pct': return `${pct(value)}%`
    case 'decimal': return value.toFixed(1)
    case 'pts': return `${Math.round(value)} đ`
    case 'rate': return value.toFixed(2)
    default: return String(value)
  }
}

function cellClass(row: MetricRow, side: 'me' | 'community'): string {
  const meVal = row.meValue
  const comVal = row.communityValue
  if (meVal === null || comVal === null) return 'compare-cell--na'
  if (side === 'me') {
    if (row.higherIsBetter) return meVal > comVal ? 'compare-cell--better' : meVal < comVal ? 'compare-cell--worse' : 'compare-cell--equal'
    return 'compare-cell--neutral'
  }
  if (row.higherIsBetter) return comVal > meVal ? 'compare-cell--better' : comVal < meVal ? 'compare-cell--worse' : 'compare-cell--equal'
  return 'compare-cell--neutral'
}
</script>

<style scoped>
.compare-panel {
  padding: 4px 0;
}

.analytics-loading {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 48px;
  color: var(--text-muted);
}

.analytics-empty {
  text-align: center;
  padding: 48px 24px;
}

.analytics-empty-icon {
  font-size: 40px;
  margin-bottom: 12px;
}

.analytics-empty-title {
  font-size: 16px;
  font-weight: 700;
  color: var(--text-secondary);
}

.compare-accuracy-row {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
}

.compare-accuracy-card {
  flex: 1;
  text-align: center;
  padding: 16px;
  border-radius: 12px;
  background: var(--surface-card);
  border: 1px solid var(--border-default);
}

.compare-accuracy-card--better {
  border-color: #16a34a60;
  background: rgba(22, 163, 74, 0.06);
}

.compare-accuracy-value {
  font-size: 28px;
  font-weight: 800;
  font-variant-numeric: tabular-nums;
  color: var(--text-primary);
}

.compare-accuracy-card--better .compare-accuracy-value {
  color: #16a34a;
}

.compare-accuracy-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-top: 4px;
}

.compare-vs {
  font-size: 13px;
  font-weight: 800;
  color: var(--text-muted);
  flex-shrink: 0;
}

.compare-table {
  border-radius: 10px;
  overflow: hidden;
  border: 1px solid var(--border-default);
}

.compare-table-header {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  background: var(--surface-page);
  border-bottom: 1px solid var(--border-default);
  font-size: 10px;
  font-weight: 700;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  gap: 8px;
}

.compare-table-row {
  display: flex;
  align-items: center;
  padding: 10px 12px;
  gap: 8px;
  background: var(--surface-card);
  border-bottom: 1px solid var(--border-default);
}

.compare-table-row:last-child {
  border-bottom: none;
}

.compare-table-row:hover {
  background: var(--surface-hover);
}

.compare-col {
  font-size: 13px;
  font-variant-numeric: tabular-nums;
}

.compare-col--metric {
  flex: 1;
  font-weight: 600;
  color: var(--text-secondary);
  font-size: 12px;
}

.compare-col--me,
.compare-col--community {
  width: 80px;
  text-align: right;
  font-weight: 700;
  color: var(--text-primary);
}

.compare-cell--better {
  color: #16a34a !important;
}

.compare-cell--worse {
  color: #ef4444 !important;
}

.compare-cell--equal {
  color: var(--text-muted) !important;
}

.compare-cell--na {
  color: var(--text-muted) !important;
  font-weight: 500 !important;
  font-style: italic;
}

.compare-cell--neutral {
  color: var(--text-primary);
}
</style>
