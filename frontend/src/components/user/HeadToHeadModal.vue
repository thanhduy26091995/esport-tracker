<template>
  <el-dialog
    :model-value="visible"
    :title="t('dashboard.headToHead.title')"
    width="640px"
    class="h2h-dialog"
    append-to-body
    @update:model-value="(v: boolean) => emit('update:visible', v)"
    @open="onOpen"
  >
    <!-- Player pickers -->
    <div class="h2h-pickers">
      <el-select
        v-model="player1Id"
        :placeholder="t('dashboard.headToHead.selectPlayer1')"
        filterable
        class="h2h-select"
      >
        <el-option
          v-for="u in playerOptions"
          :key="u.id"
          :value="u.id"
          :label="u.is_active ? u.name : `${u.name} (${t('dashboard.headToHead.inactive')})`"
          :disabled="u.id === player2Id"
        />
      </el-select>

      <span class="h2h-vs">{{ t('dashboard.headToHead.vs') }}</span>

      <el-select
        v-model="player2Id"
        :placeholder="t('dashboard.headToHead.selectPlayer2')"
        filterable
        class="h2h-select"
      >
        <el-option
          v-for="u in playerOptions"
          :key="u.id"
          :value="u.id"
          :label="u.is_active ? u.name : `${u.name} (${t('dashboard.headToHead.inactive')})`"
          :disabled="u.id === player1Id"
        />
      </el-select>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="h2h-center">
      <el-icon class="animate-spin" :size="28"><Loading /></el-icon>
    </div>

    <!-- Prompt to pick players -->
    <div v-else-if="!data" class="h2h-center h2h-muted">
      {{ t('dashboard.headToHead.pickPrompt') }}
    </div>

    <!-- Empty: never met -->
    <div v-else-if="data.total_matches === 0" class="h2h-center h2h-muted">
      <el-icon :size="36"><DocumentDelete /></el-icon>
      <p>{{ t('dashboard.headToHead.empty') }}</p>
    </div>

    <!-- Result -->
    <div v-else class="h2h-result">
      <!-- Header: both players -->
      <div class="h2h-players">
        <div class="h2h-player">
          <UserAvatar :avatar-url="data.player1.avatar_url" :name="data.player1.name" size="lg" />
          <span class="h2h-name">{{ data.player1.name }}</span>
          <PlayerTierBadge :tier="data.player1.tier" />
        </div>
        <div class="h2h-score">
          <span class="h2h-wins h2h-wins--p1">{{ data.player1_wins }}</span>
          <span class="h2h-dash">-</span>
          <span class="h2h-wins h2h-wins--p2">{{ data.player2_wins }}</span>
        </div>
        <div class="h2h-player">
          <UserAvatar :avatar-url="data.player2.avatar_url" :name="data.player2.name" size="lg" />
          <span class="h2h-name">{{ data.player2.name }}</span>
          <PlayerTierBadge :tier="data.player2.tier" />
        </div>
      </div>

      <!-- Totals + win rate -->
      <div class="h2h-stats-row">
        <div class="h2h-stat">
          <span class="h2h-stat-val">{{ Math.round(data.player1_win_rate * 100) }}%</span>
          <span class="h2h-stat-label">{{ t('dashboard.headToHead.winRate') }}</span>
        </div>
        <div class="h2h-stat">
          <span class="h2h-stat-val">{{ data.total_matches }}</span>
          <span class="h2h-stat-label">{{ t('dashboard.headToHead.totalMatches') }}</span>
        </div>
        <div class="h2h-stat">
          <span class="h2h-stat-val">{{ Math.round(data.player2_win_rate * 100) }}%</span>
          <span class="h2h-stat-label">{{ t('dashboard.headToHead.winRate') }}</span>
        </div>
      </div>

      <!-- Doughnut chart -->
      <div class="h2h-chart">
        <Doughnut :data="chartData" :options="chartOptions" />
      </div>

      <!-- Streak -->
      <div v-if="streakText" class="h2h-streak">🔥 {{ streakText }}</div>

      <!-- Form (most-recent first) -->
      <div class="h2h-form">
        <span class="h2h-form-label">{{ t('dashboard.headToHead.form') }}</span>
        <span
          v-for="(r, i) in data.form"
          :key="i"
          class="h2h-form-badge"
          :class="r === 'W' ? 'h2h-form-badge--w' : 'h2h-form-badge--l'"
        >{{ r === 'W' ? t('dashboard.headToHead.win') : t('dashboard.headToHead.loss') }}</span>
      </div>

      <!-- Recent matches with lineups -->
      <div class="h2h-recent">
        <span class="h2h-recent-label">{{ t('dashboard.headToHead.recentMatches') }}</span>
        <div v-for="m in data.recent_matches" :key="m.match_id" class="h2h-match">
          <span class="h2h-match-type">{{ m.match_type }}</span>
          <div class="h2h-lineup">
            <span class="h2h-team" :class="{ 'h2h-team--won': m.winner_team === 1 }">
              {{ teamNames(m, 1) }}
            </span>
            <span class="h2h-lineup-vs">{{ t('dashboard.headToHead.vs') }}</span>
            <span class="h2h-team" :class="{ 'h2h-team--won': m.winner_team === 2 }">
              {{ teamNames(m, 2) }}
            </span>
          </div>
          <span class="h2h-match-date">{{ formatDate(m.match_date) }}</span>
        </div>
      </div>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Loading, DocumentDelete } from '@element-plus/icons-vue'
import { Doughnut } from 'vue-chartjs'
import { Chart as ChartJS, ArcElement, Tooltip, Legend } from 'chart.js'
import UserAvatar from '@/components/shared/UserAvatar.vue'
import PlayerTierBadge from '@/components/PlayerTierBadge.vue'
import { userService } from '@/services/userService'
import { formatDate } from '@/utils/date'
import type { UserWithStats } from '@/types/user'
import type { HeadToHeadResponse, H2HMatch } from '@/types/headToHead'

ChartJS.register(ArcElement, Tooltip, Legend)

const props = defineProps<{
  visible: boolean
  players: UserWithStats[]
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
}>()

const { t } = useI18n()

const player1Id = ref<string>('')
const player2Id = ref<string>('')
const data = ref<HeadToHeadResponse | null>(null)
const loading = ref(false)

// Active players first, then inactive — already ordered by the API, but guard anyway.
const playerOptions = computed(() => props.players)

function onOpen() {
  // Reset selections each time the dialog opens for a clean slate.
  player1Id.value = ''
  player2Id.value = ''
  data.value = null
}

watch([player1Id, player2Id], async ([p1, p2]) => {
  if (!p1 || !p2 || p1 === p2) {
    data.value = null
    return
  }
  loading.value = true
  try {
    data.value = await userService.getHeadToHead(p1, p2)
  } catch {
    data.value = null
  } finally {
    loading.value = false
  }
})

const chartData = computed(() => ({
  labels: [data.value?.player1.name ?? '', data.value?.player2.name ?? ''],
  datasets: [{
    data: [data.value?.player1_wins ?? 0, data.value?.player2_wins ?? 0],
    backgroundColor: ['#3b82f6', '#ef4444'],
    borderWidth: 0,
  }],
}))

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: { legend: { position: 'bottom' as const } },
  cutout: '62%',
}

const streakText = computed(() => {
  const d = data.value
  if (!d || !d.current_streak.player_id || d.current_streak.count < 2) return ''
  const name = d.current_streak.player_id === d.player1.id ? d.player1.name : d.player2.name
  return t('dashboard.headToHead.streak', { name, count: d.current_streak.count })
})

function teamNames(match: H2HMatch, team: number): string {
  const names = match.participants
    .filter(p => p.team === team)
    .map(p => p.name)
  return names.length ? names.join(' + ') : '?'
}
</script>

<style scoped>
.h2h-pickers {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}
.h2h-select { flex: 1; }
.h2h-vs { color: var(--text-muted); font-weight: 600; }

.h2h-center {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 32px 0;
}
.h2h-muted { color: var(--text-muted); }

.h2h-players {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.h2h-player {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  flex: 1;
}
.h2h-name { font-weight: 600; }
.h2h-score {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 32px;
  font-weight: 700;
}
.h2h-wins--p1 { color: #3b82f6; }
.h2h-wins--p2 { color: #ef4444; }
.h2h-dash { color: var(--text-muted); }

.h2h-stats-row {
  display: flex;
  justify-content: space-around;
  margin: 16px 0;
  text-align: center;
}
.h2h-stat { display: flex; flex-direction: column; }
.h2h-stat-val { font-size: 18px; font-weight: 700; }
.h2h-stat-label { font-size: 12px; color: var(--text-muted); }

.h2h-chart {
  position: relative;
  height: 200px;
  margin: 8px 0 16px;
}

.h2h-streak {
  text-align: center;
  font-weight: 600;
  margin-bottom: 12px;
}

.h2h-form {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
  margin-bottom: 16px;
}
.h2h-form-label { font-size: 13px; color: var(--text-muted); margin-right: 4px; }
.h2h-form-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 700;
  color: #fff;
}
.h2h-form-badge--w { background: #10b981; }
.h2h-form-badge--l { background: #ef4444; }

.h2h-recent-label {
  display: block;
  font-size: 13px;
  color: var(--text-muted);
  margin-bottom: 8px;
}
.h2h-match {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 0;
  border-top: 1px solid var(--border-color, #eee);
  font-size: 13px;
}
.h2h-match-type {
  flex-shrink: 0;
  padding: 2px 6px;
  border-radius: 4px;
  background: var(--fill-color, #f1f1f1);
  font-size: 11px;
  font-weight: 600;
}
.h2h-lineup {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 6px;
  justify-content: center;
}
.h2h-team { color: var(--text-muted); }
.h2h-team--won { color: var(--text-primary, inherit); font-weight: 700; }
.h2h-lineup-vs { color: var(--text-muted); font-size: 11px; }
.h2h-match-date { flex-shrink: 0; color: var(--text-muted); font-size: 12px; }
</style>
