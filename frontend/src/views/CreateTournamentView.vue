<template>
  <div class="page-wrapper">
    <div class="page-container" style="max-width: 700px">
      <div class="page-header">
        <div class="page-header-left">
          <el-button text @click="router.back()" :icon="ArrowLeft">{{ t('common.back') }}</el-button>
          <h1 class="page-title">{{ t('tournaments.form.title') }}</h1>
        </div>
      </div>

      <div class="card">
        <div class="card-body">
          <el-form
            ref="formRef"
            :model="form"
            label-position="top"
            @submit.prevent="handleSubmit"
          >
            <el-form-item
              :label="t('tournaments.form.name')"
              prop="name"
              :rules="[{ required: true, message: t('validation.tournamentNameRequired'), trigger: 'blur' }]"
            >
              <el-input v-model="form.name" :placeholder="t('tournaments.form.namePlaceholder')" />
            </el-form-item>

            <el-form-item :label="t('tournaments.form.format', 'Format')">
              <el-radio-group v-model="form.format" @change="onFormatChange">
                <el-radio-button value="classic">{{ t('tournaments.format.classic', 'Classic') }}</el-radio-button>
                <el-radio-button value="round_robin_top4">{{ t('tournaments.format.roundRobinTop4', 'Round Robin + Top 4') }}</el-radio-button>
              </el-radio-group>
            </el-form-item>

            <!-- Classic: match type selector -->
            <el-form-item v-if="form.format === 'classic'" :label="t('tournaments.form.matchType')">
              <el-radio-group v-model="form.match_type">
                <el-radio-button value="1v1">{{ t('matches.types.oneVsOne') }}</el-radio-button>
                <el-radio-button value="2v2">{{ t('matches.types.twoVsTwo') }}</el-radio-button>
              </el-radio-group>
            </el-form-item>

            <!-- Classic: player list -->
            <el-form-item v-if="form.format === 'classic'" :label="t('tournaments.form.players')">
              <div class="player-list-wrap">
                <div class="player-list">
                  <div
                    v-for="user in userStore.users.filter(u => u.is_active)"
                    :key="user.id"
                    class="player-row"
                    :class="{ 'player-row--selected': form.player_ids.includes(user.id) }"
                    @click="toggleClassicPlayer(user.id)"
                  >
                    <span class="player-row-check">
                      <svg v-if="form.player_ids.includes(user.id)" viewBox="0 0 12 12" width="12" height="12"><path d="M1.5 6L4.5 9L10.5 3" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" fill="none"/></svg>
                    </span>
                    <span class="player-row-name">{{ user.name }}</span>
                    <PlayerTierBadge :tier="user.tier || 'normal'" />
                    <span v-if="user.handicap_rate > 0" class="player-row-handicap">-{{ user.handicap_rate }}</span>
                  </div>
                </div>
                <div class="player-list-footer">
                  <span class="player-list-count">
                    {{ t('tournaments.form.selectedCount', { count: form.player_ids.length }) }}
                    <el-tag
                      v-if="form.match_type === '2v2' && form.player_ids.length % 2 !== 0"
                      type="warning"
                      size="small"
                      style="margin-left: 6px"
                    >
                      {{ t('tournaments.form.evenPlayerCountRequired') }}
                    </el-tag>
                  </span>
                  <el-button :icon="Plus" size="small" @click="showQuickCreatePlayer = true">
                    {{ t('players.quickCreate') }}
                  </el-button>
                </div>
              </div>
            </el-form-item>

            <!-- Round Robin Top 4: team count -->
            <el-form-item v-if="form.format === 'round_robin_top4'" :label="t('tournaments.form.teamCount', 'Number of teams')">
              <el-radio-group v-model="teamCount" @change="onTeamCountChange">
                <el-radio-button :value="4">4</el-radio-button>
                <el-radio-button :value="5">5</el-radio-button>
                <el-radio-button :value="6">6</el-radio-button>
                <el-radio-button :value="7">7</el-radio-button>
                <el-radio-button :value="8">8</el-radio-button>
              </el-radio-group>
              <div class="el-form-item__helper mt-1" style="font-size:12px; color: var(--text-muted)">
                {{ t('tournaments.form.teamCountHint', { n: teamCount, p: teamCount * 2 }, `${teamCount} teams · ${teamCount * 2} players · ${teamCount * (teamCount - 1) / 2} matches`) }}
              </div>
            </el-form-item>

            <!-- Round Robin: knockout size -->
            <el-form-item v-if="form.format === 'round_robin_top4'" :label="t('tournaments.form.knockoutSize', 'Knockout stage')">
              <el-radio-group v-model="knockoutSize">
                <el-radio-button :value="2">{{ t('tournaments.form.knockoutSizeTop2', 'Top 2 · Final only') }}</el-radio-button>
                <el-radio-button :value="4" :disabled="teamCount < 4">{{ t('tournaments.form.knockoutSizeTop4', 'Top 4 · Semis + Final') }}</el-radio-button>
              </el-radio-group>
              <div class="el-form-item__helper mt-1" style="font-size:12px; color: var(--text-muted)">
                {{ knockoutSize === 2
                  ? t('tournaments.form.knockoutSizeTop2Hint', 'Top 2 teams play a single final match')
                  : t('tournaments.form.knockoutSizeTop4Hint', 'Top 4 teams — 1st vs 4th, 2nd vs 3rd, then final + 3rd place') }}
              </div>
            </el-form-item>

            <!-- Round Robin Top 4: player pool + randomize -->
            <el-form-item v-if="form.format === 'round_robin_top4'" :label="t('tournaments.form.playerPool', 'Player pool')">
              <div class="player-list-wrap">
                <div class="player-list">
                  <div
                    v-for="user in userStore.users.filter(u => u.is_active)"
                    :key="user.id"
                    class="player-row"
                    :class="{ 'player-row--selected': playerPool.includes(user.id) }"
                    @click="togglePoolPlayer(user.id)"
                  >
                    <span class="player-row-check">
                      <svg v-if="playerPool.includes(user.id)" viewBox="0 0 12 12" width="12" height="12"><path d="M1.5 6L4.5 9L10.5 3" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" fill="none"/></svg>
                    </span>
                    <span class="player-row-name">{{ user.name }}</span>
                    <PlayerTierBadge :tier="user.tier || 'normal'" />
                    <span v-if="user.handicap_rate > 0" class="player-row-handicap">-{{ user.handicap_rate }}</span>
                  </div>
                </div>
                <div class="player-list-footer">
                  <span class="player-list-count" :class="{ 'count--ready': playerPool.length === teamCount * 2, 'count--over': playerPool.length > teamCount * 2 }">
                    {{ t('tournaments.form.playerPoolCount', { selected: playerPool.length, needed: teamCount * 2 }) }}
                  </span>
                  <el-button
                    plain
                    :disabled="playerPool.length !== teamCount * 2"
                    type="success"
                    size="small"
                    @click="randomizeTeams"
                  >
                    {{ t('tournaments.form.randomize', 'Randomize teams') }}
                  </el-button>
                </div>
              </div>
            </el-form-item>

            <!-- Round Robin Top 4: team slots (manual override after randomize) -->
            <el-form-item v-if="form.format === 'round_robin_top4'" :label="t('tournaments.form.teams', 'Teams')">
              <div class="team-slots">
                <div
                  v-for="(team, idx) in form.teams"
                  :key="idx"
                  class="team-slot"
                >
                  <span class="team-slot-label">{{ t('tournaments.form.teamNumber', { n: idx + 1 }, `Team ${idx + 1}`) }}</span>
                  <el-select
                    v-model="team.player1_id"
                    :placeholder="t('tournaments.form.selectPlayer', 'Player 1')"
                    size="small"
                    class="team-slot-select"
                    filterable
                  >
                    <el-option
                      v-for="user in availablePlayers(idx, 'p1')"
                      :key="user.id"
                      :value="user.id"
                      :label="user.name"
                    />
                  </el-select>
                  <span class="team-slot-amp">&amp;</span>
                  <el-select
                    v-model="team.player2_id"
                    :placeholder="t('tournaments.form.selectPlayer', 'Player 2')"
                    size="small"
                    class="team-slot-select"
                    filterable
                  >
                    <el-option
                      v-for="user in availablePlayers(idx, 'p2')"
                      :key="user.id"
                      :value="user.id"
                      :label="user.name"
                    />
                  </el-select>
                </div>
                <el-tag v-if="teamValidationError" type="warning" size="small" class="mt-2">
                  {{ teamValidationError }}
                </el-tag>
              </div>
            </el-form-item>

            <el-form-item :label="t('tournaments.form.affectsScore')">
              <el-switch
                v-model="form.affects_score"
                :active-text="t('tournaments.form.affectsScoreActive')"
                :inactive-text="t('tournaments.form.affectsScoreInactive')"
              />
            </el-form-item>

            <el-form-item :label="t('tournaments.form.entryFee')">
              <el-input-number
                v-model="form.entry_fee"
                :min="0"
                :step="10000"
                :placeholder="t('tournaments.form.entryFeePlaceholder')"
              />
            </el-form-item>

            <el-form-item>
              <el-button
                type="primary"
                native-type="submit"
                :loading="store.loading"
                size="large"
              >
                {{ t('tournaments.form.submit') }}
              </el-button>
              <el-button @click="router.push('/tournaments')" size="large">{{ t('common.cancel') }}</el-button>
            </el-form-item>
          </el-form>

          <UserForm
            v-model="showQuickCreatePlayer"
            :loading="quickCreateLoading"
            @submit="handlePlayerCreated"
            @cancel="showQuickCreatePlayer = false"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ArrowLeft, Plus } from '@element-plus/icons-vue'
import type { FormInstance } from 'element-plus'
import { useTournamentStore } from '@/stores/tournamentStore'
import { useUserStore } from '@/stores/userStore'
import PlayerTierBadge from '@/components/PlayerTierBadge.vue'
import UserForm from '@/components/user/UserForm.vue'
import type { TournamentFormat } from '@/types/tournament'

const router = useRouter()
const { t } = useI18n()
const store = useTournamentStore()
const userStore = useUserStore()
const formRef = ref<FormInstance>()
const showQuickCreatePlayer = ref(false)
const quickCreateLoading = ref(false)

const teamCount = ref(5)
const knockoutSize = ref(4)
const playerPool = ref<string[]>([])

const form = ref({
  name: '',
  format: 'classic' as TournamentFormat,
  match_type: '1v1' as '1v1' | '2v2',
  player_ids: [] as string[],
  teams: Array.from({ length: 5 }, () => ({ player1_id: '', player2_id: '' })),
  affects_score: true,
  entry_fee: 0,
})

onMounted(() => userStore.fetchUsers())

function onFormatChange() {
  teamCount.value = 5
  knockoutSize.value = 4
  playerPool.value = []
  form.value.teams = Array.from({ length: 5 }, () => ({ player1_id: '', player2_id: '' }))
  form.value.player_ids = []
}

function onTeamCountChange() {
  playerPool.value = []
  form.value.teams = Array.from({ length: teamCount.value }, () => ({ player1_id: '', player2_id: '' }))
}

function toggleClassicPlayer(id: string) {
  const idx = form.value.player_ids.indexOf(id)
  if (idx === -1) form.value.player_ids.push(id)
  else form.value.player_ids.splice(idx, 1)
}

function togglePoolPlayer(id: string) {
  const idx = playerPool.value.indexOf(id)
  if (idx === -1) playerPool.value.push(id)
  else playerPool.value.splice(idx, 1)
}

function randomizeTeams() {
  const shuffled = [...playerPool.value].sort(() => Math.random() - 0.5)
  form.value.teams = Array.from({ length: teamCount.value }, (_, i) => ({
    player1_id: shuffled[i * 2] ?? '',
    player2_id: shuffled[i * 2 + 1] ?? '',
  }))
}

// All currently selected player IDs across all team slots
const selectedPlayerIds = computed(() => {
  const ids = new Set<string>()
  form.value.teams.forEach(t => {
    if (t.player1_id) ids.add(t.player1_id)
    if (t.player2_id) ids.add(t.player2_id)
  })
  return ids
})

// Available players for a given slot (exclude already selected, except own slot)
function availablePlayers(teamIdx: number, slot: 'p1' | 'p2') {
  const own = slot === 'p1' ? form.value.teams[teamIdx].player1_id : form.value.teams[teamIdx].player2_id
  return userStore.users.filter(u => {
    if (!u.is_active) return false
    if (u.id === own) return true
    return !selectedPlayerIds.value.has(u.id)
  })
}

const teamValidationError = computed(() => {
  if (form.value.format !== 'round_robin_top4') return ''
  const allFilled = form.value.teams.every(t => t.player1_id && t.player2_id)
  if (!allFilled) return t('tournaments.form.allTeamSlotsFill', 'All teams must be fully assigned')
  return ''
})

const handlePlayerCreated = async (data: { name: string; tier: string; handicap_rate: number }) => {
  quickCreateLoading.value = true
  try {
    await userStore.createUser(data.name, data.tier, data.handicap_rate)
    await userStore.fetchUsers()
    showQuickCreatePlayer.value = false
  } finally {
    quickCreateLoading.value = false
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    if (form.value.format === 'round_robin_top4' && teamValidationError.value) return
    try {
      const payload =
        form.value.format === 'round_robin_top4'
          ? {
              name: form.value.name,
              format: 'round_robin_top4' as TournamentFormat,
              teams: form.value.teams,
              affects_score: form.value.affects_score,
              entry_fee: form.value.entry_fee,
              knockout_size: knockoutSize.value,
            }
          : {
              name: form.value.name,
              format: 'classic' as TournamentFormat,
              match_type: form.value.match_type,
              player_ids: form.value.player_ids,
              affects_score: form.value.affects_score,
              entry_fee: form.value.entry_fee,
            }
      const tournament = await store.createTournament(payload)
      router.push(`/tournaments/${tournament.id}`)
    } catch {}
  })
}
</script>

<style scoped>
/* ── Player list (shared by Classic + Round Robin pool) ── */
.player-list-wrap {
  width: 100%;
}

.player-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 6px;
}

.player-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 10px;
  border: 1px solid var(--el-border-color, #dcdfe6);
  border-radius: 6px;
  cursor: pointer;
  user-select: none;
  transition: border-color 0.15s, background 0.15s;
  background: var(--el-bg-color, #fff);
  min-width: 0;
}

.player-row:hover {
  border-color: var(--el-color-primary, #409eff);
  background: color-mix(in srgb, var(--el-color-primary) 4%, transparent);
}

.player-row--selected {
  border-color: var(--el-color-primary, #409eff);
  background: color-mix(in srgb, var(--el-color-primary) 8%, transparent);
}

.player-row-check {
  flex-shrink: 0;
  width: 16px;
  height: 16px;
  border: 1.5px solid var(--el-border-color-darker, #b0b3ba);
  border-radius: 3px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: border-color 0.15s, background 0.15s;
  background: transparent;
  color: #fff;
}

.player-row--selected .player-row-check {
  background: var(--el-color-primary, #409eff);
  border-color: var(--el-color-primary, #409eff);
}

.player-row-name {
  flex: 1;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.player-row--selected .player-row-name {
  color: var(--el-color-primary, #409eff);
  font-weight: 600;
}

.player-row-handicap {
  font-size: 11px;
  font-weight: 600;
  color: var(--el-color-danger, #f56c6c);
  background: color-mix(in srgb, var(--el-color-danger) 10%, transparent);
  padding: 1px 5px;
  border-radius: 3px;
  flex-shrink: 0;
}

.player-list-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 10px;
  gap: 8px;
  flex-wrap: wrap;
}

.player-list-count {
  font-size: 12px;
  color: var(--text-muted, #9ca3af);
}

.player-list-count.count--ready {
  color: var(--el-color-success, #67c23a);
  font-weight: 600;
}

.player-list-count.count--over {
  color: var(--el-color-danger, #f56c6c);
  font-weight: 600;
}

/* ── Team slots ── */
.team-slots {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.team-slot {
  display: flex;
  align-items: center;
  gap: 8px;
}

.team-slot-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
  min-width: 56px;
}

.team-slot-select {
  flex: 1;
}

.team-slot-amp {
  font-weight: 700;
  color: var(--text-muted, #9ca3af);
}

@media (max-width: 640px) {
  :deep(.el-radio-group) {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  :deep(.el-input-number) {
    width: 100%;
  }

  .player-list {
    grid-template-columns: 1fr;
  }
}
</style>
