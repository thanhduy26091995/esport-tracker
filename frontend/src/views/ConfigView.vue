<template>
  <div class="page-wrapper">
    <div class="page-container">

      <!-- Header -->
      <div class="page-header">
        <div class="page-header-left">
          <h1 class="page-title">Settings</h1>
          <p class="page-subtitle">Manage system settings and game rules</p>
        </div>
      </div>

      <div v-if="configStore.loading" class="card">
        <div class="card-body flex justify-center items-center py-16">
          <el-icon class="animate-spin" :size="28" style="color:var(--text-muted)"><Loading /></el-icon>
        </div>
      </div>

      <form v-else @submit.prevent="handleSubmit" class="cfg-layout">

        <!-- Left column: main settings -->
        <div class="cfg-main">

          <!-- Settlement Mode (prominent) -->
          <div class="card">
            <div class="card-header">
              <span class="card-title">Settlement Mode</span>
              <span class="mode-badge" :class="formData.auto_settlement ? 'mode-badge--auto' : 'mode-badge--manual'">
                {{ formData.auto_settlement ? 'AUTO' : 'MANUAL' }}
              </span>
            </div>
            <div class="card-body">
              <div class="mode-toggle-row">
                <div class="mode-icon-wrap" :class="formData.auto_settlement ? 'mode-icon--auto' : 'mode-icon--manual'">
                  <el-icon :size="20"><component :is="formData.auto_settlement ? Lightning : User" /></el-icon>
                </div>
                <div class="mode-text">
                  <p class="mode-label">{{ formData.auto_settlement ? 'Automatic Settlement' : 'Manual Settlement' }}</p>
                  <p class="mode-desc">
                    {{ formData.auto_settlement
                      ? 'Settlement triggers automatically when a player reaches the debt threshold after a match.'
                      : 'Settlement must be triggered manually from the Players page. Recommended for team-controlled timing.' }}
                  </p>
                </div>
                <el-switch
                  v-model="formData.auto_settlement"
                  size="large"
                  :active-color="'var(--color-success)'"
                  :inactive-color="'var(--border-default)'"
                />
              </div>
            </div>
          </div>

          <!-- Debt Threshold -->
          <div class="card">
            <div class="card-header">
              <span class="card-title">Debt Threshold</span>
              <span class="cfg-badge cfg-badge--warning">Points</span>
            </div>
            <div class="card-body">
              <el-input
                v-model.number="formData.debt_threshold"
                type="number"
                placeholder="e.g. -6"
                :max="0"
                size="large"
                class="w-full"
              >
                <template #prefix>
                  <el-icon><Warning /></el-icon>
                </template>
                <template #suffix>pts</template>
              </el-input>
              <p class="cfg-hint">
                When a player's score reaches this value (must be ≤ 0), settlement can be triggered.
              </p>
              <p v-if="formData.debt_threshold > 0" class="cfg-error">
                Value must be less than or equal to 0
              </p>
              <div v-if="formData.debt_threshold <= 0" class="cfg-preview">
                <span class="cfg-preview-label">Trigger at</span>
                <span class="cfg-preview-val cfg-preview-val--red">{{ formData.debt_threshold }} points</span>
              </div>
            </div>
          </div>

          <!-- Point to VND -->
          <div class="card">
            <div class="card-header">
              <span class="card-title">Point Value</span>
              <span class="cfg-badge cfg-badge--info">Conversion</span>
            </div>
            <div class="card-body">
              <el-input
                v-model.number="formData.point_to_vnd"
                type="number"
                placeholder="e.g. 22000"
                :min="1"
                size="large"
                class="w-full"
              >
                <template #prefix>
                  <el-icon><Money /></el-icon>
                </template>
                <template #suffix>VND / pt</template>
              </el-input>
              <p class="cfg-hint">
                How much 1 point is worth in Vietnamese Dong. Used for all debt calculations.
              </p>
              <p v-if="formData.point_to_vnd <= 0" class="cfg-error">
                Value must be greater than 0
              </p>
              <div v-if="formData.point_to_vnd > 0" class="cfg-preview">
                <span class="cfg-preview-label">Example: -10 pts =</span>
                <span class="cfg-preview-val cfg-preview-val--red">{{ formatVND(formData.point_to_vnd * 10) }} debt</span>
              </div>
            </div>
          </div>

          <!-- Points Per Win -->
          <div class="card">
            <div class="card-header">
              <span class="card-title">Points Per Win</span>
              <span class="cfg-badge cfg-badge--info">Per Match</span>
            </div>
            <div class="card-body">
              <el-input
                v-model.number="formData.points_per_win"
                type="number"
                placeholder="e.g. 1"
                :min="1"
                size="large"
                class="w-full"
              >
                <template #prefix>
                  <el-icon><Trophy /></el-icon>
                </template>
                <template #suffix>pts</template>
              </el-input>
              <p class="cfg-hint">
                Points awarded to the winning team (and deducted from the losers) per match. Default is 1.
              </p>
              <p v-if="formData.points_per_win <= 0" class="cfg-error">
                Value must be greater than 0
              </p>
              <div v-if="formData.points_per_win > 0" class="cfg-preview">
                <span class="cfg-preview-label">Winner gets</span>
                <span class="cfg-preview-val" style="color:var(--color-success)">+{{ formData.points_per_win }} pts</span>
                <span class="cfg-preview-label ml-2">Loser gets</span>
                <span class="cfg-preview-val cfg-preview-val--red">-{{ formData.points_per_win }} pts</span>
              </div>
            </div>
          </div>

          <!-- Fund Split -->
          <div class="card">
            <div class="card-header">
              <span class="card-title">Fund Split</span>
              <span class="cfg-badge cfg-badge--success">{{ formData.fund_split_percent }}% → Fund</span>
            </div>
            <div class="card-body">
              <div class="split-row">
                <el-slider
                  v-model="formData.fund_split_percent"
                  :min="0" :max="100" :step="1"
                  :marks="marks"
                  class="flex-1"
                />
                <el-input
                  v-model.number="formData.fund_split_percent"
                  type="number" :min="0" :max="100"
                  size="large" class="w-24"
                >
                  <template #suffix>%</template>
                </el-input>
              </div>
              <p class="cfg-hint">
                Percentage of each settlement going to the fund. The rest is split among winners.
              </p>
              <div v-if="formData.fund_split_percent >= 0 && formData.fund_split_percent <= 100"
                   class="split-preview">
                <div class="split-bar">
                  <div class="split-bar-fund" :style="{ width: formData.fund_split_percent + '%' }">
                    <span v-if="formData.fund_split_percent >= 15">{{ formData.fund_split_percent }}% Fund</span>
                  </div>
                  <div class="split-bar-winners" :style="{ width: (100 - formData.fund_split_percent) + '%' }">
                    <span v-if="(100 - formData.fund_split_percent) >= 15">{{ 100 - formData.fund_split_percent }}% Winners</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Actions -->
          <div class="cfg-actions">
            <el-button @click="handleReset" :disabled="!hasChanges">
              Reset
            </el-button>
            <el-button
              type="primary"
              size="large"
              native-type="submit"
              :loading="configStore.loading"
              :disabled="!isFormValid || !hasChanges"
            >
              Save Changes
            </el-button>
          </div>
        </div>

        <!-- Right column: live summary -->
        <div class="cfg-sidebar">
          <div class="card cfg-summary-card">
            <div class="card-header">
              <span class="card-title">Current Config</span>
            </div>
            <div class="card-body cfg-summary-body">
              <div class="cfg-summary-item">
                <span class="cfg-summary-label">Settlement Mode</span>
                <span class="cfg-summary-val"
                  :class="configStore.autoSettlement ? 'cfg-summary-val--green' : 'cfg-summary-val--muted'">
                  {{ configStore.autoSettlement ? 'Auto' : 'Manual' }}
                </span>
              </div>
              <div class="cfg-summary-item">
                <span class="cfg-summary-label">Debt Threshold</span>
                <span class="cfg-summary-val cfg-summary-val--red">{{ configStore.debtThreshold }} pts</span>
              </div>
              <div class="cfg-summary-item">
                <span class="cfg-summary-label">Point Value</span>
                <span class="cfg-summary-val">{{ formatVND(configStore.pointToVnd) }}</span>
              </div>
              <div class="cfg-summary-item">
                <span class="cfg-summary-label">Fund Split</span>
                <span class="cfg-summary-val cfg-summary-val--green">{{ configStore.fundSplitPercent }}%</span>
              </div>
              <div class="cfg-summary-item">
                <span class="cfg-summary-label">Points Per Win</span>
                <span class="cfg-summary-val">{{ configStore.pointsPerWin }} pts</span>
              </div>
            </div>
          </div>

          <div class="card">
            <div class="card-header">
              <span class="card-title">Important Notice</span>
            </div>
            <div class="card-body">
              <p class="cfg-notice-text">
                Changes apply to all future matches and settlements. Existing records will not be recalculated.
              </p>
            </div>
          </div>
        </div>

      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, h } from 'vue'
import { ElMessage } from 'element-plus'
import { Loading, Warning, Money, User, Trophy } from '@element-plus/icons-vue'
import { useConfigStore } from '@/stores/configStore'
import { formatVND } from '@/utils/formatters'

const Lightning = {
  render: () => h('svg', { viewBox: '0 0 24 24', fill: 'currentColor', style: 'width:1em;height:1em' }, [
    h('path', { d: 'M13 2L4.5 13.5H11L10 22L20.5 10H14L13 2Z' })
  ])
}

const configStore = useConfigStore()

const formData = ref({
  debt_threshold: 0,
  point_to_vnd: 0,
  fund_split_percent: 0,
  auto_settlement: false,
  points_per_win: 1
})

const marks = { 0: '0', 25: '25', 50: '50', 75: '75', 100: '100' }

onMounted(async () => {
  await configStore.fetchConfigs()
  resetFormData()
})

watch(
  () => [configStore.debtThreshold, configStore.pointToVnd, configStore.fundSplitPercent, configStore.autoSettlement, configStore.pointsPerWin],
  () => resetFormData(),
  { deep: true }
)

const resetFormData = () => {
  formData.value = {
    debt_threshold: configStore.debtThreshold,
    point_to_vnd: configStore.pointToVnd,
    fund_split_percent: configStore.fundSplitPercent,
    auto_settlement: configStore.autoSettlement,
    points_per_win: configStore.pointsPerWin
  }
}

const isFormValid = computed(() =>
  formData.value.debt_threshold <= 0 &&
  formData.value.point_to_vnd > 0 &&
  formData.value.fund_split_percent >= 0 &&
  formData.value.fund_split_percent <= 100 &&
  formData.value.points_per_win > 0
)

const hasChanges = computed(() =>
  formData.value.debt_threshold !== configStore.debtThreshold ||
  formData.value.point_to_vnd !== configStore.pointToVnd ||
  formData.value.fund_split_percent !== configStore.fundSplitPercent ||
  formData.value.auto_settlement !== configStore.autoSettlement ||
  formData.value.points_per_win !== configStore.pointsPerWin
)

const handleReset = () => {
  resetFormData()
  ElMessage.info('Reset to saved values')
}

const handleSubmit = async () => {
  if (!isFormValid.value) { ElMessage.error('Please fix validation errors before saving'); return }
  try {
    await configStore.updateAllConfigs({
      debt_threshold: formData.value.debt_threshold.toString(),
      point_to_vnd: formData.value.point_to_vnd.toString(),
      fund_split_percent: formData.value.fund_split_percent.toString(),
      auto_settlement: formData.value.auto_settlement.toString(),
      points_per_win: formData.value.points_per_win.toString(),
    })
    ElMessage.success('Configuration saved')
  } catch {}
}
</script>

<style scoped>
/* Layout */
.cfg-layout {
  display: grid;
  grid-template-columns: 1fr;
  gap: 16px;
}
@media (min-width: 1024px) {
  .cfg-layout { grid-template-columns: 1fr 280px; align-items: start; }
}

.cfg-main { display: flex; flex-direction: column; gap: 16px; }

/* Settlement Mode toggle */
.mode-badge {
  font-size: 10px; font-weight: 800; letter-spacing: 0.08em;
  padding: 3px 8px; border-radius: 6px;
}
.mode-badge--manual { background: var(--border-subtle); color: var(--text-muted); }
.mode-badge--auto   { background: var(--color-success-bg); color: var(--color-success); border: 1px solid var(--color-success-border); }

.mode-toggle-row {
  display: flex; align-items: flex-start; gap: 14px;
}

.mode-icon-wrap {
  width: 40px; height: 40px; border-radius: 10px;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
  transition: background 0.2s;
}
.mode-icon--manual { background: var(--surface-page); color: var(--text-muted); }
.mode-icon--auto   { background: var(--color-success-bg); color: var(--color-success); }

.mode-text { flex: 1; min-width: 0; }
.mode-label { font-size: 14px; font-weight: 700; color: var(--text-primary); margin-bottom: 4px; }
.mode-desc  { font-size: 12px; color: var(--text-muted); line-height: 1.5; }

/* Badges */
.cfg-badge {
  font-size: 11px; font-weight: 600; padding: 3px 8px; border-radius: 6px;
}
.cfg-badge--warning { background: var(--color-warning-bg); color: var(--color-warning); }
.cfg-badge--info    { background: var(--color-info-bg); color: var(--color-info); }
.cfg-badge--success { background: var(--color-success-bg); color: var(--color-success); }

/* Field helpers */
.cfg-hint {
  font-size: 12px; color: var(--text-muted); margin-top: 8px; line-height: 1.5;
}
.cfg-error {
  font-size: 12px; color: var(--color-danger); margin-top: 6px; font-weight: 500;
}
.cfg-preview {
  display: flex; align-items: center; gap: 8px;
  margin-top: 10px; padding: 10px 12px;
  background: var(--surface-page); border-radius: 8px;
  font-size: 12px;
}
.cfg-preview-label { color: var(--text-muted); }
.cfg-preview-val { font-weight: 700; }
.cfg-preview-val--red { color: var(--color-danger); }

/* Fund split bar */
.split-row { display: flex; align-items: center; gap: 16px; margin-bottom: 4px; }

.split-preview { margin-top: 14px; }
.split-bar {
  display: flex; height: 28px; border-radius: 8px; overflow: hidden; gap: 2px;
}
.split-bar-fund {
  background: var(--color-success);
  display: flex; align-items: center; justify-content: center;
  font-size: 11px; font-weight: 700; color: white;
  transition: width 0.3s;
  border-radius: 8px 0 0 8px; min-width: 0;
}
.split-bar-winners {
  background: var(--color-info);
  display: flex; align-items: center; justify-content: center;
  font-size: 11px; font-weight: 700; color: white;
  transition: width 0.3s;
  border-radius: 0 8px 8px 0; min-width: 0;
}

/* Actions */
.cfg-actions {
  display: flex; justify-content: flex-end; gap: 10px; padding-top: 4px;
}

/* Sidebar summary */
.cfg-sidebar { display: flex; flex-direction: column; gap: 16px; }

.cfg-summary-body { display: flex; flex-direction: column; gap: 0; }

.cfg-summary-item {
  display: flex; align-items: center; justify-content: space-between;
  padding: 10px 0;
  border-bottom: 1px solid var(--border-subtle);
}
.cfg-summary-item:last-child { border-bottom: none; }

.cfg-summary-label { font-size: 12px; color: var(--text-muted); font-weight: 500; }
.cfg-summary-val   { font-size: 13px; font-weight: 700; color: var(--text-primary); }
.cfg-summary-val--green { color: var(--color-success); }
.cfg-summary-val--red   { color: var(--color-danger); }
.cfg-summary-val--muted { color: var(--text-muted); font-weight: 500; }

.cfg-notice-text {
  font-size: 12px; color: var(--text-muted); line-height: 1.6;
}
</style>
