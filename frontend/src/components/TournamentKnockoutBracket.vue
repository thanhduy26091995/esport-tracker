<template>
  <div class="bracket">
    <!-- Top 4: Semis -->
    <template v-if="knockoutSize !== 2">
      <div class="bracket-row">
        <div class="bracket-col">
          <p class="stage-label">{{ t('tournaments.knockout.semi1', 'Semi-final 1') }}</p>
          <BracketCard :match="semi1" :teams="teams" :label="semiLabel(semi1)" />
        </div>
        <div class="bracket-col">
          <p class="stage-label">{{ t('tournaments.knockout.semi2', 'Semi-final 2') }}</p>
          <BracketCard :match="semi2" :teams="teams" :label="semiLabel(semi2)" />
        </div>
      </div>
    </template>

    <!-- Final + 3rd place (top 4) or just Final (top 2) -->
    <div :class="knockoutSize === 2 ? 'bracket-row bracket-row--single' : 'bracket-row'">
      <template v-if="knockoutSize !== 2">
        <div class="bracket-col">
          <p class="stage-label stage-label--3rd">{{ t('tournaments.knockout.thirdPlace', '3rd Place') }}</p>
          <BracketCard :match="thirdPlace" :teams="teams" />
        </div>
      </template>
      <div class="bracket-col">
        <p class="stage-label stage-label--final">{{ t('tournaments.knockout.final', 'Final') }}</p>
        <BracketCard :match="finalMatch" :teams="teams" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, defineComponent, h } from 'vue'
import { useI18n } from 'vue-i18n'
import type { TournamentMatch, TournamentTeam } from '@/types/tournament'

const props = defineProps<{
  matches: TournamentMatch[]
  teams: TournamentTeam[]
  knockoutSize?: number
}>()

const { t } = useI18n()

const semi1      = computed(() => props.matches.find(m => m.stage === 'semi' && m.match_order === 1) ?? null)
const semi2      = computed(() => props.matches.find(m => m.stage === 'semi' && m.match_order === 2) ?? null)
const finalMatch = computed(() => props.matches.find(m => m.stage === 'final') ?? null)
const thirdPlace = computed(() => props.matches.find(m => m.stage === 'third_place') ?? null)

function semiLabel(match: TournamentMatch | null): string {
  if (!match) return ''
  const t1 = props.teams.find(t => t.id === match.team1_team_id)
  const t2 = props.teams.find(t => t.id === match.team2_team_id)
  if (!t1 || !t2) return ''
  return `S${indexOf(t1)} vs S${indexOf(t2)}`
}

function indexOf(team: TournamentTeam): number {
  // seed is determined by standings order; use team list order as fallback
  return props.teams.indexOf(team) + 1
}

// Inline sub-component to render a single bracket card
const BracketCard = defineComponent({
  props: {
    match: { type: Object as () => TournamentMatch | null, default: null },
    teams: { type: Array as () => TournamentTeam[], default: () => [] },
    label: { type: String, default: '' },
  },
  setup(p) {
    const { t } = useI18n()

    const team1 = computed(() => p.match ? p.teams.find(t => t.id === p.match!.team1_team_id) : null)
    const team2 = computed(() => p.match ? p.teams.find(t => t.id === p.match!.team2_team_id) : null)

    const teamName = (team: TournamentTeam | null | undefined) =>
      team ? `${team.player1?.name} & ${team.player2?.name}` : t('tournaments.knockout.tbd', 'TBD')

    return () => {
      if (!p.match) {
        return h('div', { class: 'bracket-card bracket-card--pending' }, [
          h('span', { class: 'tbd' }, t('tournaments.knockout.tbd', 'TBD')),
        ])
      }

      const isCompleted = p.match.status === 'completed'
      const w = p.match.effective_winner

      return h('div', { class: 'bracket-card' }, [
        // Team 1 row
        h('div', { class: ['bracket-team', w === 1 ? 'bracket-team--winner' : w === 2 ? 'bracket-team--loser' : ''] }, [
          h('span', { class: 'bracket-team-name' }, teamName(team1.value)),
          isCompleted ? h('span', { class: 'bracket-score' }, p.match.actual_score1) : null,
        ]),
        h('div', { class: 'bracket-sep' }),
        // Team 2 row
        h('div', { class: ['bracket-team', w === 2 ? 'bracket-team--winner' : w === 1 ? 'bracket-team--loser' : ''] }, [
          h('span', { class: 'bracket-team-name' }, teamName(team2.value)),
          isCompleted ? h('span', { class: 'bracket-score' }, p.match.actual_score2) : null,
        ]),
      ])
    }
  },
})
</script>

<style scoped>
.bracket {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.bracket-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.bracket-row--single {
  grid-template-columns: 1fr;
  max-width: 320px;
  margin: 0 auto;
}

.bracket-col {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.stage-label {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted, #9ca3af);
  margin: 0;
}

.stage-label--final {
  color: var(--el-color-warning);
}

.stage-label--3rd {
  color: var(--el-color-info);
}

/* BracketCard styles (global because the component is created with h()) */
</style>

<style>
.bracket-card {
  border: 1px solid var(--el-border-color, #dcdfe6);
  border-radius: 6px;
  overflow: hidden;
  background: var(--el-bg-color, #fff);
}

.bracket-card--pending {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  color: var(--text-muted, #9ca3af);
  font-size: 13px;
}

.bracket-team {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  gap: 8px;
  font-size: 13px;
}

.bracket-team--winner {
  background: color-mix(in srgb, var(--el-color-success) 10%, transparent);
  font-weight: 700;
  color: var(--el-color-success);
}

.bracket-team--loser {
  color: var(--text-muted, #9ca3af);
}

.bracket-team-name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.bracket-score {
  font-weight: 700;
  font-size: 16px;
  min-width: 20px;
  text-align: right;
}

.bracket-sep {
  height: 1px;
  background: var(--el-border-color, #dcdfe6);
}

.tbd {
  font-size: 13px;
  font-style: italic;
}
</style>
