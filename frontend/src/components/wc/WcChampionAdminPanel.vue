<template>
  <div class="card card-body wc-admin-section">
    <div class="wc-admin-section-title">🏆 Champion Prediction</div>

    <!-- Config toggle -->
    <div class="champion-config-row">
      <span class="champion-config-label">Cửa sổ dự đoán:</span>
      <el-switch
        v-model="isOpen"
        :disabled="!!config?.settled_at || configLoading"
        active-text="Đang mở"
        inactive-text="Đóng"
        @change="handleConfigChange"
      />
      <el-tag v-if="config?.settled_at" type="success" size="small" style="margin-left: 8px">
        Đã settle
      </el-tag>
    </div>

    <!-- Teams odds table with inline editing -->
    <div class="teams-toolbar">
      <el-input
        v-model="teamSearch"
        placeholder="Tìm đội..."
        size="small"
        clearable
        style="width: 200px"
      />
      <span class="teams-count">{{ filteredTeams.length }} / {{ teams.length }} đội</span>
    </div>
    <el-table :data="filteredTeams" size="small" style="margin-top: 8px" max-height="420">
      <el-table-column label="Đội" min-width="160">
        <template #default="{ row }">{{ row.flag_emoji }} {{ row.name }}</template>
      </el-table-column>
      <el-table-column label="Code" prop="code" width="58" />
      <el-table-column label="Odds hiện tại" width="90" align="right">
        <template #default="{ row }">
          <el-tag size="small" type="warning">{{ row.odds }}x</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="Sửa Odds" width="175" align="right">
        <template #default="{ row }">
          <div class="odds-edit-row">
            <el-input-number
              v-model="oddsEdit[row.id]"
              :min="1.01"
              :step="0.5"
              :precision="2"
              size="small"
              style="width: 95px"
            />
            <el-button
              size="small"
              type="primary"
              text
              :loading="savingOdds[row.id]"
              @click="handleSaveOdds(row)"
            >
              Lưu
            </el-button>
          </div>
        </template>
      </el-table-column>
    </el-table>

    <!-- Settle section -->
    <div v-if="!config?.settled_at" class="settle-section">
      <el-divider />
      <h4 class="settle-title">Công bố Vô địch</h4>
      <p class="settle-hint">Chọn đội vô địch và xác nhận. Thao tác này không thể hoàn tác.</p>
      <div class="settle-row">
        <el-select v-model="winnerTeamId" placeholder="Chọn đội vô địch" style="flex: 1" filterable>
          <el-option
            v-for="team in teams"
            :key="team.id"
            :label="`${team.flag_emoji} ${team.name}`"
            :value="team.id"
          />
        </el-select>
        <el-button
          plain
          type="danger"
          :loading="settling"
          :disabled="!winnerTeamId || !!config?.is_open"
          @click="handleSettle"
        >
          Công bố &amp; Settle
        </el-button>
      </div>
      <p v-if="config?.is_open" class="settle-warning">
        Đóng cửa sổ dự đoán trước khi settle
      </p>
    </div>

    <!-- Settle result -->
    <el-alert
      v-if="settleResult"
      type="success"
      :closable="false"
      style="margin-top: 12px"
    >
      <div>
        🏆 <strong>{{ settleResult.winner }}</strong> vô địch!
        Đã settle {{ settleResult.settled_count }} dự đoán,
        {{ settleResult.correct_count }} đúng,
        tổng {{ settleResult.total_points_awarded }} điểm được trao.
      </div>
    </el-alert>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { wcService } from '@/services/wcService'
import type { WcChampionConfig, WcChampionTeam, WcChampionSettleResult } from '@/types/wc'

const config = ref<WcChampionConfig | null>(null)
const teams = ref<WcChampionTeam[]>([])
const isOpen = ref(false)
const configLoading = ref(false)
const oddsEdit = reactive<Record<string, number>>({})
const savingOdds = reactive<Record<string, boolean>>({})
const winnerTeamId = ref('')
const settling = ref(false)
const settleResult = ref<WcChampionSettleResult | null>(null)
const teamSearch = ref('')

const filteredTeams = computed(() => {
  const q = teamSearch.value.toLowerCase()
  return q ? teams.value.filter(t => t.name.toLowerCase().includes(q) || t.code.toLowerCase().includes(q)) : teams.value
})

async function load() {
  const [cfg, teamList] = await Promise.allSettled([
    wcService.getChampionConfig(),
    wcService.getChampionTeams(),
  ])
  if (cfg.status === 'fulfilled') {
    config.value = cfg.value
    isOpen.value = cfg.value.is_open
  }
  if (teamList.status === 'fulfilled') {
    teams.value = teamList.value
    for (const t of teamList.value) {
      oddsEdit[t.id] = t.odds
    }
  }
}

async function handleConfigChange(val: boolean) {
  configLoading.value = true
  try {
    await wcService.adminUpdateChampionConfig(val)
    if (config.value) config.value.is_open = val
    ElMessage.success(val ? 'Đã mở cửa sổ dự đoán' : 'Đã đóng cửa sổ dự đoán')
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error ?? 'Lỗi khi cập nhật config')
    isOpen.value = !val
  } finally {
    configLoading.value = false
  }
}

async function handleSaveOdds(team: WcChampionTeam) {
  savingOdds[team.id] = true
  try {
    await wcService.adminUpdateChampionTeamOdds(team.id, oddsEdit[team.id])
    team.odds = oddsEdit[team.id]
    ElMessage.success(`Đã cập nhật odds ${team.name}`)
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error ?? 'Lỗi khi lưu odds')
  } finally {
    savingOdds[team.id] = false
  }
}

async function handleSettle() {
  if (!winnerTeamId.value) return
  const team = teams.value.find(t => t.id === winnerTeamId.value)
  await ElMessageBox.confirm(
    `Công bố ${team?.flag_emoji} ${team?.name} là vô địch WC2026 và settle tất cả dự đoán?`,
    'Xác nhận Settle',
    { type: 'warning', confirmButtonText: 'Xác nhận', cancelButtonText: 'Hủy' }
  )
  settling.value = true
  try {
    settleResult.value = await wcService.adminSettleChampion(winnerTeamId.value)
    config.value = await wcService.getChampionConfig()
    ElMessage.success('Settle thành công!')
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error ?? 'Lỗi khi settle')
  } finally {
    settling.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.champion-config-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 8px;
}
.champion-config-label {
  font-size: 14px;
  color: #606266;
}
.teams-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 14px;
}
.teams-count {
  font-size: 12px;
  color: #909399;
}
.odds-edit-row {
  display: flex;
  align-items: center;
  gap: 6px;
}
.settle-section { margin-top: 4px; }
.settle-title { font-size: 14px; font-weight: 700; margin: 0 0 6px; }
.settle-hint { font-size: 13px; color: #909399; margin: 0 0 10px; }
.settle-row { display: flex; gap: 10px; align-items: center; }
.settle-warning { font-size: 12px; color: #f56c6c; margin: 6px 0 0; }
</style>
