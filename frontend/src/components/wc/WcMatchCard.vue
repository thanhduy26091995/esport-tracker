<template>
  <div class="wc-match-card" :class="[`wc-match-card--${match.status}`]">
    <div class="wc-match-stage-row">
      <span class="wc-match-stage">{{ stageLabel }}</span>
      <span v-if="match.group_name" class="wc-match-group">{{ match.group_name }}</span>
      <div class="wc-match-badges">
        <span v-if="isLocked" class="wc-badge wc-badge--locked">🔒 {{ t('wc.locked') }}</span>
        <span v-if="match.settled_at" class="wc-badge wc-badge--settled">✓ {{ t('wc.settled') }}</span>
        <span class="wc-badge" :class="`wc-badge--${match.status}`">{{ statusLabel }}</span>
      </div>
    </div>

    <div class="wc-match-body">
      <div class="wc-team wc-team--home">
        <span class="wc-team-flag">{{ homeFlag }}</span>
        <span class="wc-team-name">{{ match.home_team }}</span>
        <span v-if="match.home_team_code" class="wc-team-code">{{ match.home_team_code }}</span>
      </div>

      <div class="wc-score-center">
        <div v-if="hasScore" class="wc-score-display">
          <span class="wc-score-num">{{ match.home_score }}</span>
          <span class="wc-score-sep">–</span>
          <span class="wc-score-num">{{ match.away_score }}</span>
        </div>
        <div v-else class="wc-match-time">
          <div class="wc-time-date">{{ matchDateStr }}</div>
          <div class="wc-time-clock">{{ matchTimeStr }}</div>
        </div>
        <div v-if="match.handicap_value" class="wc-handicap-info">
          <span class="wc-hc-label">
            {{ handicapTeamName }} {{ t('wc.gives') }} {{ match.handicap_value }}
          </span>
        </div>
      </div>

      <div class="wc-team wc-team--away">
        <span v-if="match.away_team_code" class="wc-team-code">{{ match.away_team_code }}</span>
        <span class="wc-team-name">{{ match.away_team }}</span>
        <span class="wc-team-flag">{{ awayFlag }}</span>
      </div>
    </div>

    <div v-if="showActions" class="wc-match-actions">
      <slot name="actions" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { WcMatch } from '@/types/wc'

const { t } = useI18n()

const props = defineProps<{
  match: WcMatch
  showActions?: boolean
}>()

const stageLabels: Record<string, string> = {
  group: 'wc.stageGroup',
  r32: 'wc.stageR32',
  r16: 'wc.stageR16',
  qf: 'wc.stageQF',
  sf: 'wc.stageSF',
  final: 'wc.stageFinal',
  third_place: 'wc.stageThirdPlace',
}

const stageLabel = computed(() => t(stageLabels[props.match.stage] ?? 'wc.stageGroup'))

const statusLabels: Record<string, string> = {
  scheduled: 'wc.statusScheduled',
  live: 'wc.statusLive',
  completed: 'wc.statusCompleted',
  cancelled: 'wc.statusCancelled',
}
const statusLabel = computed(() => t(statusLabels[props.match.status] ?? 'wc.statusScheduled'))

const hasScore = computed(() =>
  props.match.status === 'live' || props.match.status === 'completed',
)

const isLocked = computed(() => {
  if (!props.match.bets_locked_at) return false
  return new Date(props.match.bets_locked_at) <= new Date()
})

const matchDate = computed(() => new Date(props.match.match_date))
const matchDateStr = computed(() =>
  matchDate.value.toLocaleDateString('vi-VN', { day: '2-digit', month: '2-digit' }),
)
const matchTimeStr = computed(() =>
  matchDate.value.toLocaleTimeString('vi-VN', { hour: '2-digit', minute: '2-digit' }),
)

const handicapTeamName = computed(() =>
  props.match.handicap_team === 'home' ? props.match.home_team : props.match.away_team,
)

function teamFlag(code?: string): string {
  if (!code) return '🏳️'
  const flag: Record<string, string> = {
    USA: '🇺🇸', MEX: '🇲🇽', CAN: '🇨🇦', BRA: '🇧🇷', ARG: '🇦🇷', FRA: '🇫🇷',
    ENG: '🏴󠁧󠁢󠁥󠁮󠁧󠁿', GER: '🇩🇪', ESP: '🇪🇸', POR: '🇵🇹', ITA: '🇮🇹',
    NED: '🇳🇱', BEL: '🇧🇪', URU: '🇺🇾', COL: '🇨🇴', CHI: '🇨🇱',
    JPN: '🇯🇵', KOR: '🇰🇷', AUS: '🇦🇺', MAR: '🇲🇦', SEN: '🇸🇳',
    NGA: '🇳🇬', GHA: '🇬🇭', CMR: '🇨🇲', EGY: '🇪🇬', TUN: '🇹🇳',
    SAU: '🇸🇦', IRN: '🇮🇷', QAT: '🇶🇦', SUI: '🇨🇭', AUT: '🇦🇹',
    DEN: '🇩🇰', SWE: '🇸🇪', NOR: '🇳🇴', POL: '🇵🇱', CRO: '🇭🇷',
    SRB: '🇷🇸', SCO: '🏴󠁧󠁢󠁳󠁣󠁴󠁿', WAL: '🏴󠁧󠁢󠁷󠁬󠁳󠁿', SVK: '🇸🇰', CZE: '🇨🇿',
    HUN: '🇭🇺', ROU: '🇷🇴', SLO: '🇸🇮', ALB: '🇦🇱', GEO: '🇬🇪',
    TUR: '🇹🇷', UKR: '🇺🇦', ECU: '🇪🇨', PER: '🇵🇪', BOL: '🇧🇴',
    PAR: '🇵🇾', VEN: '🇻🇪', PAN: '🇵🇦', CRI: '🇨🇷', HON: '🇭🇳',
    JAM: '🇯🇲', HAI: '🇭🇹', VIE: '🇻🇳', THA: '🇹🇭', IDN: '🇮🇩',
  }
  return flag[code.toUpperCase()] ?? '🏳️'
}

const homeFlag = computed(() => teamFlag(props.match.home_team_code))
const awayFlag = computed(() => teamFlag(props.match.away_team_code))
</script>

<style scoped>
.wc-match-card {
  background: var(--surface-card);
  border: 1px solid var(--border-default);
  border-radius: 14px;
  padding: 14px 16px;
  transition: border-color 0.15s, box-shadow 0.15s;
}

.wc-match-card:hover {
  border-color: #16a34a40;
  box-shadow: 0 4px 20px rgba(22, 163, 74, 0.08);
}

.wc-match-card--live {
  border-color: #16a34a60;
  background: linear-gradient(135deg, rgba(22, 163, 74, 0.03), var(--surface-card));
}

.wc-match-stage-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.wc-match-stage {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.wc-match-group {
  font-size: 11px;
  font-weight: 700;
  color: var(--color-primary);
  background: rgba(59, 130, 246, 0.1);
  padding: 1px 7px;
  border-radius: 4px;
}

.wc-match-badges {
  margin-left: auto;
  display: flex;
  gap: 6px;
  align-items: center;
}

.wc-badge {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 6px;
}

.wc-badge--scheduled {
  background: rgba(100, 116, 139, 0.12);
  color: #64748b;
}

.wc-badge--live {
  background: rgba(22, 163, 74, 0.12);
  color: #16a34a;
  animation: pulse-live 1.5s infinite;
}

.wc-badge--completed {
  background: rgba(59, 130, 246, 0.12);
  color: #3b82f6;
}

.wc-badge--cancelled {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.wc-badge--locked {
  background: rgba(217, 119, 6, 0.1);
  color: #d97706;
}

.wc-badge--settled {
  background: rgba(16, 185, 129, 0.1);
  color: #10b981;
}

@keyframes pulse-live {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}

.wc-match-body {
  display: flex;
  align-items: center;
  gap: 12px;
}

.wc-team {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.wc-team--home {
  justify-content: flex-end;
  text-align: right;
}

.wc-team--away {
  justify-content: flex-start;
  text-align: left;
}

.wc-team-flag {
  font-size: 20px;
  line-height: 1;
  flex-shrink: 0;
}

.wc-team-name {
  font-size: 14px;
  font-weight: 700;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 120px;
}

.wc-team-code {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  flex-shrink: 0;
}

.wc-score-center {
  flex-shrink: 0;
  text-align: center;
  min-width: 80px;
}

.wc-score-display {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.wc-score-num {
  font-size: 26px;
  font-weight: 800;
  color: var(--text-primary);
  letter-spacing: -0.02em;
  tabular-nums: true;
  line-height: 1;
}

.wc-score-sep {
  font-size: 18px;
  font-weight: 300;
  color: var(--text-muted);
  line-height: 1;
}

.wc-match-time {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.wc-time-date {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
}

.wc-time-clock {
  font-size: 18px;
  font-weight: 700;
  color: var(--text-primary);
  tabular-nums: true;
  line-height: 1.2;
}

.wc-handicap-info {
  margin-top: 6px;
}

.wc-hc-label {
  font-size: 10px;
  font-weight: 600;
  color: var(--text-muted);
  background: rgba(100, 116, 139, 0.08);
  padding: 2px 8px;
  border-radius: 4px;
}

.wc-match-actions {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--border-subtle);
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  align-items: center;
}
</style>
