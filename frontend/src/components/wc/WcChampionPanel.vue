<template>
  <div class="champion-panel">
    <!-- Status banner -->
    <div v-if="config" class="champion-status-banner" :class="statusClass">
      <span>{{ statusIcon }}</span>
      <span class="champion-status-text">{{ statusText }}</span>
      <span v-if="config.winner_team" class="champion-winner-flag">
        {{ config.winner_team.flag_emoji }} {{ config.winner_team.name }}
      </span>
    </div>

    <!-- Main 2-col layout -->
    <div class="champion-main">
      <!-- LEFT: my prediction + place form -->
      <div class="champion-left">

        <!-- My prediction cards -->
        <template v-if="myPredictions.length > 0">
          <el-card
            v-for="pred in myPredictions"
            :key="pred.id"
            class="champion-card"
            shadow="never"
          >
            <div class="my-pred-header">
              <span class="my-pred-label">Dự đoán của bạn</span>
              <el-button
                v-if="config?.is_open && !config.settled_at"
                size="small" type="danger" text
                @click="handleDelete(pred.id)"
              >Xóa</el-button>
            </div>
            <div class="my-pred-body">
              <span class="pred-flag">{{ pred.flag_emoji }}</span>
              <div>
                <span class="pred-team">{{ pred.team_name }}</span>
                <div class="pred-meta">
                  <el-tag size="small" type="warning">{{ pred.odds_snapshot }}x</el-tag>
                  <span class="pred-points">{{ pred.points }} điểm</span>
                  <span class="pred-payout">→ <strong>{{ pred.payout_if_correct }}</strong> nếu đúng</span>
                </div>
              </div>
            </div>
            <div v-if="pred.result" class="my-pred-result">
              <el-tag :type="pred.result === 'correct' ? 'success' : 'danger'" size="small">
                {{ pred.result === 'correct' ? `+${pred.points_earned} điểm` : 'Không đúng' }}
              </el-tag>
            </div>
          </el-card>
        </template>

        <!-- Place form (window open) -->
        <el-card v-if="config?.is_open && !config.settled_at" class="champion-card" shadow="never">
          <div class="form-title">🏆 Đặt dự đoán Vô địch</div>
          <el-input
            v-model="teamSearch"
            placeholder="🔍 Tìm đội..."
            size="small"
            clearable
            style="margin-bottom: 10px"
          />
          <div class="teams-pick-grid">
            <div
              v-for="team in filteredTeams"
              :key="team.id"
              class="team-pick-card"
              :class="{ 'team-pick-card--selected': selectedTeamId === team.id }"
              @click="selectedTeamId = team.id"
            >
              <span class="pick-flag">{{ team.flag_emoji }}</span>
              <span class="pick-name">{{ team.name }}</span>
              <el-tag size="small" type="warning" effect="plain" class="pick-odds">{{ team.odds }}x</el-tag>
            </div>
          </div>
          <div v-if="selectedTeam" class="form-footer">
            <div class="points-row">
              <span class="points-label">Điểm cược (1–5)</span>
              <el-input-number v-model="selectedPoints" :min="1" :max="5" size="small" style="width: 80px" />
            </div>
            <div class="payout-preview">
              Nếu đúng →
              <strong class="payout-value">{{ payoutPreview }} điểm</strong>
              <span class="payout-formula">({{ selectedPoints }} × {{ selectedTeam.odds }}x)</span>
            </div>
            <el-button type="primary" :loading="placing" @click="handlePlace" style="margin-top: 10px; width: 100%">
              Đặt dự đoán {{ selectedTeam.flag_emoji }} {{ selectedTeam.name }}
            </el-button>
          </div>
          <div v-else-if="!filteredTeams.length" class="empty-teams">
            Không tìm thấy đội nào.
          </div>
        </el-card>

        <!-- Window closed notice (no predictions) -->
        <el-card v-if="config && !config.is_open && !config.settled_at && myPredictions.length === 0" class="champion-card" shadow="never">
          <div class="closed-notice">🔴 Cửa sổ dự đoán đã đóng — bạn chưa đặt dự đoán.</div>
        </el-card>

        <!-- All predictions table (below left on mobile, below form on desktop) -->
        <el-card v-if="allPredictions.length > 0" class="champion-card" shadow="never">
          <template #header>
            <span class="card-header-title">Tất cả dự đoán ({{ allPredictions.length }})</span>
          </template>
          <div style="overflow-x: auto">
          <el-table :data="allPredictions" size="small" max-height="300">
            <el-table-column label="Người dùng" prop="user_name" min-width="100" />
            <el-table-column label="Đội" min-width="110">
              <template #default="{ row }">{{ row.flag_emoji }} {{ row.team_name }}</template>
            </el-table-column>
            <el-table-column label="Điểm" prop="points" width="55" align="right" />
            <el-table-column label="Nếu đúng" width="80" align="right">
              <template #default="{ row }">{{ row.payout_if_correct }}</template>
            </el-table-column>
            <el-table-column label="KQ" width="68" align="center">
              <template #default="{ row }">
                <el-tag v-if="row.result" :type="row.result === 'correct' ? 'success' : 'danger'" size="small">
                  {{ row.result === 'correct' ? 'Đúng' : 'Sai' }}
                </el-tag>
                <span v-else style="color: #c0c4cc">–</span>
              </template>
            </el-table-column>
          </el-table>
          </div>
        </el-card>
      </div>

      <!-- RIGHT: teams odds grid -->
      <div class="champion-right">
        <el-card class="champion-card" shadow="never">
          <template #header>
            <div class="card-header-row">
              <span class="card-header-title">Tỉ lệ cược ({{ teams.length }} đội)</span>
              <el-input
                v-model="oddsSearch"
                placeholder="Tìm..."
                size="small"
                clearable
                style="width: 120px"
              />
            </div>
          </template>
          <div v-if="!teams.length" class="empty-teams">Đang tải...</div>
          <div v-else class="odds-grid">
            <div
              v-for="team in filteredOddsTeams"
              :key="team.id"
              class="odds-chip"
              :class="{ 'odds-chip--mine': myPredictions.some(p => p.team_name === team.name) }"
            >
              <span class="odds-chip-flag">{{ team.flag_emoji }}</span>
              <span class="odds-chip-name">{{ team.name }}</span>
              <el-tag size="small" type="warning" effect="plain">{{ team.odds }}x</el-tag>
            </div>
          </div>
        </el-card>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { wcService } from '@/services/wcService'
import type { WcChampionConfig, WcChampionTeam, WcChampionPredictionMine, WcChampionPredictionPublic } from '@/types/wc'

const config = ref<WcChampionConfig | null>(null)
const teams = ref<WcChampionTeam[]>([])
const myPredictions = ref<WcChampionPredictionMine[]>([])
const allPredictions = ref<WcChampionPredictionPublic[]>([])

const selectedTeamId = ref('')
const selectedPoints = ref(3)
const placing = ref(false)
const teamSearch = ref('')
const oddsSearch = ref('')

const selectedTeam = computed(() => teams.value.find(t => t.id === selectedTeamId.value) ?? null)
const payoutPreview = computed(() =>
  selectedTeam.value ? Math.floor(selectedPoints.value * selectedTeam.value.odds) : 0
)
const pickedTeamIds = computed(() => new Set(myPredictions.value.map(p => p.team_id)))

const filteredTeams = computed(() => {
  const q = teamSearch.value.toLowerCase()
  const base = teams.value.filter(t => !pickedTeamIds.value.has(t.id))
  return q ? base.filter(t => t.name.toLowerCase().includes(q) || t.code.toLowerCase().includes(q)) : base
})
const filteredOddsTeams = computed(() => {
  const q = oddsSearch.value.toLowerCase()
  return q ? teams.value.filter(t => t.name.toLowerCase().includes(q) || t.code.toLowerCase().includes(q)) : teams.value
})

const statusClass = computed(() => {
  if (!config.value) return ''
  if (config.value.settled_at) return 'status-settled'
  if (config.value.is_open) return 'status-open'
  return 'status-closed'
})
const statusIcon = computed(() => {
  if (!config.value) return ''
  if (config.value.settled_at) return '🏆'
  if (config.value.is_open) return '🟢'
  return '🔴'
})
const statusText = computed(() => {
  if (!config.value) return ''
  if (config.value.settled_at) return 'Đã có kết quả vô địch'
  if (config.value.is_open) return 'Đang nhận dự đoán vô địch'
  return 'Cửa sổ dự đoán đã đóng'
})

async function load() {
  const [cfg, teamList, preds, mine] = await Promise.allSettled([
    wcService.getChampionConfig(),
    wcService.getChampionTeams(),
    wcService.getChampionPredictions(),
    wcService.getMyChampionPredictions(),
  ])
  if (cfg.status === 'fulfilled') config.value = cfg.value
  if (teamList.status === 'fulfilled') teams.value = teamList.value ?? []
  if (preds.status === 'fulfilled') allPredictions.value = preds.value ?? []
  if (mine.status === 'fulfilled') myPredictions.value = mine.value ?? []
}

async function handlePlace() {
  if (!selectedTeamId.value) { ElMessage.warning('Vui lòng chọn đội'); return }
  placing.value = true
  try {
    const newPred = await wcService.placeChampionPrediction(selectedTeamId.value, selectedPoints.value)
    myPredictions.value = [...myPredictions.value, newPred]
    allPredictions.value = await wcService.getChampionPredictions()
    selectedTeamId.value = ''
    ElMessage.success('Đã đặt dự đoán!')
  } catch {
    // errors handled by wcApi interceptor
  } finally {
    placing.value = false
  }
}

async function handleDelete(predId: string) {
  await ElMessageBox.confirm('Xóa dự đoán vô địch này?', 'Xác nhận', { type: 'warning' })
  try {
    await wcService.deleteChampionPrediction(predId)
    myPredictions.value = myPredictions.value.filter(p => p.id !== predId)
    allPredictions.value = await wcService.getChampionPredictions()
    ElMessage.success('Đã xóa dự đoán')
  } catch {
    // errors handled by wcApi interceptor
  }
}

onMounted(load)
</script>

<style scoped>
.champion-panel {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 4px 0;
}

/* Status banner */
.champion-status-banner {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  border-radius: 8px;
  font-weight: 600;
  font-size: 14px;
}
.status-open    { background: #f0f9eb; color: #67c23a; }
.status-closed  { background: #fef0f0; color: #f56c6c; }
.status-settled { background: #fdf6ec; color: #e6a23c; }
.champion-winner-flag { margin-left: auto; font-size: 18px; }

/* 2-col layout */
.champion-main {
  display: grid;
  grid-template-columns: 380px 1fr;
  gap: 14px;
  align-items: start;
}
@media (max-width: 700px) {
  .champion-main { grid-template-columns: 1fr; }
}
@media (max-width: 420px) {
  .teams-pick-grid { grid-template-columns: 1fr; }
}
.champion-left, .champion-right { display: flex; flex-direction: column; gap: 14px; }

.champion-card { border-radius: 10px; }

/* My prediction */
.my-pred-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
.my-pred-label { font-weight: 600; font-size: 13px; color: #606266; }
.my-pred-body { display: flex; align-items: center; gap: 14px; }
.pred-flag { font-size: 32px; flex-shrink: 0; }
.pred-team { font-weight: 700; font-size: 17px; display: block; margin-bottom: 6px; }
.pred-meta { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.pred-points { font-size: 13px; color: #409eff; font-weight: 600; }
.pred-payout { font-size: 13px; color: #606266; }
.my-pred-result { margin-top: 12px; }

/* Place form */
.form-title { font-size: 15px; font-weight: 700; margin-bottom: 12px; }

.teams-pick-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px;
  max-height: 300px;
  overflow-y: auto;
  margin-bottom: 12px;
}
.team-pick-card {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 7px 10px;
  border: 1.5px solid var(--border-default, #dcdfe6);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.12s;
  user-select: none;
  min-width: 0;
}
.team-pick-card:hover { border-color: #409eff; background: rgba(64,158,255,0.04); }
.team-pick-card--selected {
  border-color: #409eff;
  background: rgba(64,158,255,0.10);
  box-shadow: 0 0 0 1px #409eff;
}
.pick-flag { font-size: 18px; flex-shrink: 0; }
.pick-name { flex: 1; font-size: 12px; font-weight: 600; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.pick-odds { flex-shrink: 0; }

.form-footer { border-top: 1px solid var(--border-subtle, #f0f0f0); padding-top: 12px; }
.points-row { display: flex; align-items: center; gap: 10px; margin-bottom: 10px; }
.points-label { font-size: 13px; color: #606266; white-space: nowrap; }
.payout-preview { font-size: 14px; color: #303133; }
.payout-value { color: #67c23a; font-size: 16px; }
.payout-formula { color: #909399; margin-left: 6px; font-size: 12px; }

.closed-notice { font-size: 14px; color: #909399; text-align: center; padding: 16px 0; }
.empty-teams { font-size: 13px; color: #c0c4cc; text-align: center; padding: 16px 0; }

/* Right column */
.card-header-title { font-weight: 600; font-size: 14px; }
.card-header-row { display: flex; justify-content: space-between; align-items: center; gap: 10px; }

/* Odds grid */
.odds-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(155px, 1fr));
  gap: 6px;
  max-height: 480px;
  overflow-y: auto;
}
.odds-chip {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 8px;
  border-radius: 7px;
  background: var(--surface-card, #fafafa);
  border: 1px solid var(--border-subtle, #f0f0f0);
  font-size: 12px;
  min-width: 0;
}
.odds-chip--mine {
  border-color: #409eff;
  background: rgba(64,158,255,0.08);
}
.odds-chip-flag { font-size: 16px; flex-shrink: 0; }
.odds-chip-name { flex: 1; font-weight: 600; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
</style>
