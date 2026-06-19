<template>
  <div>
    <div v-if="predictions.length === 0" class="empty-state">
      <div class="empty-state-title">{{ t('wc.noPredictions') }}</div>
    </div>
    <div v-else class="wc-bet-list">
      <div v-for="pred in predictions" :key="pred.id" class="wc-bet-row">
        <div class="wc-bet-main">
          <div class="wc-bet-match-info">
            <span class="wc-bet-teams">{{ pred.home_team }} vs {{ pred.away_team }}</span>
            <span class="wc-bet-date">{{ formatDate(pred.created_at) }}</span>
          </div>
          <div class="wc-bet-details">
            <span class="wc-bet-type">
              {{ pred.prediction_type === 'handicap' ? t('wc.predictionTypeHandicap') : pred.prediction_type === 'over_under' ? t('wc.predictionTypeOverUnder') : t('wc.predictionTypeExactScore') }}
            </span>
            <span class="wc-bet-choice">
              <template v-if="pred.prediction_type === 'handicap'">
                {{ pred.prediction_choice === 'home' ? pred.home_team : pred.away_team }}
              </template>
              <template v-else>
                {{ pred.predicted_home_score }}–{{ pred.predicted_away_score }}
              </template>
            </span>

            <!-- Inline points edit -->
            <template v-if="editingId === pred.id">
              <el-input-number
                v-model="editPoints"
                :min="1"
                size="small"
                controls-position="right"
                style="width: 110px"
              />
              <el-button size="small" type="success" :loading="saving" @click="saveEdit(pred)">✓</el-button>
              <el-button size="small" text @click="cancelEdit">✕</el-button>
            </template>
            <template v-else>
              <span class="wc-bet-stake">{{ pred.points }} × {{ pred.multiplier_snapshot.toFixed(2) }}</span>
              <template v-if="isEditable(pred)">
                <!-- <el-button
                  size="small"
                  text
                  class="wc-bet-action-btn"
                  @click="startEdit(pred)"
                >Sửa</el-button> -->
                <el-button
                  size="small"
                  text
                  class="wc-bet-action-btn wc-bet-action-btn--delete"
                  :loading="deletingId === pred.id"
                  @click="handleDelete(pred)"
                >Xoá</el-button>
              </template>
            </template>
          </div>
        </div>

        <div class="wc-bet-result">
          <span v-if="!pred.result" class="wc-result-badge wc-result--pending">
            {{ t('wc.resultPending') }}
          </span>
          <span v-else-if="pred.result === 'correct'" class="wc-result-badge wc-result--correct">
            +{{ fmtPts((pred.points_earned ?? 0) - pred.points) }} {{ t('wc.resultCorrect') }}
          </span>
          <span v-else-if="pred.result === 'win_half'" class="wc-result-badge wc-result--win-half">
            +{{ fmtPts((pred.points_earned ?? 0) - pred.points) }} {{ t('wc.resultWinHalf') }}
          </span>
          <span v-else-if="pred.result === 'lose_half'" class="wc-result-badge wc-result--lose-half">
            -{{ fmtPts(pred.points - (pred.points_earned ?? 0)) }} {{ t('wc.resultLoseHalf') }}
          </span>
          <span v-else-if="pred.result === 'incorrect'" class="wc-result-badge wc-result--incorrect">
            -{{ pred.points }} {{ t('wc.resultIncorrect') }}
          </span>
          <span v-else class="wc-result-badge wc-result--void">
            ±0 {{ t('wc.resultVoid') }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessageBox } from 'element-plus'
import { useWcStore } from '@/stores/wcStore'
import type { WcPredictionWithMatch } from '@/types/wc'

const { t } = useI18n()
const store = useWcStore()

function fmtPts(v: number): string {
  return parseFloat(v.toFixed(2)).toString()
}

defineProps<{ predictions: WcPredictionWithMatch[] }>()

const editingId = ref<string | null>(null)
const editPoints = ref(0)
const saving = ref(false)
const deletingId = ref<string | null>(null)

function isEditable(pred: WcPredictionWithMatch): boolean {
  if (pred.result) return false
  if (pred.match_status === 'completed' || pred.match_status === 'cancelled') return false
  if (!pred.predictions_open) return false
  if (pred.predictions_locked_at && new Date(pred.predictions_locked_at) <= new Date()) return false
  return true
}

function cancelEdit() {
  editingId.value = null
}

async function saveEdit(pred: WcPredictionWithMatch) {
  saving.value = true
  try {
    await store.updatePredictionPoints(pred.id, editPoints.value)
    editingId.value = null
  } finally {
    saving.value = false
  }
}

async function handleDelete(pred: WcPredictionWithMatch) {
  await ElMessageBox.confirm(
    `Xoá dự đoán ${pred.prediction_type === 'handicap' ? (pred.prediction_choice === 'home' ? pred.home_team : pred.away_team) : `${pred.predicted_home_score}–${pred.predicted_away_score}`} (${pred.points} pts)?`,
    'Xác nhận xoá dự đoán',
    { confirmButtonText: 'Xoá', cancelButtonText: 'Hủy', type: 'warning' },
  )
  deletingId.value = pred.id
  try {
    await store.deletePrediction(pred.id)
  } finally {
    deletingId.value = null
  }
}

function formatDate(s: string) {
  return new Date(s).toLocaleDateString('vi-VN', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' })
}
</script>

<style scoped>
.wc-bet-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.wc-bet-row {
  background: var(--surface-card);
  border: 1px solid var(--border-default);
  border-radius: 10px;
  padding: 12px 14px;
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.wc-bet-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.wc-bet-match-info {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.wc-bet-teams {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.wc-bet-date {
  font-size: 11px;
  color: var(--text-muted);
}

.wc-bet-details {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.wc-bet-type {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  background: var(--surface-page);
  padding: 2px 7px;
  border-radius: 4px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.wc-bet-choice {
  font-size: 13px;
  font-weight: 700;
  color: var(--text-primary);
}

.wc-bet-stake {
  font-size: 12px;
  color: var(--text-secondary);
  tabular-nums: true;
}

.wc-bet-action-btn {
  font-size: 14px;
  padding: 0 4px;
  color: var(--text-muted);
  opacity: 0.6;
  transition: opacity 0.15s, color 0.15s;
}

.wc-bet-action-btn:hover {
  opacity: 1;
  color: var(--text-primary);
}

.wc-bet-action-btn--delete:hover {
  color: #ef4444;
}

.wc-bet-result {
  flex-shrink: 0;
}

.wc-result-badge {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 700;
  tabular-nums: true;
}

.wc-result--pending {
  background: rgba(100, 116, 139, 0.1);
  color: #64748b;
}

.wc-result--correct {
  background: rgba(22, 163, 74, 0.12);
  color: #16a34a;
}

.wc-result--win-half {
  background: rgba(22, 163, 74, 0.08);
  color: #16a34a;
}

.wc-result--lose-half {
  background: rgba(239, 68, 68, 0.07);
  color: #ef4444;
}

.wc-result--incorrect {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.wc-result--void {
  background: rgba(217, 119, 6, 0.1);
  color: #d97706;
}
</style>
