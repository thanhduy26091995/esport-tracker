<template>
  <el-dialog
    :model-value="modelValue"
    title="Trigger Debt Settlement"
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
          <p class="debtor-sub">Debt Settlement</p>
        </div>
        <div class="debtor-amount">{{ formatVND(totalAmount) }}</div>
      </div>

      <!-- Distribution -->
      <div class="dist-grid">
        <div class="dist-item dist-item--green">
          <span class="dist-label">Fund ({{ fundSplitPercent }}%)</span>
          <span class="dist-val">{{ formatVND(fundAmount) }}</span>
        </div>
        <div class="dist-item dist-item--blue">
          <span class="dist-label">Winners ({{ 100 - fundSplitPercent }}%)</span>
          <span class="dist-val">{{ formatVND(winnerAmount) }}</span>
        </div>
      </div>

      <!-- Winner selection -->
      <div class="section">
        <p class="section-label">Select Winners <span style="color:var(--color-danger)">*</span></p>
        <p class="section-hint">Must have a positive score to be eligible.</p>
        <el-select
          v-model="selectedWinners"
          multiple
          placeholder="Select winners"
          class="w-full"
          size="large"
        >
          <el-option
            v-for="user in eligibleWinners"
            :key="user.id"
            :label="`${user.name} (+${user.current_score} pts)`"
            :value="user.id"
          >
            <div class="opt-row">
              <span>{{ user.name }}</span>
              <span class="opt-score">+{{ user.current_score }}</span>
            </div>
          </el-option>
        </el-select>
        <p v-if="selectedWinners.length > 0" class="section-hint mt-2">
          {{ selectedWinners.length }} winner(s) — each receives {{ formatVND(Math.floor(winnerAmount / selectedWinners.length)) }}
        </p>
      </div>

      <!-- Payout preview -->
      <div v-if="selectedWinners.length > 0" class="payout-table">
        <div class="payout-header">Payout Distribution</div>
        <div v-for="w in winnerDistribution" :key="w.id" class="payout-row">
          <span class="payout-name">{{ w.name }}</span>
          <div class="payout-right">
            <span class="payout-cash">{{ formatVND(w.amount) }}</span>
            <span class="payout-pts">−{{ w.pointsDeducted }} pts</span>
          </div>
        </div>
      </div>

      <!-- Warning -->
      <el-alert type="warning" :closable="false" show-icon class="mt-4">
        <template #title>
          <span style="font-size:13px">
            After settlement, <strong>{{ debtor.name }}</strong>'s score resets to 0 and all their matches are locked.
          </span>
        </template>
      </el-alert>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleCancel">Cancel</el-button>
        <el-button
          type="primary"
          size="large"
          @click="handleConfirm"
          :disabled="selectedWinners.length === 0"
          :loading="loading"
        >
          Confirm Settlement
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Warning } from '@element-plus/icons-vue'
import type { User } from '@/types/user'
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
  confirm: [winnerIds: string[]]
  cancel: []
}>()

const selectedWinners = ref<string[]>([])

const eligibleWinners = computed(() => {
  if (!props.debtor) return []
  return props.users.filter(u => u.id !== props.debtor!.id && u.current_score > 0 && u.is_active)
})

const debtPoints    = computed(() => props.debtor ? Math.abs(props.debtor.current_score) : 0)
const totalAmount   = computed(() => debtPoints.value * props.pointToVnd)
const fundAmount    = computed(() => Math.floor((totalAmount.value * props.fundSplitPercent) / 100))
const winnerAmount  = computed(() => totalAmount.value - fundAmount.value)

const winnerDistribution = computed(() => {
  if (!selectedWinners.value.length) return []
  const n = selectedWinners.value.length
  return selectedWinners.value.map(id => ({
    id,
    name: props.users.find(u => u.id === id)?.name || 'Unknown',
    amount: Math.floor(winnerAmount.value / n),
    pointsDeducted: Math.floor(debtPoints.value / n)
  }))
})

watch(() => props.modelValue, v => { if (!v) selectedWinners.value = [] })

const handleConfirm = () => emit('confirm', selectedWinners.value)
const handleCancel  = () => { emit('cancel'); emit('update:modelValue', false) }
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

.payout-table {
  border: 1px solid var(--border-default);
  border-radius: 10px; overflow: hidden; margin-bottom: 4px;
}
.payout-header {
  background: var(--surface-page);
  padding: 8px 14px;
  font-size: 11px; font-weight: 700; text-transform: uppercase;
  letter-spacing: 0.06em; color: var(--text-muted);
  border-bottom: 1px solid var(--border-default);
}
.payout-row {
  display: flex; align-items: center; justify-content: space-between;
  padding: 10px 14px;
  border-bottom: 1px solid var(--border-subtle);
}
.payout-row:last-child { border-bottom: none; }

.payout-name  { font-size: 13px; font-weight: 600; color: var(--text-primary); }
.payout-right { display: flex; flex-direction: column; align-items: flex-end; gap: 2px; }
.payout-cash  { font-size: 13px; font-weight: 700; color: var(--color-success); }
.payout-pts   { font-size: 11px; color: var(--color-danger); }

.dialog-footer { display: flex; justify-content: flex-end; gap: 10px; }
</style>
