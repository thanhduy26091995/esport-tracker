<template>
  <div class="wc-custom-bet-card">
    <div class="wc-cb-header">
      <span class="wc-cb-title">
        {{ bet.title }}
        <span v-if="bet.line != null" class="wc-cb-line">@{{ bet.line }}</span>
      </span>
      <el-tag :type="statusTagType" size="small" class="wc-cb-status">{{ statusLabel }}</el-tag>
    </div>

    <div class="wc-cb-options">
      <div
        v-for="opt in bet.options"
        :key="opt.id"
        class="wc-cb-option"
        :class="{
          'wc-cb-option--winner': opt.is_winner,
          'wc-cb-option--my-pick': bet.my_entry?.option_id === opt.id,
          'wc-cb-option--selectable': canBet && !bet.my_entry,
          'wc-cb-option--selected': selectedOptionId === opt.id,
        }"
        @click="canBet && !bet.my_entry && selectOption(opt.id)"
      >
        <span class="wc-cb-opt-label">
          <el-icon v-if="opt.is_winner" class="wc-cb-winner-icon"><CircleCheck /></el-icon>
          {{ opt.label }}
        </span>
        <span class="wc-cb-opt-odds">@{{ opt.odds.toFixed(2) }}</span>
      </div>
    </div>

    <!-- My entry info -->
    <div v-if="bet.my_entry" class="wc-cb-my-entry">
      <template v-if="bet.my_entry.status === 'pending'">
        <span class="wc-cb-entry-info">
          Đã cược <strong>{{ bet.my_entry.stake }}</strong> điểm
        </span>
        <el-button
          v-if="bet.status === 'open'"
          size="small"
          text
          type="danger"
          :loading="cancelling"
          @click="handleCancel"
        >Huỷ</el-button>
        <span v-else class="wc-cb-pending-label">Đang chờ kết quả</span>
      </template>
      <template v-else-if="bet.my_entry.status === 'won'">
        <span class="wc-cb-result wc-cb-result--win">
          Thắng +{{ ((bet.my_entry.payout ?? 0) - bet.my_entry.stake).toFixed(2) }} điểm
        </span>
      </template>
      <template v-else-if="bet.my_entry.status === 'lost'">
        <span class="wc-cb-result wc-cb-result--lose">
          Thua -{{ bet.my_entry.stake }} điểm
        </span>
      </template>
      <template v-else-if="bet.my_entry.status === 'void'">
        <span class="wc-cb-result wc-cb-result--void">Kèo đã huỷ — đã hoàn tiền</span>
      </template>
    </div>

    <!-- Place bet form -->
    <div v-if="canBet && !bet.my_entry && selectedOptionId" class="wc-cb-place-form">
      <el-input-number
        v-model="stake"
        :min="wcStore.minPoints"
        :max="wcStore.maxPoints"
        size="small"
        controls-position="right"
        style="width: 110px"
      />
      <el-button
        type="primary"
        size="small"
        :loading="placing"
        @click="handlePlace"
      >Đặt cược</el-button>
      <el-button size="small" text @click="selectedOptionId = null">Huỷ chọn</el-button>
    </div>

    <div v-if="bet.status === 'void' && !bet.my_entry" class="wc-cb-void-notice">
      Kèo đã huỷ
    </div>

  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { CircleCheck } from '@element-plus/icons-vue'
import { useWcStore } from '@/stores/wcStore'
import { wcService } from '@/services/wcService'
import type { WcCustomBetWithOptions } from '@/types/wc'


const props = defineProps<{ bet: WcCustomBetWithOptions }>()
const emit = defineEmits<{ (e: 'refresh'): void }>()

const wcStore = useWcStore()

const selectedOptionId = ref<string | null>(null)
const stake = ref(wcStore.minPoints)
const placing = ref(false)
const cancelling = ref(false)

const canBet = computed(() => props.bet.status === 'open')

const statusTagType = computed(() => {
  switch (props.bet.status) {
    case 'open': return 'success'
    case 'closed': return 'warning'
    case 'settled': return 'info'
    case 'void': return 'danger'
    default: return 'info'
  }
})

const statusLabel = computed(() => {
  switch (props.bet.status) {
    case 'open': return 'Đang mở'
    case 'closed': return 'Đã đóng'
    case 'settled': return 'Đã tất toán'
    case 'void': return 'Đã huỷ'
    default: return props.bet.status
  }
})

function selectOption(id: string) {
  selectedOptionId.value = selectedOptionId.value === id ? null : id
  stake.value = wcStore.minPoints
}

async function handlePlace() {
  if (!selectedOptionId.value) return
  placing.value = true
  try {
    await wcService.placeCustomBetEntry(props.bet.id, selectedOptionId.value, stake.value)
    selectedOptionId.value = null
    ElMessage.success('Đặt cược thành công!')
    emit('refresh')
  } catch {
    // error shown by wcApi interceptor
  } finally {
    placing.value = false
  }
}

async function handleCancel() {
  if (!props.bet.my_entry) return
  cancelling.value = true
  try {
    await wcService.cancelCustomBetEntry(props.bet.my_entry.id)
    ElMessage.success('Đã huỷ cược')
    emit('refresh')
  } catch {
    // error shown by wcApi interceptor
  } finally {
    cancelling.value = false
  }
}
</script>

<style scoped>
.wc-custom-bet-card {
  background: var(--surface-card);
  border: 1px solid var(--border-default);
  border-radius: 10px;
  padding: 14px 16px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.wc-cb-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.wc-cb-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  flex: 1;
  display: flex;
  align-items: center;
  gap: 6px;
}

.wc-cb-line {
  font-size: 12px;
  font-weight: 700;
  color: var(--el-color-warning);
  background: rgba(var(--el-color-warning-rgb, 230, 162, 60), 0.1);
  padding: 1px 6px;
  border-radius: 4px;
}

.wc-cb-options {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.wc-cb-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  border-radius: 8px;
  border: 1.5px solid var(--border-default);
  background: var(--surface-page);
  transition: border-color 0.15s, background 0.15s;
}

.wc-cb-option--selectable {
  cursor: pointer;
}

.wc-cb-option--selectable:hover {
  border-color: var(--el-color-primary-light-5);
}

.wc-cb-option--selected {
  border-color: var(--el-color-primary);
  background: rgba(var(--el-color-primary-rgb, 64, 158, 255), 0.06);
}

.wc-cb-option--my-pick {
  border-color: var(--el-color-primary-light-3);
}

.wc-cb-option--winner {
  border-color: #16a34a;
  background: rgba(22, 163, 74, 0.08);
}

.wc-cb-opt-label {
  font-size: 13px;
  color: var(--text-primary);
  display: flex;
  align-items: center;
  gap: 5px;
}

.wc-cb-winner-icon {
  color: #16a34a;
}

.wc-cb-opt-odds {
  font-size: 12px;
  font-weight: 700;
  color: var(--el-color-primary);
}

.wc-cb-my-entry {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}

.wc-cb-entry-info {
  color: var(--text-secondary);
}

.wc-cb-pending-label {
  font-size: 12px;
  color: var(--text-muted);
}

.wc-cb-result {
  font-size: 13px;
  font-weight: 700;
}

.wc-cb-result--win {
  color: #16a34a;
}

.wc-cb-result--lose {
  color: #ef4444;
}

.wc-cb-result--void {
  color: var(--text-muted);
}

.wc-cb-place-form {
  display: flex;
  align-items: center;
  gap: 8px;
}

.wc-cb-void-notice {
  font-size: 12px;
  color: var(--text-muted);
  text-align: center;
}

</style>
