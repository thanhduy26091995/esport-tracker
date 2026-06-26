<template>
  <div class="analytics-my">
    <!-- Period filter – always visible -->
    <div class="analytics-period-bar">
      <div class="analytics-period-pills">
        <button
          v-for="p in periodOptions"
          :key="p.value"
          class="analytics-period-pill"
          :class="{ 'analytics-period-pill--active': currentPeriod === p.value }"
          @click="selectPeriod(p.value)"
        >
          {{ p.label }}
        </button>
      </div>
      <div v-if="currentPeriod === 'custom'" class="analytics-picker-row">
        <el-date-picker
          v-model="customRange"
          type="daterange"
          size="small"
          format="DD/MM/YYYY"
          value-format="YYYY-MM-DD"
          start-placeholder="Từ ngày"
          end-placeholder="Đến ngày"
          @change="onCustomRange"
        />
      </div>
    </div>

    <div v-if="loading" class="analytics-loading">
      <div class="analytics-loading-inner">{{ t('common.loading') }}</div>
    </div>

    <div v-else-if="!data || data.settled_matches === 0" class="analytics-empty">
      <div class="analytics-empty-icon">📊</div>
      <div class="analytics-empty-title">{{ t('wc.analytics.noData') }}</div>
      <div class="analytics-empty-desc">{{ t('wc.analytics.noDataDesc') }}</div>
    </div>

    <template v-else>
      <!-- Profile badge -->
      <div v-if="data.profile_label" class="analytics-profile-card">
        <div class="analytics-profile-icon">{{ profileIcon(data.profile_label) }}</div>
        <div class="analytics-profile-info">
          <div class="analytics-profile-name">{{ profileName(data.profile_label) }}</div>
          <div class="analytics-profile-sub">{{ t('wc.analytics.profileLabel') }}</div>
        </div>
      </div>

      <!-- Stat grid -->
      <div class="analytics-stat-grid">
        <div class="analytics-stat-card analytics-stat-card--accent">
          <div class="analytics-stat-value">{{ pct(data.accuracy) }}%</div>
          <div class="analytics-stat-label">{{ t('wc.analytics.accuracy') }}</div>
        </div>
        <div class="analytics-stat-card">
          <div class="analytics-stat-value">{{ data.settled_matches }}</div>
          <div class="analytics-stat-label">{{ t('wc.analytics.settledMatches') }}</div>
        </div>
        <div class="analytics-stat-card analytics-stat-card--win">
          <div class="analytics-stat-value">{{ data.wins }}</div>
          <div class="analytics-stat-label">{{ t('wc.analytics.wins') }}</div>
        </div>
        <div class="analytics-stat-card analytics-stat-card--loss">
          <div class="analytics-stat-value">{{ data.losses }}</div>
          <div class="analytics-stat-label">{{ t('wc.analytics.losses') }}</div>
        </div>
      </div>

      <!-- Streak row -->
      <div class="analytics-streak-row">
        <div class="analytics-streak-badge analytics-streak-badge--win">
          <span class="analytics-streak-num">{{ data.current_win_streak }}</span>
          <span class="analytics-streak-text">{{ t('wc.analytics.currentWinStreak') }}</span>
        </div>
        <div class="analytics-streak-badge analytics-streak-badge--lose">
          <span class="analytics-streak-num">{{ data.current_lose_streak }}</span>
          <span class="analytics-streak-text">{{ t('wc.analytics.currentLoseStreak') }}</span>
        </div>
        <div class="analytics-streak-badge analytics-streak-badge--best">
          <span class="analytics-streak-num">{{ data.longest_win_streak }}</span>
          <span class="analytics-streak-text">{{ t('wc.analytics.longestWinStreak') }}</span>
        </div>
      </div>

      <!-- Accuracy timeline chart -->
      <div v-if="data.accuracy_timeline?.length > 0" class="analytics-section">
        <div class="analytics-section-title">{{ t('wc.analytics.accuracyTimeline') }}</div>
        <AccuracyTimelineChart :points="data.accuracy_timeline" />
      </div>

      <!-- Charts row: bet type + home bias -->
      <div class="analytics-charts-row">
        <div class="analytics-section analytics-section--half">
          <div class="analytics-section-title">{{ t('wc.analytics.betDistribution') }}</div>
          <BetTypeChart
            :handicap="data.bet_type_distribution.handicap"
            :exact-score="data.bet_type_distribution.exact_score"
            :over-under="data.bet_type_distribution.over_under"
          />
        </div>
        <div v-if="data.compare_metrics.home_bias !== null" class="analytics-section analytics-section--half">
          <div class="analytics-section-title">{{ t('wc.analytics.homeBias') }}</div>
          <div class="analytics-bias">
            <div class="analytics-bias-row">
              <span class="analytics-bias-label">{{ t('wc.analytics.homeTeam') }}</span>
              <span class="analytics-bias-pct analytics-bias-pct--home">{{ pct(data.compare_metrics.home_bias!) }}%</span>
            </div>
            <div class="analytics-bias-track">
              <div
                class="analytics-bias-fill analytics-bias-fill--home"
                :style="{ width: pct(data.compare_metrics.home_bias!) + '%' }"
              />
              <div
                class="analytics-bias-fill analytics-bias-fill--away"
                :style="{ width: pct(1 - data.compare_metrics.home_bias!) + '%' }"
              />
            </div>
            <div class="analytics-bias-row">
              <span class="analytics-bias-label">{{ t('wc.analytics.awayTeam') }}</span>
              <span class="analytics-bias-pct analytics-bias-pct--away">{{ pct(1 - data.compare_metrics.home_bias!) }}%</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Favorite teams -->
      <div v-if="data.favorite_teams?.length > 0" class="analytics-section">
        <div class="analytics-section-title">{{ t('wc.analytics.favoriteTeams') }}</div>
        <div class="analytics-list">
          <div v-for="(entry, i) in data.favorite_teams.slice(0, 5)" :key="entry.team" class="analytics-list-row">
            <span class="analytics-list-rank">{{ i + 1 }}</span>
            <span class="analytics-list-name">{{ entry.team }}</span>
            <span class="analytics-list-count">{{ entry.bet_count }} {{ t('wc.analytics.bets') }}</span>
          </div>
        </div>
      </div>

      <!-- Favorite scorelines -->
      <div v-if="data.favorite_scorelines?.length > 0" class="analytics-section">
        <div class="analytics-section-title">{{ t('wc.analytics.favoriteScorelines') }}</div>
        <div class="analytics-list">
          <div v-for="(entry, i) in data.favorite_scorelines.slice(0, 5)" :key="entry.scoreline" class="analytics-list-row">
            <span class="analytics-list-rank">{{ i + 1 }}</span>
            <span class="analytics-list-name analytics-list-name--mono">{{ entry.scoreline }}</span>
            <span class="analytics-list-count">{{ entry.count }} {{ t('wc.analytics.bets') }}</span>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { MyAnalyticsResponse } from '@/types/wc'
import AccuracyTimelineChart from './AccuracyTimelineChart.vue'
import BetTypeChart from './BetTypeChart.vue'

defineProps<{
  data: MyAnalyticsResponse | null
  loading: boolean
}>()

const emit = defineEmits<{
  'period-change': [payload: { period: string; dateFrom?: string; dateTo?: string }]
}>()

const { t } = useI18n()

const currentPeriod = ref('30d')
const customRange = ref<[string, string] | null>(null)

const periodOptions = computed(() => [
  { value: 'today', label: t('wc.analytics.period.today') },
  { value: '7d', label: t('wc.analytics.period.7d') },
  { value: '14d', label: t('wc.analytics.period.14d') },
  { value: '30d', label: t('wc.analytics.period.30d') },
  { value: 'custom', label: t('wc.analytics.period.custom') },
])

function selectPeriod(value: string) {
  currentPeriod.value = value
  if (value !== 'custom') {
    customRange.value = null
    emit('period-change', { period: value, dateFrom: undefined, dateTo: undefined })
  }
}

function onCustomRange(range: [string, string] | null) {
  if (range?.[0] && range?.[1]) {
    emit('period-change', { period: 'custom', dateFrom: range[0], dateTo: range[1] })
  }
}

function pct(v: number): number {
  return Math.round(v * 100)
}

const PROFILE_ICONS: Record<string, string> = {
  balanced_predictor: '⚖️',
  aggressive_predictor: '🔥',
  conservative_predictor: '🛡️',
  underdog_lover: '🐉',
  goal_hunter: '🎯',
}

function profileIcon(label: string): string {
  return PROFILE_ICONS[label] ?? '🎲'
}

function profileName(label: string): string {
  return t(`wc.analytics.profiles.${label}`)
}
</script>

<style scoped>
.analytics-my {
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

.analytics-empty-desc {
  font-size: 13px;
  color: var(--text-muted);
  margin-top: 6px;
}

.analytics-period-bar {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 16px;
}

.analytics-period-pills {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.analytics-period-pill {
  display: inline-flex;
  align-items: center;
  padding: 4px 12px;
  border-radius: 20px;
  border: 1px solid var(--border-default);
  background: var(--surface-card);
  color: var(--text-secondary);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s;
  white-space: nowrap;
}

.analytics-period-pill:hover {
  border-color: #16a34a60;
  color: var(--text-primary);
}

.analytics-period-pill--active {
  background: #16a34a;
  border-color: #16a34a;
  color: #fff;
}

.analytics-picker-row {
  margin-top: 4px;
}

.analytics-profile-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: var(--surface-card);
  border-radius: 10px;
  margin-bottom: 16px;
  border: 1px solid var(--border-default);
}

.analytics-profile-icon {
  font-size: 28px;
  line-height: 1;
}

.analytics-profile-name {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-primary);
}

.analytics-profile-sub {
  font-size: 11px;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-top: 2px;
}

.analytics-stat-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 10px;
  margin-bottom: 16px;
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
.analytics-stat-card--win .analytics-stat-value { color: #16a34a; }
.analytics-stat-card--loss .analytics-stat-value { color: #ef4444; }

.analytics-streak-row {
  display: flex;
  gap: 8px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.analytics-streak-badge {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 10px 16px;
  border-radius: 10px;
  flex: 1;
  min-width: 90px;
}

.analytics-streak-badge--win {
  background: rgba(22, 163, 74, 0.08);
  border: 1px solid rgba(22, 163, 74, 0.25);
}

.analytics-streak-badge--lose {
  background: rgba(239, 68, 68, 0.08);
  border: 1px solid rgba(239, 68, 68, 0.25);
}

.analytics-streak-badge--best {
  background: rgba(245, 158, 11, 0.08);
  border: 1px solid rgba(245, 158, 11, 0.25);
}

.analytics-streak-num {
  font-size: 24px;
  font-weight: 800;
  font-variant-numeric: tabular-nums;
}

.analytics-streak-badge--win .analytics-streak-num { color: #16a34a; }
.analytics-streak-badge--lose .analytics-streak-num { color: #ef4444; }
.analytics-streak-badge--best .analytics-streak-num { color: #d97706; }

.analytics-streak-text {
  font-size: 9px;
  font-weight: 600;
  color: var(--text-muted);
  text-align: center;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin-top: 2px;
}

.analytics-section {
  margin-bottom: 20px;
}

.analytics-section-title {
  font-size: 11px;
  font-weight: 700;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  margin-bottom: 10px;
}

.analytics-charts-row {
  display: flex;
  gap: 20px;
  margin-bottom: 4px;
}

.analytics-section--half {
  flex: 1;
  min-width: 0;
}

.analytics-bias {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 8px 0;
}

.analytics-bias-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.analytics-bias-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
}

.analytics-bias-pct {
  font-size: 14px;
  font-weight: 800;
  font-variant-numeric: tabular-nums;
}

.analytics-bias-pct--home { color: #3b82f6; }
.analytics-bias-pct--away { color: #f59e0b; }

.analytics-bias-track {
  display: flex;
  height: 10px;
  border-radius: 5px;
  overflow: hidden;
  background: var(--surface-page);
}

.analytics-bias-fill {
  height: 100%;
  transition: width 0.4s ease;
}

.analytics-bias-fill--home { background: #3b82f6; }
.analytics-bias-fill--away { background: #f59e0b; }

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

@media (max-width: 600px) {
  .analytics-stat-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .analytics-charts-row {
    flex-direction: column;
  }
}
</style>
