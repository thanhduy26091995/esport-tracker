<template>
  <div class="community-panel">
    <div v-if="loading" class="analytics-loading">
      <div class="analytics-loading-inner">{{ t('common.loading') }}</div>
    </div>

    <div v-else-if="!data" class="analytics-empty">
      <div class="analytics-empty-icon">🌍</div>
      <div class="analytics-empty-title">{{ t('wc.analytics.community.noData') }}</div>
    </div>

    <template v-else>
      <!-- Stat cards -->
      <div class="community-stat-grid">
        <div class="analytics-stat-card">
          <div class="analytics-stat-value">{{ data.total_bets_placed }}</div>
          <div class="analytics-stat-label">{{ t('wc.analytics.community.totalBets') }}</div>
        </div>
        <div class="analytics-stat-card">
          <div class="analytics-stat-value">{{ data.active_users }}</div>
          <div class="analytics-stat-label">{{ t('wc.analytics.community.activeUsers') }}</div>
        </div>
        <div class="analytics-stat-card analytics-stat-card--accent">
          <div class="analytics-stat-value">{{ pct(data.avg_accuracy) }}%</div>
          <div class="analytics-stat-label">{{ t('wc.analytics.community.avgAccuracy') }}</div>
        </div>
      </div>

      <!-- Charts row: prediction distribution + trending teams -->
      <div class="analytics-charts-row">
        <div class="analytics-section analytics-section--half">
          <div class="analytics-section-title">{{ t('wc.analytics.community.predictionDist') }}</div>
          <PredictionDistributionChart
            :home="data.prediction_distribution.home"
            :away="data.prediction_distribution.away"
            :other="data.prediction_distribution.other"
          />
        </div>
        <div v-if="data.trending_teams?.length > 0" class="analytics-section analytics-section--half">
          <div class="analytics-section-title">{{ t('wc.analytics.community.trendingTeams') }}</div>
          <TrendingTeamsChart :teams="data.trending_teams" />
        </div>
      </div>

      <!-- Trending scorelines -->
      <div v-if="data.trending_scorelines?.length > 0" class="analytics-section">
        <div class="analytics-section-title">{{ t('wc.analytics.community.trendingScorelines') }}</div>
        <div class="analytics-list">
          <div
            v-for="(entry, i) in data.trending_scorelines.slice(0, 5)"
            :key="entry.scoreline"
            class="analytics-list-row"
          >
            <span class="analytics-list-rank">{{ i + 1 }}</span>
            <span class="analytics-list-name analytics-list-name--mono">{{ entry.scoreline }}</span>
            <span class="analytics-list-count">{{ entry.count }} {{ t('wc.analytics.bets') }}</span>
          </div>
        </div>
      </div>

      <!-- Top predictors -->
      <div v-if="data.top_predictors?.length > 0" class="analytics-section">
        <div class="analytics-section-title">{{ t('wc.analytics.community.topPredictors') }}</div>
        <div class="predictors-table">
          <div class="predictors-table-header">
            <span class="predictors-col predictors-col--rank">#</span>
            <span class="predictors-col predictors-col--name">{{ t('wc.analytics.compare.metric') }}</span>
            <span class="predictors-col predictors-col--stat">{{ t('wc.analytics.accuracy') }}</span>
            <span class="predictors-col predictors-col--stat">{{ t('wc.analytics.settledMatches') }}</span>
          </div>
          <div
            v-for="(predictor, i) in data.top_predictors"
            :key="predictor.user_id"
            class="predictors-table-row"
            :class="{ 'predictors-table-row--top3': i < 3 }"
          >
            <span class="predictors-col predictors-col--rank">
              <span v-if="i === 0">🥇</span>
              <span v-else-if="i === 1">🥈</span>
              <span v-else-if="i === 2">🥉</span>
              <span v-else>{{ i + 1 }}</span>
            </span>
            <span class="predictors-col predictors-col--name">
              <img
                v-if="predictor.avatar_url"
                :src="predictor.avatar_url"
                class="predictors-avatar"
                :alt="predictor.name"
                @error="(e: Event) => ((e.target as HTMLImageElement).style.display = 'none')"
              />
              <span class="predictors-name">{{ predictor.name }}</span>
            </span>
            <span class="predictors-col predictors-col--stat predictors-accuracy">
              {{ pct(predictor.accuracy) }}%
            </span>
            <span class="predictors-col predictors-col--stat predictors-matches">
              {{ predictor.settled_matches }}
            </span>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { CommunityAnalyticsResponse } from '@/types/wc'
import PredictionDistributionChart from './PredictionDistributionChart.vue'
import TrendingTeamsChart from './TrendingTeamsChart.vue'

defineProps<{
  data: CommunityAnalyticsResponse | null
  loading: boolean
}>()

const { t } = useI18n()

function pct(v: number): number {
  return Math.round(v * 100)
}
</script>

<style scoped>
.community-panel {
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

.community-stat-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  margin-bottom: 20px;
}

.analytics-stat-card {
  background: var(--surface-card);
  border-radius: 10px;
  padding: 12px 8px;
  text-align: center;
  border: 1px solid var(--border-default);
}

.analytics-stat-value {
  font-size: 22px;
  font-weight: 800;
  font-variant-numeric: tabular-nums;
  color: var(--text-primary);
}

.analytics-stat-label {
  font-size: 10px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin-top: 3px;
}

.analytics-stat-card--accent .analytics-stat-value { color: #16a34a; }

.analytics-charts-row {
  display: flex;
  gap: 20px;
  margin-bottom: 4px;
}

.analytics-section {
  margin-bottom: 20px;
}

.analytics-section--half {
  flex: 1;
  min-width: 0;
}

.analytics-section-title {
  font-size: 11px;
  font-weight: 700;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  margin-bottom: 10px;
}

.analytics-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.analytics-list-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  background: var(--surface-card);
  border-radius: 8px;
  border: 1px solid var(--border-default);
}

.analytics-list-rank {
  font-size: 12px;
  font-weight: 700;
  color: var(--text-muted);
  min-width: 18px;
}

.analytics-list-name {
  flex: 1;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.analytics-list-name--mono {
  font-family: 'Courier New', monospace;
  font-size: 15px;
  font-weight: 700;
  letter-spacing: 0.05em;
}

.analytics-list-count {
  font-size: 12px;
  color: var(--text-muted);
  font-weight: 500;
  white-space: nowrap;
}

.predictors-table {
  border-radius: 10px;
  overflow: hidden;
  border: 1px solid var(--border-default);
}

.predictors-table-header {
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
  gap: 10px;
}

.predictors-table-row {
  display: flex;
  align-items: center;
  padding: 10px 12px;
  gap: 10px;
  background: var(--surface-card);
  border-bottom: 1px solid var(--border-default);
  transition: background 0.12s;
}

.predictors-table-row:last-child {
  border-bottom: none;
}

.predictors-table-row:hover {
  background: var(--surface-hover);
}

.predictors-table-row--top3 {
  background: rgba(22, 163, 74, 0.04);
}

.predictors-col {
  flex-shrink: 0;
}

.predictors-col--rank {
  width: 28px;
  font-size: 13px;
  font-weight: 700;
  color: var(--text-muted);
}

.predictors-col--name {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.predictors-col--stat {
  width: 70px;
  text-align: right;
  font-size: 13px;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

.predictors-avatar {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  object-fit: cover;
  border: 1.5px solid var(--border-default);
  flex-shrink: 0;
}

.predictors-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.predictors-accuracy {
  color: #16a34a;
  font-weight: 800;
}

.predictors-matches {
  color: var(--text-muted);
}

@media (max-width: 600px) {
  .community-stat-grid {
    grid-template-columns: repeat(3, 1fr);
  }

  .analytics-charts-row {
    flex-direction: column;
  }
}
</style>
