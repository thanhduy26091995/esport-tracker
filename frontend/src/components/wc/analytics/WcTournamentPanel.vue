<template>
  <div class="wc-tournament-panel">
    <div v-if="loading" class="analytics-loading">
      <div class="analytics-loading-inner">{{ t('common.loading') }}</div>
    </div>

    <div v-else-if="!data" class="analytics-empty">
      <div class="analytics-empty-icon">⚽</div>
      <div class="analytics-empty-title">{{ t('wc.analytics.tournament.noData') }}</div>
    </div>

    <template v-else>
      <!-- Section 1: Stat cards -->
      <div class="analytics-stat-grid">
        <div class="analytics-stat-card analytics-stat-card--accent">
          <div class="analytics-stat-value">{{ data.match_stats.total_goals }}</div>
          <div class="analytics-stat-label">{{ t('wc.analytics.tournament.totalGoals') }}</div>
        </div>
        <div class="analytics-stat-card">
          <div class="analytics-stat-value">{{ data.match_stats.avg_goals_per_match.toFixed(2) }}</div>
          <div class="analytics-stat-label">{{ t('wc.analytics.tournament.avgGoals') }}</div>
        </div>
        <div class="analytics-stat-card">
          <div class="analytics-stat-value">{{ data.match_stats.total_matches }}</div>
          <div class="analytics-stat-label">{{ t('wc.analytics.tournament.totalMatches') }}</div>
        </div>
        <div class="analytics-stat-card">
          <div class="analytics-stat-value">{{ data.match_stats.clean_sheets }}</div>
          <div class="analytics-stat-label">{{ t('wc.analytics.tournament.cleanSheets') }}</div>
        </div>
      </div>

      <!-- Section 2: Top Scorers -->
      <div class="analytics-section">
        <div class="analytics-section-title">
          {{ t('wc.analytics.tournament.topScorers') }}
          <span v-if="data.scorers_updated_at" class="section-title-meta">
            · {{ t('wc.analytics.tournament.updatedAt') }} {{ formatTime(data.scorers_updated_at) }}
          </span>
        </div>
        <div v-if="!data.top_scorers?.length" class="analytics-section-empty">
          {{ t('wc.analytics.tournament.scorersUnavailable') }}
        </div>
        <div v-else class="predictors-table">
          <div class="predictors-table-header">
            <span class="predictors-col predictors-col--rank">#</span>
            <span class="predictors-col predictors-col--name">{{ t('wc.analytics.tournament.player') }}</span>
            <span class="predictors-col predictors-col--stat">⚽</span>
            <span class="predictors-col predictors-col--stat">🎯</span>
            <span class="predictors-col predictors-col--stat">ST</span>
          </div>
          <div
            v-for="(scorer, i) in data.top_scorers"
            :key="scorer.player_name"
            class="predictors-table-row"
            :class="{ 'predictors-table-row--top3': i < 3 }"
          >
            <span class="predictors-col predictors-col--rank">
              <span v-if="i === 0">🥇</span>
              <span v-else-if="i === 1">🥈</span>
              <span v-else-if="i === 2">🥉</span>
              <span v-else>{{ scorer.rank }}</span>
            </span>
            <span class="predictors-col predictors-col--name">
              <img
                v-if="scorer.team_crest"
                :src="scorer.team_crest"
                class="predictors-avatar"
                :alt="scorer.team_code"
                @error="(e: Event) => ((e.target as HTMLImageElement).style.display = 'none')"
              />
              <span class="predictors-name-stack">
                <span class="predictors-name">{{ scorer.player_name }}</span>
                <span class="predictors-sub">{{ scorer.team_name }}</span>
              </span>
            </span>
            <span class="predictors-col predictors-col--stat predictors-accuracy">{{ scorer.goals }}</span>
            <span class="predictors-col predictors-col--stat predictors-matches">{{ scorer.assists ?? '-' }}</span>
            <span class="predictors-col predictors-col--stat predictors-matches">{{ scorer.played_matches }}</span>
          </div>
        </div>
      </div>

      <!-- Section 3: Result Breakdown -->
      <div class="analytics-section">
        <div class="analytics-section-title">{{ t('wc.analytics.tournament.resultBreakdown') }}</div>
        <div class="analytics-streak-row">
          <div class="analytics-streak-badge analytics-streak-badge--win">
            <span class="analytics-streak-num">{{ data.match_stats.home_wins }}</span>
            <span class="analytics-streak-text">{{ t('wc.analytics.tournament.homeWins') }}</span>
          </div>
          <div class="analytics-streak-badge analytics-streak-badge--best">
            <span class="analytics-streak-num">{{ data.match_stats.draws }}</span>
            <span class="analytics-streak-text">{{ t('wc.analytics.tournament.draws') }}</span>
          </div>
          <div class="analytics-streak-badge analytics-streak-badge--lose">
            <span class="analytics-streak-num">{{ data.match_stats.away_wins }}</span>
            <span class="analytics-streak-text">{{ t('wc.analytics.tournament.awayWins') }}</span>
          </div>
        </div>
        <div class="analytics-bias-track result-bar">
          <div
            class="analytics-bias-fill analytics-bias-fill--home"
            :style="{ width: pct(data.match_stats.home_wins, data.match_stats.total_matches) + '%' }"
          />
          <div
            class="analytics-bias-fill result-fill--draw"
            :style="{ width: pct(data.match_stats.draws, data.match_stats.total_matches) + '%' }"
          />
          <div
            class="analytics-bias-fill analytics-bias-fill--away"
            :style="{ width: pct(data.match_stats.away_wins, data.match_stats.total_matches) + '%' }"
          />
        </div>
      </div>

      <!-- Section 4: Half-Time Analysis -->
      <div v-if="data.half_time_stats" class="analytics-section">
        <div class="analytics-section-title">{{ t('wc.analytics.tournament.halfTimeAnalysis') }}</div>
        <div class="analytics-streak-row">
          <div class="analytics-streak-badge analytics-streak-badge--best">
            <span class="analytics-streak-num">{{ data.half_time_stats.first_half_goals }}</span>
            <span class="analytics-streak-text">{{ t('wc.analytics.tournament.firstHalf') }}</span>
          </div>
          <div class="analytics-streak-badge analytics-streak-badge--best">
            <span class="analytics-streak-num">{{ data.half_time_stats.second_half_goals }}</span>
            <span class="analytics-streak-text">{{ t('wc.analytics.tournament.secondHalf') }}</span>
          </div>
          <div class="analytics-streak-badge analytics-streak-badge--win">
            <span class="analytics-streak-num">{{ data.half_time_stats.comebacks }}</span>
            <span class="analytics-streak-text">{{ t('wc.analytics.tournament.comebacks') }}</span>
          </div>
          <div class="analytics-streak-badge analytics-streak-badge--lose">
            <span class="analytics-streak-num">{{ data.half_time_stats.held_lead }}</span>
            <span class="analytics-streak-text">{{ t('wc.analytics.tournament.heldLead') }}</span>
          </div>
          <div class="analytics-streak-badge analytics-streak-badge--purple">
            <span class="analytics-streak-num">{{ data.half_time_stats.own_goals }}</span>
            <span class="analytics-streak-text">{{ t('wc.analytics.tournament.ownGoals') }}</span>
          </div>
          <div class="analytics-streak-badge analytics-streak-badge--pink">
            <span class="analytics-streak-num">{{ data.half_time_stats.penalty_goals }}</span>
            <span class="analytics-streak-text">{{ t('wc.analytics.tournament.penaltyGoals') }}</span>
          </div>
        </div>
      </div>

      <!-- Section 5: Goal Timing -->
      <div v-if="data.goal_timing?.length" class="analytics-section">
        <div class="analytics-section-title">{{ t('wc.analytics.tournament.goalTiming') }}</div>
        <div class="bar-list">
          <div v-for="b in data.goal_timing" :key="b.label" class="bar-row">
            <span class="bar-row-label">{{ b.label }}</span>
            <div class="bar-row-track">
              <div
                class="bar-row-fill"
                :style="{ width: barPct(b.goals, maxTimingGoals) + '%' }"
              />
            </div>
            <span class="bar-row-count">{{ b.goals }}</span>
          </div>
        </div>
      </div>

      <!-- Section 6: Goals by Stage -->
      <div v-if="data.match_stats.goals_by_stage?.length" class="analytics-section">
        <div class="analytics-section-title">{{ t('wc.analytics.tournament.goalsByStage') }}</div>
        <div class="bar-list">
          <div v-for="s in data.match_stats.goals_by_stage" :key="s.stage" class="bar-row">
            <span class="bar-row-label bar-row-label--wide">{{ s.stage }}</span>
            <div class="bar-row-track">
              <div
                class="bar-row-fill bar-row-fill--blue"
                :style="{ width: barPct(s.goals, maxStageGoals) + '%' }"
              />
            </div>
            <span class="bar-row-count">{{ s.goals }} / {{ s.matches }}m</span>
          </div>
        </div>
      </div>

      <!-- Section 7: Goals by Group -->
      <div v-if="data.goals_by_group?.length" class="analytics-section">
        <div class="analytics-section-title">{{ t('wc.analytics.tournament.goalsByGroup') }}</div>
        <div class="bar-list">
          <div v-for="g in data.goals_by_group" :key="g.group" class="bar-row">
            <span class="bar-row-label">{{ g.group }}</span>
            <div class="bar-row-track">
              <div
                class="bar-row-fill bar-row-fill--amber"
                :style="{ width: barPct(g.goals, maxGroupGoals) + '%' }"
              />
            </div>
            <span class="bar-row-count">{{ g.goals }} / {{ g.matches }}m</span>
          </div>
        </div>
      </div>

      <!-- Section 8: Top Scoring Matches -->
      <div v-if="data.top_scoring_matches?.length" class="analytics-section">
        <div class="analytics-section-title">{{ t('wc.analytics.tournament.topScoringMatches') }}</div>
        <div class="analytics-list">
          <div
            v-for="m in data.top_scoring_matches.slice(0, 5)"
            :key="`${m.home_team}-${m.away_team}-${m.date}`"
            class="analytics-list-row match-row"
          >
            <span class="match-score">{{ m.home_score }}:{{ m.away_score }}</span>
            <span class="analytics-list-name match-teams">
              {{ m.home_team }}
              <span class="match-vs">vs</span>
              {{ m.away_team }}
              <span v-if="m.round" class="match-round">· {{ m.round }}</span>
            </span>
            <span class="analytics-list-count">{{ m.total_goals }} {{ t('wc.analytics.tournament.goals') }}</span>
          </div>
        </div>
      </div>

      <!-- Section 9: Team Statistics -->
      <div v-if="data.team_stats?.length" class="analytics-section">
        <div class="analytics-section-title">{{ t('wc.analytics.tournament.teamStats') }}</div>
        <div class="predictors-table">
          <div class="predictors-table-header">
            <span class="predictors-col predictors-col--rank">#</span>
            <span class="predictors-col predictors-col--name">{{ t('wc.analytics.tournament.team') }}</span>
            <span class="predictors-col predictors-col--stat">{{ t('wc.analytics.tournament.gf') }}</span>
            <span class="predictors-col predictors-col--stat">{{ t('wc.analytics.tournament.ga') }}</span>
            <span class="predictors-col predictors-col--stat">{{ t('wc.analytics.tournament.gd') }}</span>
            <span class="predictors-col predictors-col--stat">{{ t('wc.analytics.tournament.mp') }}</span>
          </div>
          <div
            v-for="(ts, i) in data.team_stats.slice(0, 20)"
            :key="ts.team_name"
            class="predictors-table-row"
          >
            <span class="predictors-col predictors-col--rank">{{ i + 1 }}</span>
            <span class="predictors-col predictors-col--name">
              <span class="predictors-name">{{ ts.team_name }}</span>
            </span>
            <span class="predictors-col predictors-col--stat predictors-accuracy">{{ ts.goals_for }}</span>
            <span class="predictors-col predictors-col--stat predictors-matches">{{ ts.goals_against }}</span>
            <span
              class="predictors-col predictors-col--stat"
              :class="gd(ts) > 0 ? 'stat-positive' : gd(ts) < 0 ? 'stat-negative' : 'predictors-matches'"
            >{{ gd(ts) > 0 ? '+' : '' }}{{ gd(ts) }}</span>
            <span class="predictors-col predictors-col--stat predictors-matches">{{ ts.matches }}</span>
          </div>
        </div>
      </div>

      <!-- Section 10: Venue Statistics -->
      <div v-if="data.venue_stats?.length" class="analytics-section">
        <div class="analytics-section-title">{{ t('wc.analytics.tournament.venues') }}</div>
        <div class="predictors-table">
          <div class="predictors-table-header">
            <span class="predictors-col predictors-col--name">{{ t('wc.analytics.tournament.venue') }}</span>
            <span class="predictors-col predictors-col--stat">{{ t('wc.analytics.tournament.matches') }}</span>
            <span class="predictors-col predictors-col--stat">{{ t('wc.analytics.tournament.goals') }}</span>
            <span class="predictors-col predictors-col--stat">{{ t('wc.analytics.tournament.avgGoals') }}</span>
          </div>
          <div
            v-for="v in data.venue_stats"
            :key="v.venue"
            class="predictors-table-row"
          >
            <span class="predictors-col predictors-col--name">
              <span class="predictors-name">{{ v.venue }}</span>
            </span>
            <span class="predictors-col predictors-col--stat predictors-matches">{{ v.matches }}</span>
            <span class="predictors-col predictors-col--stat predictors-accuracy">{{ v.goals }}</span>
            <span class="predictors-col predictors-col--stat predictors-matches">
              {{ v.matches > 0 ? (v.goals / v.matches).toFixed(2) : '–' }}
            </span>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { WcAnalyticsResponse, WcTeamStat } from '@/types/wc'

const props = defineProps<{
  data: WcAnalyticsResponse | null
  loading: boolean
}>()

const { t } = useI18n()

const maxTimingGoals = computed(() =>
  Math.max(...(props.data?.goal_timing?.map(b => b.goals) ?? []), 1)
)
const maxStageGoals = computed(() =>
  Math.max(...(props.data?.match_stats.goals_by_stage?.map(s => s.goals) ?? []), 1)
)
const maxGroupGoals = computed(() =>
  Math.max(...(props.data?.goals_by_group?.map(g => g.goals) ?? []), 1)
)

function barPct(val: number, max: number): number {
  if (max === 0) return 0
  return Math.round((val / max) * 100)
}

function pct(val: number, total: number): number {
  if (total === 0) return 0
  return Math.round((val / total) * 100)
}

function gd(ts: WcTeamStat): number {
  return ts.goals_for - ts.goals_against
}

function formatTime(iso?: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  return d.toLocaleTimeString('vi-VN', { hour: '2-digit', minute: '2-digit' })
}
</script>

<style scoped>
.wc-tournament-panel {
  padding: 4px 0;
}

/* Loading / Empty */
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

/* Stat grid */
.analytics-stat-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
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

/* Section */
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

.section-title-meta {
  font-weight: 400;
  font-size: 10px;
  text-transform: none;
  letter-spacing: 0;
  color: var(--text-muted);
}

.analytics-section-empty {
  font-size: 13px;
  color: var(--text-muted);
  padding: 12px 0;
}

/* Predictors table */
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
  gap: 8px;
}

.predictors-table-row {
  display: flex;
  align-items: center;
  padding: 10px 12px;
  gap: 8px;
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
  width: 44px;
  text-align: right;
  font-size: 13px;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

.predictors-avatar {
  width: 22px;
  height: 22px;
  border-radius: 3px;
  object-fit: contain;
  flex-shrink: 0;
}

.predictors-name-stack {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.predictors-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.predictors-sub {
  font-size: 10px;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.predictors-accuracy { color: #16a34a; font-weight: 800; }
.predictors-matches { color: var(--text-muted); }
.stat-positive { color: #16a34a; font-weight: 700; }
.stat-negative { color: #ef4444; font-weight: 700; }

/* Streak badges (result / half-time) */
.analytics-streak-row {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.analytics-streak-badge {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 10px 16px;
  border-radius: 10px;
  flex: 1;
  min-width: 80px;
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
.analytics-streak-badge--purple {
  background: rgba(99, 102, 241, 0.08);
  border: 1px solid rgba(99, 102, 241, 0.25);
}
.analytics-streak-badge--pink {
  background: rgba(236, 72, 153, 0.08);
  border: 1px solid rgba(236, 72, 153, 0.25);
}

.analytics-streak-num {
  font-size: 24px;
  font-weight: 800;
  font-variant-numeric: tabular-nums;
}
.analytics-streak-badge--win .analytics-streak-num { color: #16a34a; }
.analytics-streak-badge--lose .analytics-streak-num { color: #ef4444; }
.analytics-streak-badge--best .analytics-streak-num { color: #d97706; }
.analytics-streak-badge--purple .analytics-streak-num { color: #6366f1; }
.analytics-streak-badge--pink .analytics-streak-num { color: #ec4899; }

.analytics-streak-text {
  font-size: 9px;
  font-weight: 600;
  color: var(--text-muted);
  text-align: center;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin-top: 2px;
}

/* Bias bar (result breakdown) */
.analytics-bias-track {
  display: flex;
  height: 10px;
  border-radius: 5px;
  overflow: hidden;
  background: var(--surface-page);
}

.result-bar {
  height: 12px;
  border-radius: 6px;
  margin-top: 12px;
}

.analytics-bias-fill {
  height: 100%;
  transition: width 0.4s ease;
}

.analytics-bias-fill--home { background: #3b82f6; }
.analytics-bias-fill--away { background: #ef4444; }
.result-fill--draw { background: #f59e0b; }

/* Bar list (timing / stage / group) */
.bar-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.bar-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.bar-row-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-secondary);
  min-width: 44px;
  white-space: nowrap;
  flex-shrink: 0;
}

.bar-row-label--wide {
  min-width: 80px;
}

.bar-row-track {
  flex: 1;
  height: 8px;
  border-radius: 4px;
  background: var(--surface-page);
  overflow: hidden;
}

.bar-row-fill {
  height: 100%;
  border-radius: 4px;
  background: #16a34a;
  transition: width 0.4s ease;
}

.bar-row-fill--blue { background: #3b82f6; }
.bar-row-fill--amber { background: #f59e0b; }

.bar-row-count {
  font-size: 11px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  color: var(--text-secondary);
  min-width: 52px;
  text-align: right;
  flex-shrink: 0;
}

/* Top scoring matches */
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

.analytics-list-name {
  flex: 1;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.analytics-list-count {
  font-size: 12px;
  color: var(--text-muted);
  font-weight: 500;
  white-space: nowrap;
  flex-shrink: 0;
}

.match-score {
  font-size: 15px;
  font-weight: 800;
  font-variant-numeric: tabular-nums;
  font-family: 'Courier New', monospace;
  color: var(--text-primary);
  min-width: 36px;
  text-align: center;
  flex-shrink: 0;
}

.match-teams {
  display: flex;
  align-items: center;
  gap: 5px;
  flex-wrap: wrap;
}

.match-vs {
  font-size: 10px;
  color: var(--text-muted);
  font-weight: 500;
  text-transform: uppercase;
}

.match-round {
  font-size: 10px;
  color: var(--text-muted);
  font-weight: 400;
}

@media (max-width: 600px) {
  .analytics-stat-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .analytics-streak-row {
    flex-wrap: wrap;
  }

  .analytics-streak-badge {
    min-width: 70px;
    padding: 8px 10px;
  }

  .analytics-streak-num {
    font-size: 20px;
  }

  .predictors-col--stat {
    width: 36px;
  }
}
</style>
