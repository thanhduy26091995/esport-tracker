<template>
  <div class="champion-panel">
    <!-- Status banner -->
    <div v-if="config" class="champion-status-banner" :class="statusClass">
      <span class="champion-status-icon">{{ statusIcon }}</span>
      <span class="champion-status-text">{{ statusText }}</span>
      <span v-if="config.winner_team" class="champion-winner-flag">
        {{ config.winner_team.flag_emoji }} {{ config.winner_team.name }}
      </span>
    </div>

    <!-- My prediction card (when already placed) -->
    <el-card v-if="myPrediction" class="my-prediction-card" shadow="never">
      <div class="my-prediction-header">
        <span class="my-prediction-label">Dự đoán của bạn</span>
        <div class="my-prediction-actions" v-if="config?.is_open && !config.settled_at">
          <el-button size="small" type="danger" text @click="handleDelete">Xóa</el-button>
        </div>
      </div>
      <div class="my-prediction-info">
        <span class="champion-flag">{{ myPrediction.flag_emoji }}</span>
        <span class="champion-team-name">{{ myPrediction.team_name }}</span>
        <el-tag size="small" type="info">{{ myPrediction.odds_snapshot }}x</el-tag>
        <span class="champion-points">{{ myPrediction.points }} điểm</span>
        <span class="champion-payout">→ {{ myPrediction.payout_if_correct }} điểm nếu đúng</span>
      </div>
      <div v-if="myPrediction.result" class="my-prediction-result">
        <el-tag :type="myPrediction.result === 'correct' ? 'success' : 'danger'">
          {{ myPrediction.result === 'correct' ? `+${myPrediction.points_earned} điểm` : 'Không đúng' }}
        </el-tag>
      </div>
    </el-card>

    <!-- Place prediction form (when window is open and no prediction yet) -->
    <el-card v-if="config?.is_open && !config.settled_at && !myPrediction" class="place-form-card" shadow="never">
      <h3 class="form-title">Dự đoán Vô địch</h3>
      <div class="form-row">
        <el-select v-model="selectedTeamId" placeholder="Chọn đội vô địch" class="team-select" filterable>
          <el-option
            v-for="team in teams"
            :key="team.id"
            :label="`${team.flag_emoji} ${team.name} (${team.odds}x)`"
            :value="team.id"
          />
        </el-select>
        <div class="points-input-wrap">
          <span class="points-label">Điểm cược (1–5)</span>
          <el-input-number
            v-model="selectedPoints"
            :min="1"
            :max="5"
            size="small"
            style="width: 80px"
          />
        </div>
      </div>
      <div v-if="selectedTeam" class="payout-preview">
        <span>Nếu đúng → </span>
        <strong>{{ payoutPreview }} điểm</strong>
        <span class="payout-formula">({{ selectedPoints }} × {{ selectedTeam.odds }}x)</span>
      </div>
      <el-button type="primary" :loading="placing" @click="handlePlace" style="margin-top: 12px">
        Đặt dự đoán
      </el-button>
    </el-card>

    <!-- Teams odds table -->
    <el-card class="teams-card" shadow="never">
      <template #header>
        <span>Tỉ lệ cược vô địch</span>
      </template>
      <el-table :data="teams" size="small" :show-header="true">
        <el-table-column label="Đội" min-width="140">
          <template #default="{ row }">
            <span>{{ row.flag_emoji }} {{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column label="Code" prop="code" width="60" />
        <el-table-column label="Tỉ lệ" width="80" align="right">
          <template #default="{ row }">
            <el-tag size="small" type="warning">{{ row.odds }}x</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- All predictions leaderboard -->
    <el-card v-if="allPredictions.length > 0" class="predictions-card" shadow="never">
      <template #header>
        <span>Tất cả dự đoán ({{ allPredictions.length }})</span>
      </template>
      <el-table :data="allPredictions" size="small">
        <el-table-column label="Người dùng" prop="user_name" min-width="110" />
        <el-table-column label="Đội" min-width="120">
          <template #default="{ row }">
            {{ row.flag_emoji }} {{ row.team_name }}
          </template>
        </el-table-column>
        <el-table-column label="Điểm" prop="points" width="60" align="right" />
        <el-table-column label="Nếu đúng" width="90" align="right">
          <template #default="{ row }">{{ row.payout_if_correct }}</template>
        </el-table-column>
        <el-table-column label="KQ" width="80" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.result" :type="row.result === 'correct' ? 'success' : 'danger'" size="small">
              {{ row.result === 'correct' ? 'Đúng' : 'Sai' }}
            </el-tag>
            <span v-else class="result-pending">–</span>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { wcService } from '@/services/wcService'
import type { WcChampionConfig, WcChampionTeam, WcChampionPredictionMine, WcChampionPredictionPublic } from '@/types/wc'

const config = ref<WcChampionConfig | null>(null)
const teams = ref<WcChampionTeam[]>([])
const myPrediction = ref<WcChampionPredictionMine | null>(null)
const allPredictions = ref<WcChampionPredictionPublic[]>([])

const selectedTeamId = ref<string>('')
const selectedPoints = ref(3)
const placing = ref(false)

const selectedTeam = computed(() => teams.value.find(t => t.id === selectedTeamId.value) ?? null)
const payoutPreview = computed(() =>
  selectedTeam.value ? Math.floor(selectedPoints.value * selectedTeam.value.odds) : 0
)

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
  if (config.value.settled_at) return 'Đã có kết quả'
  if (config.value.is_open) return 'Đang nhận dự đoán'
  return 'Cửa sổ đã đóng'
})

async function load() {
  const [cfg, teamList, preds, mine] = await Promise.allSettled([
    wcService.getChampionConfig(),
    wcService.getChampionTeams(),
    wcService.getChampionPredictions(),
    wcService.getMyChampionPrediction(),
  ])
  if (cfg.status === 'fulfilled') config.value = cfg.value
  if (teamList.status === 'fulfilled') teams.value = teamList.value
  if (preds.status === 'fulfilled') allPredictions.value = preds.value
  if (mine.status === 'fulfilled') myPrediction.value = mine.value
}

async function handlePlace() {
  if (!selectedTeamId.value) {
    ElMessage.warning('Vui lòng chọn đội')
    return
  }
  placing.value = true
  try {
    myPrediction.value = await wcService.placeChampionPrediction(selectedTeamId.value, selectedPoints.value)
    allPredictions.value = await wcService.getChampionPredictions()
    ElMessage.success('Đã đặt dự đoán!')
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error ?? 'Lỗi khi đặt dự đoán')
  } finally {
    placing.value = false
  }
}

async function handleDelete() {
  await ElMessageBox.confirm('Xóa dự đoán vô địch của bạn?', 'Xác nhận', { type: 'warning' })
  try {
    await wcService.deleteChampionPrediction()
    myPrediction.value = null
    allPredictions.value = await wcService.getChampionPredictions()
    ElMessage.success('Đã xóa dự đoán')
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error ?? 'Lỗi khi xóa')
  }
}

onMounted(load)
</script>

<style scoped>
.champion-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 8px 0;
}

.champion-status-banner {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  border-radius: 8px;
  font-weight: 600;
  font-size: 14px;
}

.status-open { background: #f0f9eb; color: #67c23a; }
.status-closed { background: #fef0f0; color: #f56c6c; }
.status-settled { background: #fdf6ec; color: #e6a23c; }

.champion-status-icon { font-size: 16px; }
.champion-winner-flag { margin-left: auto; font-size: 18px; }

.my-prediction-card, .place-form-card, .teams-card, .predictions-card {
  border-radius: 10px;
}

.my-prediction-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}
.my-prediction-label { font-weight: 600; font-size: 13px; color: #606266; }
.my-prediction-info {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}
.champion-flag { font-size: 22px; }
.champion-team-name { font-weight: 700; font-size: 16px; }
.champion-points { font-size: 14px; color: #409eff; font-weight: 600; }
.champion-payout { font-size: 13px; color: #909399; }
.my-prediction-result { margin-top: 10px; }

.form-title { margin: 0 0 14px; font-size: 15px; font-weight: 700; }
.form-row {
  display: flex;
  align-items: center;
  gap: 14px;
  flex-wrap: wrap;
}
.team-select { flex: 1; min-width: 200px; }
.points-input-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
}
.points-label { font-size: 13px; color: #606266; white-space: nowrap; }

.payout-preview {
  margin-top: 10px;
  font-size: 14px;
  color: #303133;
}
.payout-formula { color: #909399; margin-left: 6px; font-size: 12px; }

.result-pending { color: #c0c4cc; }
</style>
