<template>
  <div>
    <div v-if="predictions.length === 0" class="empty-state">
      <div class="empty-state-title">{{ t('wc.noPredictions') }}</div>
    </div>
    <div v-else class="wc-pred-list">
      <template v-for="item in displayItems" :key="item.kind === 'header' ? item.dateKey : item.pred.id">

        <!-- Date group header -->
        <button v-if="item.kind === 'header'" class="wc-date-header" @click="toggleGroup(item.dateKey)">
          <span class="wc-date-label">{{ item.label }}</span>
          <span class="wc-date-meta">
            <span class="wc-date-count">{{ item.count }} dự đoán</span>
            <span class="wc-date-net" :class="item.netPoints >= 0 ? 'wc-date-net--pos' : 'wc-date-net--neg'">
              {{ item.netPoints >= 0 ? '+' : '' }}{{ fmtPts(item.netPoints) }} điểm
            </span>
          </span>
          <span class="wc-date-chevron" :class="{ 'wc-date-chevron--collapsed': collapsedDates.has(item.dateKey) }">▾</span>
        </button>

        <!-- Prediction row -->
        <div v-else-if="!isHidden(item)" class="wc-bet-row">
          <div class="wc-bet-main">
            <div class="wc-bet-match-info">
              <span class="wc-bet-teams">{{ item.pred.home_team }} vs {{ item.pred.away_team }}</span>
              <WcHandicapLine
                v-if="item.pred.prediction_type === 'handicap'"
                :homeTeam="item.pred.home_team"
                :awayTeam="item.pred.away_team"
                :handicapValue="matchById(item.pred.match_id)?.handicap_value"
                :handicapTeam="matchById(item.pred.match_id)?.handicap_team"
              />
              <span class="wc-bet-date">{{ formatDate(item.pred.created_at) }}</span>
            </div>
            <div class="wc-bet-details">
              <span class="wc-bet-type">{{ betTypeLabel(item.pred.prediction_type) }}</span>
              <span class="wc-bet-choice">
                <template v-if="item.pred.prediction_type === 'handicap'">
                  {{ item.pred.prediction_choice === 'home' ? item.pred.home_team : item.pred.away_team }}
                  <span v-if="item.pred.handicap_snapshot != null" class="wc-pred-hcap-snap">
                    {{ handicapLabel(item.pred) }}
                  </span>
                </template>
                <template v-else-if="item.pred.prediction_type === 'over_under'">
                  {{ item.pred.prediction_choice === 'over' ? t('wc.choiceOver') : t('wc.choiceUnder') }}
                  <span v-if="ouLine(item.pred) != null" class="wc-pred-ou-line">{{ ouLine(item.pred) }}</span>
                </template>
                <template v-else-if="item.pred.prediction_type === 'custom'">
                  <span v-if="item.pred.bet_title" class="wc-bet-custom-title">{{ item.pred.bet_title }}</span>
                  {{ item.pred.prediction_choice }}
                </template>
                <template v-else>
                  {{ item.pred.predicted_home_score }}–{{ item.pred.predicted_away_score }}
                </template>
              </span>

              <!-- Cancel custom bet (pending) -->
              <template v-if="item.pred.prediction_type === 'custom' && !item.pred.result && item.pred.match_status !== 'live'">
                <el-button
                  size="small" text type="danger"
                  class="wc-bet-action-btn wc-bet-action-btn--delete"
                  :loading="cancellingId === item.pred.id"
                  @click="handleCancelCustom(item.pred)"
                >Huỷ</el-button>
              </template>
              <template v-else-if="editingId === item.pred.id">
                <el-input-number
                  v-model="editPoints"
                  :min="store.minPoints"
                  :max="store.maxPoints"
                  size="small"
                  controls-position="right"
                  style="width: 110px"
                />
                <el-button size="small" type="success" :loading="saving" @click="saveEdit(item.pred)">✓</el-button>
                <el-button size="small" text @click="cancelEdit">✕</el-button>
              </template>
              <template v-else>
                <span class="wc-bet-stake">{{ item.pred.points }} × {{ item.pred.multiplier_snapshot.toFixed(2) }}</span>
                <template v-if="isEditable(item.pred)">
                  <el-button
                    size="small" text
                    class="wc-bet-action-btn wc-bet-action-btn--delete"
                    :loading="deletingId === item.pred.id"
                    @click="handleDelete(item.pred)"
                  >Xoá</el-button>
                </template>
              </template>
            </div>

            <!-- Potential outcomes — only for pending predictions -->
            <div v-if="!item.pred.result" class="wc-pred-outcomes">
              <template v-if="item.pred.prediction_type === 'handicap' || item.pred.prediction_type === 'over_under'">
                <span class="wc-pred-outcome wc-pred-outcome--win">Thắng +{{ fmtPts(item.pred.points * (item.pred.multiplier_snapshot - 1)) }}</span>
                <span class="wc-pred-sep">·</span>
                <span class="wc-pred-outcome wc-pred-outcome--win-half">Thắng½ +{{ fmtPts(item.pred.points * (item.pred.multiplier_snapshot - 1) / 2) }}</span>
                <span class="wc-pred-sep">·</span>
                <span class="wc-pred-outcome wc-pred-outcome--lose-half">Thua½ -{{ fmtPts(item.pred.points / 2) }}</span>
                <span class="wc-pred-sep">·</span>
                <span class="wc-pred-outcome wc-pred-outcome--lose">Thua -{{ item.pred.points }}</span>
              </template>
              <template v-else-if="item.pred.prediction_type === 'exact_score'">
                <span class="wc-pred-outcome wc-pred-outcome--win">Đúng +{{ fmtPts(item.pred.points * (item.pred.multiplier_snapshot - 1)) }}</span>
                <span class="wc-pred-sep">·</span>
                <span class="wc-pred-outcome wc-pred-outcome--lose">Sai -{{ item.pred.points }}</span>
              </template>
              <template v-else-if="item.pred.prediction_type === 'custom'">
                <span class="wc-pred-outcome wc-pred-outcome--win">Thắng +{{ fmtPts(item.pred.points * (item.pred.multiplier_snapshot - 1)) }}</span>
                <span class="wc-pred-sep">·</span>
                <span class="wc-pred-outcome wc-pred-outcome--lose">Thua -{{ item.pred.points }}</span>
              </template>
            </div>
          </div>

          <div class="wc-bet-result">
            <span v-if="!item.pred.result" class="wc-result-badge wc-result--pending">
              {{ t('wc.resultPending') }}
            </span>
            <span v-else-if="item.pred.result === 'correct'" class="wc-result-badge wc-result--correct">
              +{{ fmtPts((item.pred.points_earned ?? 0) - item.pred.points) }} {{ t('wc.resultCorrect') }}
            </span>
            <span v-else-if="item.pred.result === 'win_half'" class="wc-result-badge wc-result--win-half">
              +{{ fmtPts((item.pred.points_earned ?? 0) - item.pred.points) }} {{ t('wc.resultWinHalf') }}
            </span>
            <span v-else-if="item.pred.result === 'lose_half'" class="wc-result-badge wc-result--lose-half">
              -{{ fmtPts(item.pred.points - (item.pred.points_earned ?? 0)) }} {{ t('wc.resultLoseHalf') }}
            </span>
            <span v-else-if="item.pred.result === 'incorrect'" class="wc-result-badge wc-result--incorrect">
              -{{ item.pred.points }} {{ t('wc.resultIncorrect') }}
            </span>
            <span v-else class="wc-result-badge wc-result--void">
              ±0 {{ t('wc.resultVoid') }}
            </span>
          </div>
        </div>

      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useWcStore } from '@/stores/wcStore'
import { useWcBetTypeLabel } from '@/utils/wcBetType'
import WcHandicapLine from '@/components/wc/WcHandicapLine.vue'
import { wcService } from '@/services/wcService'
import type { WcPredictionWithMatch } from '@/types/wc'

const { t } = useI18n()
const betTypeLabel = useWcBetTypeLabel()
const store = useWcStore()

type HeaderItem = { kind: 'header'; dateKey: string; label: string; netPoints: number; count: number }
type PredItem   = { kind: 'pred';   pred: WcPredictionWithMatch; groupDateKey: string | null }
type DisplayItem = HeaderItem | PredItem

const props = defineProps<{ predictions: WcPredictionWithMatch[]; groupByDate?: boolean }>()

const collapsedDates = ref<Set<string>>(new Set())

function toggleGroup(dateKey: string) {
  const s = new Set(collapsedDates.value)
  s.has(dateKey) ? s.delete(dateKey) : s.add(dateKey)
  collapsedDates.value = s
}

function isHidden(item: PredItem): boolean {
  return !!item.groupDateKey && collapsedDates.value.has(item.groupDateKey)
}

function dateKey(pred: WcPredictionWithMatch): string {
  // sv-SE locale gives YYYY-MM-DD in the browser's local timezone (GMT+7 for Vietnam)
  return new Date(pred.match_date).toLocaleDateString('sv-SE')
}

function dateLabel(key: string): string {
  return new Date(key + 'T00:00:00').toLocaleDateString('vi-VN', {
    weekday: 'long', day: '2-digit', month: '2-digit', year: 'numeric',
  })
}

const displayItems = computed((): DisplayItem[] => {
  if (!props.groupByDate) {
    return props.predictions.map(pred => ({ kind: 'pred', pred, groupDateKey: null }))
  }

  // Group by match date descending
  const groups = new Map<string, WcPredictionWithMatch[]>()
  for (const pred of props.predictions) {
    const key = dateKey(pred)
    if (!groups.has(key)) groups.set(key, [])
    groups.get(key)!.push(pred)
  }
  const sortedKeys = [...groups.keys()].sort((a, b) => b.localeCompare(a))

  const items: DisplayItem[] = []
  for (const key of sortedKeys) {
    const preds = groups.get(key)!
    const netPoints = preds.reduce((sum, p) => sum + (p.points_earned ?? 0) - p.points, 0)
    items.push({ kind: 'header', dateKey: key, label: dateLabel(key), netPoints, count: preds.length })
    for (const pred of preds) {
      items.push({ kind: 'pred', pred, groupDateKey: key })
    }
  }
  return items
})

// --- helpers ---

function matchById(matchId: string) {
  return store.matches.find(m => m.id === matchId)
}

function fmtPts(v: number): string {
  return parseFloat(v.toFixed(2)).toString()
}

function ouLine(pred: WcPredictionWithMatch): number | null | undefined {
  return matchById(pred.match_id)?.ou_line
}

function handicapLabel(pred: WcPredictionWithMatch): string {
  const val = pred.handicap_snapshot
  if (val == null) return ''
  const pickedHome = pred.prediction_choice === 'home'
  const hcapOnHome = pred.handicap_team_snapshot === 'home'
  const adjustedForPicked = pickedHome === hcapOnHome ? -val : val
  const sign = adjustedForPicked > 0 ? '+' : ''
  return `(${sign}${adjustedForPicked})`
}

function formatDate(s: string) {
  return new Date(s).toLocaleDateString('vi-VN', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' })
}

// --- edit / delete state ---

const editingId    = ref<string | null>(null)
const editPoints   = ref(0)
const saving       = ref(false)
const deletingId   = ref<string | null>(null)
const cancellingId = ref<string | null>(null)

function isEditable(pred: WcPredictionWithMatch): boolean {
  if (pred.result) return false
  if (pred.match_status === 'completed' || pred.match_status === 'cancelled') return false
  if (!pred.predictions_open) return false
  if (pred.predictions_locked_at && new Date(pred.predictions_locked_at) <= new Date()) return false
  return true
}

function cancelEdit() { editingId.value = null }

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
  const label = pred.prediction_type === 'handicap'
    ? (pred.prediction_choice === 'home' ? pred.home_team : pred.away_team)
    : pred.prediction_type === 'over_under'
    ? (pred.prediction_choice === 'over' ? t('wc.choiceOver') : t('wc.choiceUnder'))
    : `${pred.predicted_home_score}–${pred.predicted_away_score}`
  await ElMessageBox.confirm(
    `Xoá dự đoán ${label} (${pred.points} pts)?`,
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

async function handleCancelCustom(pred: WcPredictionWithMatch) {
  await ElMessageBox.confirm(
    `Huỷ cược kèo phụ "${pred.bet_title}" — ${pred.prediction_choice} (${pred.points} pts)?`,
    'Xác nhận huỷ',
    { confirmButtonText: 'Huỷ cược', cancelButtonText: 'Đóng', type: 'warning' },
  )
  cancellingId.value = pred.id
  try {
    await wcService.cancelCustomBetEntry(pred.id)
    ElMessage.success('Đã huỷ cược')
    await store.fetchPredictions()
  } catch {
    // error shown by wcApi interceptor
  } finally {
    cancellingId.value = null
  }
}
</script>

<style scoped>
.wc-pred-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

/* --- Date group header --- */
.wc-date-header {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 8px 12px;
  background: var(--surface-page);
  border: 1px solid var(--border-default);
  border-radius: 8px;
  cursor: pointer;
  text-align: left;
  transition: background 0.12s;
}

.wc-date-header:hover {
  background: var(--surface-hover, rgba(0,0,0,0.04));
}

.wc-date-label {
  font-size: 13px;
  font-weight: 700;
  color: var(--text-primary);
  flex: 1;
  text-transform: capitalize;
}

.wc-date-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-shrink: 0;
}

.wc-date-count {
  font-size: 11px;
  color: var(--text-muted);
}

.wc-date-net {
  font-size: 12px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}

.wc-date-net--pos { color: #16a34a; }
.wc-date-net--neg { color: #ef4444; }

.wc-date-chevron {
  font-size: 14px;
  color: var(--text-muted);
  transition: transform 0.2s;
  display: inline-block;
}

.wc-date-chevron--collapsed {
  transform: rotate(-90deg);
}

/* --- Prediction row --- */
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
  display: flex;
  align-items: center;
  gap: 4px;
}

.wc-bet-custom-title {
  font-size: 10px;
  font-weight: 600;
  color: var(--el-color-warning);
  background: var(--el-color-warning-light-9);
  padding: 1px 5px;
  border-radius: 4px;
}

.wc-pred-hcap-snap,
.wc-pred-ou-line {
  font-size: 11px;
  font-weight: 700;
  color: var(--el-color-warning);
  background: rgba(230, 162, 60, 0.1);
  padding: 1px 5px;
  border-radius: 4px;
}

.wc-pred-outcomes {
  display: flex;
  align-items: center;
  gap: 5px;
  flex-wrap: wrap;
  margin-top: 2px;
}

.wc-pred-sep {
  font-size: 10px;
  color: var(--text-muted);
}

.wc-pred-outcome {
  font-size: 11px;
  font-weight: 600;
  padding: 1px 6px;
  border-radius: 4px;
}

.wc-pred-outcome--win       { color: #16a34a; background: rgba(22, 163, 74, 0.08); }
.wc-pred-outcome--win-half  { color: #16a34a; background: rgba(22, 163, 74, 0.05); }
.wc-pred-outcome--lose-half { color: #ef4444; background: rgba(239, 68, 68, 0.05); }
.wc-pred-outcome--lose      { color: #ef4444; background: rgba(239, 68, 68, 0.08); }

.wc-bet-stake {
  font-size: 12px;
  color: var(--text-secondary);
  font-variant-numeric: tabular-nums;
}

.wc-bet-action-btn {
  font-size: 14px;
  padding: 0 4px;
  color: var(--text-muted);
  opacity: 0.6;
  transition: opacity 0.15s, color 0.15s;
}

.wc-bet-action-btn:hover          { opacity: 1; color: var(--text-primary); }
.wc-bet-action-btn--delete:hover  { color: #ef4444; }

.wc-bet-result {
  flex-shrink: 0;
}

.wc-result-badge {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}

.wc-result--pending  { background: rgba(100, 116, 139, 0.1); color: #64748b; }
.wc-result--correct  { background: rgba(22, 163, 74, 0.12);  color: #16a34a; }
.wc-result--win-half { background: rgba(22, 163, 74, 0.08);  color: #16a34a; }
.wc-result--lose-half{ background: rgba(239, 68, 68, 0.07);  color: #ef4444; }
.wc-result--incorrect{ background: rgba(239, 68, 68, 0.1);   color: #ef4444; }
.wc-result--void     { background: rgba(217, 119, 6, 0.1);   color: #d97706; }
</style>
