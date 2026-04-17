<template>
  <div class="page-wrapper">
    <div class="page-container">
      <!-- Header -->
      <div class="page-header">
        <div class="page-header-left">
          <el-button text @click="router.push('/tournaments')" :icon="ArrowLeft">{{ t('common.back') }}</el-button>
          <div>
            <h1 class="page-title">{{ store.currentTournament?.name ?? '…' }}</h1>
            <div class="header-tags" v-if="store.currentTournament">
              <el-tag size="small">{{ getMatchTypeLabel(store.currentTournament.match_type) }}</el-tag>
              <el-tag
                :type="store.currentTournament.status === 'completed' ? 'success' : 'primary'"
                size="small"
              >
                {{ getTournamentStatusLabel(store.currentTournament.status) }}
              </el-tag>
              <el-tag
                :type="store.currentTournament.affects_score ? 'warning' : 'info'"
                size="small"
              >
                {{ getTournamentAffectsScoreLabel(store.currentTournament.affects_score) }}
              </el-tag>
              <el-tag v-if="store.currentTournament.entry_fee > 0" type="default" size="small">
                {{ t('tournaments.detail.entryFee', { amount: formatVND(store.currentTournament.entry_fee) }) }}
              </el-tag>
            </div>
          </div>
        </div>
        <el-button
          v-if="store.currentTournament?.status === 'active'"
          type="success"
          size="large"
          :loading="store.loading"
          :icon="CircleCheck"
          @click="handleComplete"
          plain
        >
          {{ t('tournaments.detail.completeButton') }}
        </el-button>
      </div>

      <div v-if="store.loading && !store.currentTournament" class="text-center py-12">
        <el-icon class="is-loading" :size="32"><Loading /></el-icon>
      </div>

      <template v-else-if="store.currentTournament">
        <div class="grid grid-cols-1 lg:grid-cols-3 gap-6 detail-layout">
          <!-- Left: Schedule -->
          <div class="lg:col-span-2">
            <div v-for="round in groupedRounds" :key="round.number" class="mb-6">
              <h3 class="section-title">{{ t('tournaments.detail.round', { number: round.number }) }}</h3>
              <div class="round-matches">
              <div
                v-for="match in round.matches"
                :key="match.id"
                class="match-card card"
              >
                <div class="card-body">
                  <!-- Teams row -->
                  <div class="match-teams">
                    <div class="team team--left">
                      <div class="team-players">
                        <span
                          v-for="pid in getTeam1Ids(match)"
                          :key="pid"
                          class="player-name"
                        >
                          {{ getPlayerName(pid) }}
                          <PlayerTierBadge :tier="getPlayerTier(pid)" />
                        </span>
                      </div>
                      <div v-if="match.handicap_team1 !== 0" class="handicap-badge handicap--red">
                        -{{ match.handicap_team1 }}
                      </div>
                    </div>

                    <div class="match-vs">
                      <template v-if="match.status === 'completed'">
                        <span
                          class="score-display"
                          :class="{ 'score--winner': match.effective_winner === 1 }"
                        >
                          {{ match.actual_score1 }}
                        </span>
                        <span class="vs-sep">:</span>
                        <span
                          class="score-display"
                          :class="{ 'score--winner': match.effective_winner === 2 }"
                        >
                          {{ match.actual_score2 }}
                        </span>
                      </template>
                      <template v-else>
                        <span class="vs-text">{{ t('common.vs') }}</span>
                      </template>
                    </div>

                    <div class="team team--right">
                      <div v-if="match.handicap_team2 !== 0" class="handicap-badge handicap--blue">
                        -{{ match.handicap_team2 }}
                      </div>
                      <div class="team-players">
                        <span
                          v-for="pid in getTeam2Ids(match)"
                          :key="pid"
                          class="player-name"
                        >
                          {{ getPlayerName(pid) }}
                          <PlayerTierBadge :tier="getPlayerTier(pid)" />
                        </span>
                      </div>
                    </div>
                  </div>

                  <!-- Completed: winner label -->
                  <div v-if="match.status === 'completed'" class="match-result-label">
                    <el-tag v-if="match.effective_winner === 0" type="info" size="small">{{ t('tournaments.detail.draw') }}</el-tag>
                    <el-tag v-else-if="match.effective_winner === 1" type="success" size="small">
                      {{ t('tournaments.detail.wins', { team: getTeam1Label(match) }) }}
                    </el-tag>
                    <el-tag v-else type="success" size="small">
                      {{ t('tournaments.detail.wins', { team: getTeam2Label(match) }) }}
                    </el-tag>
                  </div>

                  <!-- Pending: score input -->
                  <div
                    v-else-if="store.currentTournament?.status === 'active'"
                    class="result-input-row"
                  >
                    <div class="score-input-grid">
                      <div class="score-input-cell">
                        <span class="score-input-label">{{ getTeam1Label(match) }}</span>
                        <el-input-number
                          v-model="scoreInputs[match.id].score1"
                          :min="0"
                          :max="99"
                          size="small"
                          class="score-input-number"
                        />
                      </div>
                      <div class="score-input-cell">
                        <span class="score-input-label">{{ getTeam2Label(match) }}</span>
                        <el-input-number
                          v-model="scoreInputs[match.id].score2"
                          :min="0"
                          :max="99"
                          size="small"
                          class="score-input-number"
                        />
                      </div>
                    </div>

                    <!-- Effective winner preview -->
                    <div class="result-input-actions">
                      <el-tag
                        v-if="effectiveWinnerPreview(match) !== null"
                        :type="effectiveWinnerPreview(match) === 0 ? 'info' : 'warning'"
                        size="small"
                      >
                        →
                        <template v-if="effectiveWinnerPreview(match) === 0">{{ t('tournaments.detail.draw') }}</template>
                        <template v-else-if="effectiveWinnerPreview(match) === 1">
                          {{ t('tournaments.detail.wins', { team: getTeam1Label(match) }) }}
                        </template>
                        <template v-else>
                          {{ t('tournaments.detail.wins', { team: getTeam2Label(match) }) }}
                        </template>
                      </el-tag>

                      <el-button
                        size="small"
                        type="primary"
                        :loading="store.loading"
                        @click="handleRecordResult(match)"
                      >
                        {{ t('tournaments.detail.submitResult') }}
                      </el-button>
                    </div>
                  </div>
                </div>
              </div>
              </div>
            </div>
          </div>

          <!-- Right: Participants + Standings -->
          <div>
            <!-- Participants -->
            <div class="card mb-4">
              <div class="card-body">
                <h3 class="section-title">{{ t('tournaments.detail.participants', { count: store.currentTournament.participants.length }) }}</h3>
                <div
                  v-for="p in store.currentTournament.participants"
                  :key="p.id"
                  class="participant-row"
                >
                  {{ p.user?.name }}
                  <PlayerTierBadge :tier="p.tier_snapshot || p.user?.tier || 'normal'" />
                  <span
                    v-if="p.handicap_rate_snapshot > 0"
                    style="font-size: 11px; color: #909399; margin-left: 4px"
                  >
                    -{{ p.handicap_rate_snapshot }}
                  </span>
                </div>
              </div>
            </div>

            <!-- Standings -->
            <div class="card">
              <div class="card-body">
                <h3 class="section-title">{{ t('tournaments.detail.standings') }}</h3>
                <div class="standings-table-wrap">
                  <el-table :data="standings" size="small" class="standings-table">
                  <el-table-column label="#" type="index" width="36" />
                  <el-table-column :label="t('tournaments.detail.player')" min-width="110">
                    <template #default="{ row }">
                      <div class="flex items-center gap-1 flex-wrap">
                        <span>{{ row.user?.name }}</span>
                        <PlayerTierBadge :tier="row.user?.tier || 'normal'" />
                      </div>
                    </template>
                  </el-table-column>
                  <el-table-column prop="wins" :label="t('tournaments.detail.winsShort')" width="38" align="center" />
                  <el-table-column prop="draws" :label="t('tournaments.detail.drawsShort')" width="38" align="center" />
                  <el-table-column prop="losses" :label="t('tournaments.detail.lossesShort')" width="38" align="center" />
                  <el-table-column :label="t('tournaments.detail.goalDifference')" width="50" align="center">
                    <template #default="{ row }">
                      <span :class="row.goals_for - row.goals_against >= 0 ? 'text-success' : 'text-danger'">
                        {{ row.goals_for - row.goals_against >= 0 ? '+' : '' }}{{ row.goals_for - row.goals_against }}
                      </span>
                    </template>
                  </el-table-column>
                  <el-table-column prop="points" :label="t('tournaments.detail.pointsShort')" width="56" align="center">
                    <template #default="{ row }">
                      <strong>{{ row.points }}</strong>
                    </template>
                  </el-table-column>
                  </el-table>
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ArrowLeft, CircleCheck, Loading } from '@element-plus/icons-vue'
import { ElMessageBox } from 'element-plus'
import { useTournamentStore } from '@/stores/tournamentStore'
import PlayerTierBadge from '@/components/PlayerTierBadge.vue'
import type { Tournament, TournamentMatch, TournamentStanding } from '@/types/tournament'
import { formatVND } from '@/utils/formatters'
import { getMatchTypeLabel, getTournamentAffectsScoreLabel, getTournamentStatusLabel } from '@/utils/tournamentLabels'

const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const store = useTournamentStore()

const tournamentId = route.params.id as string

onMounted(() => store.fetchTournament(tournamentId))

// ── Score inputs keyed by match id ──────────────────────────────────────────
const scoreInputs = reactive<Record<string, { score1: number; score2: number }>>({})

const ensureInput = (matchId: string) => {
  if (!scoreInputs[matchId]) scoreInputs[matchId] = { score1: 0, score2: 0 }
}

// ── Helpers ──────────────────────────────────────────────────────────────────
const getTeam1Ids = (m: TournamentMatch) =>
  [m.team1_player1_id, m.team1_player2_id].filter(Boolean) as string[]

const getTeam2Ids = (m: TournamentMatch) =>
  [m.team2_player1_id, m.team2_player2_id].filter(Boolean) as string[]

const participantMap = computed(() => {
  const map = new Map<string, { name: string; tier: string }>()
  for (const p of store.currentTournament?.participants ?? []) {
    map.set(p.user_id, { name: p.user?.name ?? p.user_id, tier: p.tier_snapshot || p.user?.tier || 'normal' })
  }
  return map
})

const getPlayerName = (id: string) => participantMap.value.get(id)?.name ?? id
const getPlayerTier = (id: string) => participantMap.value.get(id)?.tier ?? 'normal'

const getTeam1Label = (m: TournamentMatch) =>
  getTeam1Ids(m).map(getPlayerName).join(' & ')

const getTeam2Label = (m: TournamentMatch) =>
  getTeam2Ids(m).map(getPlayerName).join(' & ')

// ── Grouped rounds ───────────────────────────────────────────────────────────
const groupedRounds = computed(() => {
  const tournament = store.currentTournament
  if (!tournament) return []

  const rounds = new Map<number, TournamentMatch[]>()
  for (const m of tournament.matches) {
    if (!rounds.has(m.round)) rounds.set(m.round, [])
    rounds.get(m.round)!.push(m)
    ensureInput(m.id)
  }

  return Array.from(rounds.entries())
    .sort(([a], [b]) => a - b)
    .map(([number, matches]) => ({
      number,
      matches: matches.sort((a, b) => a.match_order - b.match_order),
    }))
})

// ── Effective winner preview (before submit) ─────────────────────────────────
const effectiveWinnerPreview = (m: TournamentMatch): number | null => {
  const input = scoreInputs[m.id]
  if (!input) return null

  const s1 = input.score1 - m.handicap_team1
  const s2 = input.score2 - m.handicap_team2

  if (s1 > s2) return 1
  if (s2 > s1) return 2
  return 0
}

// ── Record result ─────────────────────────────────────────────────────────────
const handleRecordResult = async (match: TournamentMatch) => {
  const input = scoreInputs[match.id]
  if (!input) return
  await store.recordResult(tournamentId, match.id, {
    actual_score1: input.score1,
    actual_score2: input.score2,
    recorded_by: 'admin',
  })
}

// ── Complete tournament ───────────────────────────────────────────────────────
const handleComplete = () => {
  ElMessageBox.confirm(
    t('tournaments.detail.completeConfirm'),
    t('tournaments.detail.completeTitle'),
    {
      confirmButtonText: t('tournaments.detail.completeButton'),
      cancelButtonText: t('common.cancel'),
      type: 'warning',
    }
  )
    .then(() => store.completeTournament(tournamentId))
    .catch(() => {})
}

// ── Standings calculation ─────────────────────────────────────────────────────
const standings = computed((): TournamentStanding[] => {
  const tournament = store.currentTournament
  if (!tournament) return []
  return computeStandings(tournament)
})

function computeStandings(tournament: Tournament): TournamentStanding[] {
  const map = new Map<string, TournamentStanding>()

  for (const p of tournament.participants) {
    map.set(p.user_id, {
      user: p.user,
      wins: 0,
      draws: 0,
      losses: 0,
      goals_for: 0,
      goals_against: 0,
      points: 0,
    })
  }

  for (const m of tournament.matches) {
    if (m.status !== 'completed' || m.actual_score1 === undefined) continue

    const team1Ids = [m.team1_player1_id, m.team1_player2_id].filter(Boolean) as string[]
    const team2Ids = [m.team2_player1_id, m.team2_player2_id].filter(Boolean) as string[]
    const s1 = m.actual_score1
    const s2 = m.actual_score2!
    const winner = m.effective_winner

    for (const id of team1Ids) {
      const s = map.get(id)
      if (!s) continue
      s.goals_for += s1
      s.goals_against += s2
      if (winner === 1) { s.wins++; s.points += 3 }
      else if (winner === 0) { s.draws++; s.points += 1 }
      else { s.losses++ }
    }

    for (const id of team2Ids) {
      const s = map.get(id)
      if (!s) continue
      s.goals_for += s2
      s.goals_against += s1
      if (winner === 2) { s.wins++; s.points += 3 }
      else if (winner === 0) { s.draws++; s.points += 1 }
      else { s.losses++ }
    }
  }

  return Array.from(map.values()).sort(
    (a, b) =>
      b.points - a.points ||
      (b.goals_for - b.goals_against) - (a.goals_for - a.goals_against)
  )
}
</script>

<style scoped>
.section-title {
  font-size: 14px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 12px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.header-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 6px;
}

.round-matches {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
}

.match-teams {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  align-items: center;
  column-gap: 14px;
  row-gap: 8px;
}

.team {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.team--left {
  grid-column: 1;
  grid-row: 1;
  justify-content: center;
  text-align: center;
}

.team--right {
  grid-column: 2;
  grid-row: 1;
  justify-content: center;
  text-align: center;
}

.team-players {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
  align-items: center;
}

.player-name {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  font-size: 13px;
  font-weight: 500;
  line-height: 1.4;
}

.handicap-badge {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 5px;
  border-radius: 4px;
  white-space: nowrap;
}

.handicap--red {
  background: #fee2e2;
  color: #b91c1c;
}

.handicap--blue {
  background: #dbeafe;
  color: #1d4ed8;
}

.match-vs {
  grid-column: 1 / -1;
  grid-row: 2;
  display: flex;
  align-items: center;
  gap: 4px;
  min-width: 84px;
  justify-content: center;
}

.vs-text {
  font-size: 12px;
  font-weight: 700;
  color: var(--text-muted);
}

.vs-sep {
  font-size: 18px;
  font-weight: 700;
  color: var(--text-muted);
}

.score-display {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-secondary);
  min-width: 24px;
  text-align: center;
}

.score--winner {
  color: var(--color-success, #15803d);
}

.match-result-label {
  margin-top: 8px;
  text-align: center;
}

.result-input-row {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid var(--border-subtle);
  gap: 10px;
}

.score-input-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.score-input-cell {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.score-input-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  text-align: center;
  line-height: 1.35;
}

.score-input-number {
  width: 100%;
  max-width: 120px;
}

.result-input-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-wrap: wrap;
  gap: 8px;
}

.participant-row {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 0;
  border-bottom: 1px solid var(--border-subtle);
  font-size: 13px;
}

.participant-row:last-child {
  border-bottom: none;
}

.text-success {
  color: var(--color-success, #15803d);
}

.text-danger {
  color: var(--color-danger, #b91c1c);
}

.text-muted {
  color: var(--text-muted);
}

.standings-table-wrap {
  width: 100%;
  overflow-x: auto;
}

.standings-table {
  min-width: 360px;
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: center;
    text-align: center;
  }

  .header-tags {
    justify-content: center;
  }

  .match-teams {
    grid-template-columns: 1fr;
    gap: 8px;
  }

  .team,
  .team--left,
  .team--right {
    justify-content: center;
    text-align: center;
  }

  .team-players {
    align-items: center;
  }

  .player-name {
    justify-content: center;
    flex-wrap: wrap;
  }

  .match-vs {
    order: -1;
    min-width: 0;
  }
}
</style>
