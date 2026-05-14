<template>
  <el-dialog
    :model-value="modelValue"
    :title="t('settlements.triggerTitle')"
    @update:model-value="$emit('update:modelValue', $event)"
    width="90%"
    style="max-width: 600px"
    destroy-on-close
  >
    <div v-if="debtor">
      <!-- Debtor summary -->
      <div class="debtor-block">
        <div class="debtor-icon">
          <el-icon :size="20" color="white"><Warning /></el-icon>
        </div>
        <div class="debtor-info">
          <p class="debtor-name">{{ debtor.name }}</p>
          <p class="debtor-sub">{{ t('settlements.triggerSubtitle') }}</p>
        </div>
        <div class="debtor-amount">{{ formatVND(totalAmount) }}</div>
      </div>

      <!-- Distribution (based on allocated points) -->
      <div class="dist-grid">
        <div class="dist-item dist-item--green">
          <span class="dist-label">{{ t('settlements.fundShare', { percent: fundSplitPercent }) }}</span>
          <span class="dist-val">{{ formatVND(fundAmount) }}</span>
        </div>
        <div class="dist-item dist-item--blue">
          <span class="dist-label">{{ t('settlements.winnersShare', { percent: 100 - fundSplitPercent }) }}</span>
          <span class="dist-val">{{ formatVND(winnerAmount) }}</span>
        </div>
      </div>

      <!-- Debtor outcome preview -->
      <div v-if="selectedWinnerIds.length > 0" class="debtor-outcome">
        <span class="debtor-outcome-label">{{ t('settlements.debtorAfter', { name: debtor!.name }) }}</span>
        <span class="debtor-outcome-score" :class="debtorScoreAfter < 0 ? 'score-neg' : 'score-zero'">
          {{ debtorScoreAfter > 0 ? '+' : '' }}{{ debtorScoreAfter }} {{ t('common.pointsUnit') }}
        </span>
      </div>

      <!-- Winner selection -->
      <div class="section">
        <p class="section-label">{{ t('settlements.selectWinners') }} <span style="color:var(--color-danger)">*</span></p>
        <p class="section-hint">{{ t('settlements.eligibleHint') }}</p>
        <el-select
          v-model="selectedWinnerIds"
          multiple
          :placeholder="t('settlements.selectWinnersPlaceholder')"
          class="w-full"
          size="large"
        >
          <el-option
            v-for="user in eligibleWinners"
            :key="user.id"
            :label="`${user.name} (+${user.current_score} ${t('common.pointsUnit')})`"
            :value="user.id"
          >
            <div class="opt-row">
              <span>{{ user.name }}</span>
              <span class="opt-score">+{{ user.current_score }}</span>
            </div>
          </el-option>
        </el-select>
      </div>

      <!-- Per-winner point allocation -->
      <div v-if="selectedWinnerIds.length > 0" class="section">
        <div class="alloc-header">
          <p class="section-label">{{ t('settlements.allocatePoints') }}</p>
          <span
            class="alloc-badge"
            :class="unallocated === 0 ? 'alloc-badge--ok' : unallocated > 0 ? 'alloc-badge--warn' : 'alloc-badge--err'"
          >
            <template v-if="unallocated > 0">{{ t('settlements.unallocated', { points: unallocated }) }}</template>
            <template v-else-if="unallocated < 0">{{ t('settlements.overAllocated', { points: -unallocated }) }}</template>
            <template v-else>{{ t('settlements.fullyAllocated') }}</template>
          </span>
        </div>

        <div class="winner-alloc-list">
          <div v-for="w in winnerDistribution" :key="w.id" class="winner-alloc-row">
            <div class="winner-alloc-top">
              <span class="winner-alloc-name">{{ w.name }}</span>
              <div class="winner-alloc-meta">
                <span class="winner-alloc-score">+{{ w.currentScore }} {{ t('common.pointsUnit') }}</span>
                <span class="winner-alloc-arrow">→</span>
                <span
                  class="winner-alloc-after"
                  :class="w.currentScore - w.pointsToDeduct < 0 ? 'score-negative' : ''"
                >
                  {{ w.currentScore - w.pointsToDeduct >= 0 ? '+' : '' }}{{ w.currentScore - w.pointsToDeduct }}
                </span>
              </div>
            </div>
            <div class="winner-alloc-slider-row">
              <el-slider
                v-model="pointAllocations[w.id]"
                :min="1"
                :max="w.currentScore"
                :show-tooltip="false"
                size="small"
                class="winner-alloc-slider"
              />
              <div class="winner-alloc-input-wrap">
                <el-input-number
                  v-model="pointAllocations[w.id]"
                  :min="1"
                  :max="w.currentScore"
                  :controls="false"
                  size="small"
                  class="winner-alloc-input"
                />
                <span class="pts-label">{{ t('common.pointsUnit') }}</span>
              </div>
            </div>
            <div class="winner-alloc-money">
              {{ formatVND(Math.floor((winnerAmount * w.pointsToDeduct) / debtPoints)) }}
            </div>
          </div>
        </div>
      </div>

      <!-- Warning -->
      <el-alert type="warning" :closable="false" show-icon class="mt-4">
        <template #title>
          <span style="font-size:13px">
            {{ t('settlements.postSettlementWarning', { name: debtor.name }) }}
          </span>
        </template>
      </el-alert>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleCancel">{{ t('common.cancel') }}</el-button>
        <el-button
          type="primary"
          size="large"
          @click="handleConfirm"
          :disabled="selectedWinnerIds.length === 0"
          :loading="loading"
        >
          {{ t('settlements.confirmSettlement') }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import { Warning } from '@element-plus/icons-vue'
import type { User } from '@/types/user'
import type { WinnerInput } from '@/types/settlement'
import { formatVND } from '@/utils/formatters'

interface Props {
  modelValue: boolean
  debtor: User | null
  users: User[]
  pointToVnd: number
  fundSplitPercent: number
  loading?: boolean
}

const props = withDefaults(defineProps<Props>(), { loading: false })

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  confirm: [winners: WinnerInput[]]
  cancel: []
}>()
const { t } = useI18n()

const selectedWinnerIds = ref<string[]>([])
const pointAllocations = reactive<Record<string, number>>({})

const eligibleWinners = computed(() => {
  if (!props.debtor) return []
  return props.users.filter(u => u.id !== props.debtor!.id && u.current_score > 0 && u.is_active)
})

const debtPoints = computed(() => props.debtor ? Math.abs(props.debtor.current_score) : 0)

const winnerDistribution = computed(() =>
  selectedWinnerIds.value.map(id => {
    const user = props.users.find(u => u.id === id)
    return {
      id,
      name: user?.name || t('errors.notFound'),
      currentScore: user?.current_score ?? 0,
      pointsToDeduct: pointAllocations[id] ?? 1,
    }
  })
)

const totalAllocated = computed(() =>
  winnerDistribution.value.reduce((sum, w) => sum + w.pointsToDeduct, 0)
)

// All money values based on what is actually allocated, not the full debt
const totalAmount  = computed(() => totalAllocated.value * props.pointToVnd)
const fundAmount   = computed(() => Math.floor((totalAmount.value * props.fundSplitPercent) / 100))
const winnerAmount = computed(() => totalAmount.value - fundAmount.value)

const debtorScoreAfter = computed(() =>
  props.debtor ? props.debtor.current_score + totalAllocated.value : 0
)

const unallocated = computed(() => debtPoints.value - totalAllocated.value)

// When the selected winner list changes, re-initialise allocations
watch(selectedWinnerIds, (ids) => {
  const n = ids.length
  if (n === 0) return

  // Distribute debt equally, cap each at the winner's score
  const base = Math.floor(debtPoints.value / n)
  let remainder = debtPoints.value - base * n

  ids.forEach(id => {
    const user = props.users.find(u => u.id === id)
    const maxAlloc = user?.current_score ?? debtPoints.value
    const alloc = Math.min(base, maxAlloc)
    pointAllocations[id] = alloc
  })

  // Distribute remainder to winners who still have capacity
  for (const id of ids) {
    if (remainder <= 0) break
    const user = props.users.find(u => u.id === id)
    const maxAlloc = user?.current_score ?? debtPoints.value
    const canAdd = maxAlloc - (pointAllocations[id] ?? 0)
    const add = Math.min(remainder, canAdd)
    pointAllocations[id] = (pointAllocations[id] ?? 0) + add
    remainder -= add
  }
}, { deep: true })

watch(() => props.modelValue, v => {
  if (!v) {
    selectedWinnerIds.value = []
    Object.keys(pointAllocations).forEach(k => delete pointAllocations[k])
  }
})

const handleConfirm = () => {
  const winners: WinnerInput[] = winnerDistribution.value.map(w => ({
    id: w.id,
    points_to_deduct: w.pointsToDeduct,
  }))
  emit('confirm', winners)
}
const handleCancel = () => { emit('cancel'); emit('update:modelValue', false) }
</script>

<style scoped>
.debtor-block {
  display: flex; align-items: center; gap: 12px;
  padding: 16px; border-radius: 14px;
  background: var(--color-danger-bg);
  border: 1px solid var(--color-danger-border);
  margin-bottom: 16px;
}

.debtor-icon {
  width: 40px; height: 40px; border-radius: 10px; flex-shrink: 0;
  background: linear-gradient(135deg, #ef4444, #b91c1c);
  display: flex; align-items: center; justify-content: center;
  box-shadow: 0 4px 10px rgba(239,68,68,0.3);
}

.debtor-info { flex: 1; min-width: 0; }
.debtor-name { font-size: 15px; font-weight: 700; color: var(--color-danger); }
.debtor-sub  { font-size: 11px; color: var(--text-muted); margin-top: 2px; }
.debtor-amount { font-size: 17px; font-weight: 800; color: var(--color-danger); flex-shrink: 0; }

.dist-grid {
  display: grid; grid-template-columns: 1fr 1fr; gap: 10px; margin-bottom: 16px;
}

.dist-item {
  display: flex; flex-direction: column; gap: 4px;
  padding: 12px 14px; border-radius: 10px; border: 1px solid;
}
.dist-item--green { background: var(--color-success-bg); border-color: var(--color-success-border); }
.dist-item--blue  { background: var(--color-info-bg);    border-color: var(--color-info-border); }

.dist-label {
  font-size: 11px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.04em;
}
.dist-item--green .dist-label { color: var(--color-success); }
.dist-item--blue  .dist-label { color: var(--color-info); }

.dist-val { font-size: 16px; font-weight: 800; }
.dist-item--green .dist-val { color: var(--color-success); }
.dist-item--blue  .dist-val { color: var(--color-info); }

.section { margin-bottom: 16px; }
.section-label { font-size: 13px; font-weight: 600; color: var(--text-primary); margin-bottom: 4px; }
.section-hint  { font-size: 11px; color: var(--text-muted); margin-bottom: 8px; }

.opt-row { display: flex; justify-content: space-between; align-items: center; }
.opt-score { font-size: 12px; font-weight: 700; color: var(--color-success); }

/* Allocation header */
.alloc-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px; }
.alloc-header .section-label { margin-bottom: 0; }

.alloc-badge {
  font-size: 11px; font-weight: 700; padding: 2px 8px; border-radius: 20px;
}
.alloc-badge--ok   { background: var(--color-success-bg); color: var(--color-success); border: 1px solid var(--color-success-border); }
.alloc-badge--warn { background: var(--color-warning-bg, #fef9ec); color: var(--color-warning, #d97706); border: 1px solid var(--color-warning-border, #fde68a); }
.alloc-badge--err  { background: var(--color-danger-bg); color: var(--color-danger); border: 1px solid var(--color-danger-border); }

/* Winner allocation list */
.winner-alloc-list {
  border: 1px solid var(--border-default);
  border-radius: 10px; overflow: hidden;
}

.winner-alloc-row {
  padding: 12px 14px;
  border-bottom: 1px solid var(--border-subtle);
}
.winner-alloc-row:last-child { border-bottom: none; }

.winner-alloc-top {
  display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px;
}

.winner-alloc-name { font-size: 13px; font-weight: 600; color: var(--text-primary); }

.winner-alloc-meta { display: flex; align-items: center; gap: 6px; font-size: 12px; }
.winner-alloc-score { color: var(--color-success); font-weight: 700; }
.winner-alloc-arrow { color: var(--text-muted); }
.winner-alloc-after { font-weight: 700; color: var(--color-success); }
.winner-alloc-after.score-negative { color: var(--color-danger); }

.winner-alloc-slider-row {
  display: flex; align-items: center; gap: 10px;
}
.winner-alloc-slider { flex: 1; }
.winner-alloc-input-wrap { display: flex; align-items: center; gap: 4px; flex-shrink: 0; }
.winner-alloc-input { width: 56px; }
.pts-label { font-size: 11px; color: var(--text-muted); }

.winner-alloc-money {
  font-size: 11px; color: var(--color-success); font-weight: 600;
  text-align: right; margin-top: 4px;
}

.score-negative { color: var(--color-danger) !important; }

.debtor-outcome {
  display: flex; align-items: center; justify-content: space-between;
  padding: 10px 14px; border-radius: 10px; margin-bottom: 16px;
  background: var(--surface-card); border: 1px solid var(--border-default);
}
.debtor-outcome-label { font-size: 12px; color: var(--text-muted); }
.debtor-outcome-score { font-size: 15px; font-weight: 800; }
.score-neg  { color: var(--color-danger); }
.score-zero { color: var(--text-muted); }

.dialog-footer { display: flex; justify-content: flex-end; gap: 10px; }
</style>
