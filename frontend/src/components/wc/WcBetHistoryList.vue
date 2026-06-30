<template>
  <div>
    <div v-if="bets.length === 0" class="empty-state">
      <div class="empty-state-title">{{ t('wc.noBets') }}</div>
    </div>
    <div v-else class="wc-bet-list">
      <div v-for="bet in bets" :key="bet.id" class="wc-bet-row">
        <div class="wc-bet-main">
          <div class="wc-bet-match-info">
            <span class="wc-bet-teams">{{ bet.home_team }} vs {{ bet.away_team }}</span>
            <span class="wc-bet-date">{{ formatDate(bet.created_at) }}</span>
          </div>
          <div class="wc-bet-details">
            <span class="wc-bet-type">
              {{ betTypeLabel(bet.bet_type) }}
            </span>
            <span class="wc-bet-choice">
              <template v-if="bet.bet_type === 'handicap'">
                {{ bet.bet_choice === 'home' ? bet.home_team : bet.away_team }}
              </template>
              <template v-else>
                {{ bet.predicted_home_score }}–{{ bet.predicted_away_score }}
              </template>
            </span>

            <!-- Inline stake edit -->
            <template v-if="editingId === bet.id">
              <el-input-number
                v-model="editStake"
                :min="store.minPoints"
                :max="store.maxPoints"
                size="small"
                controls-position="right"
                style="width: 110px"
              />
              <el-button size="small" type="success" :loading="saving" @click="saveEdit(bet)">✓</el-button>
              <el-button size="small" text @click="cancelEdit">✕</el-button>
            </template>
            <template v-else>
              <span class="wc-bet-stake">{{ bet.stake }} × {{ bet.odds_snapshot.toFixed(2) }}</span>
              <template v-if="isEditable(bet)">
                <el-button
                  size="small"
                  text
                  class="wc-bet-action-btn"
                  @click="startEdit(bet)"
                >Sửa</el-button>
                <el-button
                  size="small"
                  text
                  class="wc-bet-action-btn wc-bet-action-btn--delete"
                  :loading="deletingId === bet.id"
                  @click="handleDelete(bet)"
                >Xoá</el-button>
              </template>
            </template>
          </div>
        </div>

        <div class="wc-bet-result">
          <span v-if="!bet.result" class="wc-result-badge wc-result--pending">
            {{ t('wc.resultPending') }}
          </span>
          <span v-else-if="bet.result === 'win'" class="wc-result-badge wc-result--win">
            +{{ (bet.payout ?? 0) - bet.stake }} {{ t('wc.resultWin') }}
          </span>
          <span v-else-if="bet.result === 'win_half'" class="wc-result-badge wc-result--win-half">
            +{{ (bet.payout ?? 0) - bet.stake }} {{ t('wc.resultWinHalf') }}
          </span>
          <span v-else-if="bet.result === 'lose_half'" class="wc-result-badge wc-result--lose-half">
            -{{ bet.stake - (bet.payout ?? 0) }} {{ t('wc.resultLoseHalf') }}
          </span>
          <span v-else-if="bet.result === 'lose'" class="wc-result-badge wc-result--lose">
            -{{ bet.stake }} {{ t('wc.resultLose') }}
          </span>
          <span v-else class="wc-result-badge wc-result--push">
            ±0 {{ t('wc.resultPush') }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useWcStore } from '@/stores/wcStore'
import { wcService } from '@/services/wcService'
import { useWcBetTypeLabel } from '@/utils/wcBetType'
import type { WcBetWithMatch } from '@/types/wc'

const { t } = useI18n()
const betTypeLabel = useWcBetTypeLabel()
const store = useWcStore()

defineProps<{ bets: WcBetWithMatch[] }>()

const editingId = ref<string | null>(null)
const editStake = ref(0)
const saving = ref(false)
const deletingId = ref<string | null>(null)

function isEditable(bet: WcBetWithMatch): boolean {
  if (bet.result) return false
  if (bet.cancelled_at) return false
  if (bet.match_status === 'live' || bet.match_status === 'completed' || bet.match_status === 'cancelled') return false
  if (new Date(bet.match_date).getTime() <= Date.now()) return false
  if (bet.bets_locked_at && new Date(bet.bets_locked_at) <= new Date()) return false
  return true
}

function startEdit(bet: WcBetWithMatch) {
  editingId.value = bet.id
  editStake.value = bet.stake
}

function cancelEdit() {
  editingId.value = null
}

async function saveEdit(bet: WcBetWithMatch) {
  const newStake = editStake.value
  if (newStake < bet.stake) {
    try {
      const preview = await wcService.previewReduceStake(bet.id, newStake)
      if (preview.penalty > 0) {
        await ElMessageBox.confirm(
          t('wc.reducePenaltyWarning', { max: store.betReduceMaxPercent, penalty: preview.penalty }),
          t('wc.reduceStakeTitle'),
          { confirmButtonText: t('wc.cancelConfirm'), cancelButtonText: 'Hủy', type: 'warning' },
        )
      }
    } catch {
      // user cancelled or preview failed — abort
      return
    }
  }
  saving.value = true
  try {
    await store.updateBetStake(bet.id, newStake)
    editingId.value = null
  } finally {
    saving.value = false
  }
}

async function handleDelete(bet: WcBetWithMatch) {
  const penalty = store.cancelPenaltyEnabled
    ? Math.floor(bet.stake * store.cancelPenaltyPercent / 100)
    : 0

  if (store.cancelPenaltyEnabled && penalty > 0) {
    try {
      await ElMessageBox.confirm(
        t('wc.cancelPenaltyWarning', { penalty }),
        t('wc.cancelBetTitle'),
        { confirmButtonText: t('wc.cancelConfirm'), cancelButtonText: 'Hủy', type: 'warning' },
      )
    } catch {
      return
    }
  } else {
    try {
      await ElMessageBox.confirm(
        `Xoá cược này (${bet.stake} pts)?`,
        t('wc.cancelBetTitle'),
        { confirmButtonText: 'Xoá', cancelButtonText: 'Hủy', type: 'warning' },
      )
    } catch {
      return
    }
  }

  deletingId.value = bet.id
  try {
    await store.deleteBet(bet.id)
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : 'Lỗi khi huỷ cược'
    ElMessage.error(msg)
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

.wc-result--win {
  background: rgba(22, 163, 74, 0.12);
  color: #16a34a;
}

.wc-result--lose {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.wc-result--win-half {
  background: rgba(22, 163, 74, 0.08);
  color: #16a34a;
}

.wc-result--lose-half {
  background: rgba(239, 68, 68, 0.07);
  color: #ef4444;
}

.wc-result--push {
  background: rgba(217, 119, 6, 0.1);
  color: #d97706;
}
</style>
