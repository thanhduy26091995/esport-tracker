<template>
  <div class="wc-settlement-preview">
    <div class="wc-sp-controls">
      <div class="wc-sp-rate-row">
        <label class="wc-sp-label">{{ t('wc.pointRate') }}</label>
        <el-input-number
          v-model="pointRate"
          :min="1"
          :step="1000"
          controls-position="right"
          style="width: 180px"
          @change="handlePreview"
        />
        <span class="wc-sp-rate-hint">VND / điểm</span>
      </div>
    </div>

    <div v-if="previewRows.length > 0" class="wc-sp-table">
      <div class="wc-sp-thead">
        <span>Tên</span>
        <span>Số dư</span>
        <span>Hướng</span>
        <span>Số tiền</span>
      </div>
      <div
        v-for="row in previewRows"
        :key="row.wc_user_id"
        class="wc-sp-row"
      >
        <span class="wc-sp-name">{{ row.user_name }}</span>
        <span class="wc-sp-balance" :class="row.balance >= 0 ? 'text-green-600' : 'text-red-500'">
          {{ row.balance >= 0 ? '+' : '' }}{{ row.balance }}
        </span>
        <span class="wc-sp-dir" :class="`wc-dir--${row.direction}`">
          {{ dirLabel(row.direction) }}
        </span>
        <span class="wc-sp-amount">{{ formatMoney(row.amount) }}</span>
      </div>
    </div>

    <div class="wc-sp-actions">
      <el-button @click="handlePreview" :loading="previewing">{{ t('wc.previewSettlement') }}</el-button>
      <el-button type="success" @click="showCreateDialog = true" :disabled="previewRows.length === 0">
        {{ t('wc.createSettlement') }}
      </el-button>
    </div>

    <el-dialog v-model="showCreateDialog" :title="t('wc.createSettlement')" width="400px">
      <el-form :model="createForm" label-position="top">
        <el-form-item :label="t('wc.settlementName')">
          <el-input v-model="createForm.name" :placeholder="t('wc.settlementName')" />
        </el-form-item>
        <el-form-item :label="t('wc.settlementNote')">
          <el-input v-model="createForm.note" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item :label="t('wc.pointRate')">
          <el-input-number v-model="createForm.pointRate" :min="1" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="success" :loading="creating" @click="handleCreate">
          {{ t('wc.createSettlement') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useWcStore } from '@/stores/wcStore'
import type { WcSettlementPreviewRow, WcSettlementDirection } from '@/types/wc'

const { t } = useI18n()
const store = useWcStore()

const pointRate = ref(10000)
const previewRows = ref<WcSettlementPreviewRow[]>([])
const previewing = ref(false)
const creating = ref(false)
const showCreateDialog = ref(false)
const createForm = ref({ name: '', note: '', pointRate: 10000 })

watch(pointRate, () => {
  createForm.value.pointRate = pointRate.value
})

async function handlePreview() {
  previewing.value = true
  try {
    await store.previewSettlement(pointRate.value)
    previewRows.value = store.settlementPreview
  } finally {
    previewing.value = false
  }
}

async function handleCreate() {
  if (!createForm.value.name) return
  creating.value = true
  try {
    await store.createSettlement(createForm.value.name, createForm.value.pointRate, createForm.value.note)
    showCreateDialog.value = false
    previewRows.value = []
    createForm.value = { name: '', note: '', pointRate: pointRate.value }
  } finally {
    creating.value = false
  }
}

function dirLabel(d: WcSettlementDirection) {
  if (d === 'pay') return t('wc.directionPay')
  if (d === 'collect') return t('wc.directionCollect')
  return t('wc.directionEven')
}

function formatMoney(n: number) {
  return new Intl.NumberFormat('vi-VN').format(n) + ' ₫'
}
</script>

<style scoped>
.wc-settlement-preview {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.wc-sp-controls {
  background: var(--surface-page);
  border-radius: 10px;
  padding: 14px;
}

.wc-sp-rate-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.wc-sp-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
  white-space: nowrap;
}

.wc-sp-rate-hint {
  font-size: 12px;
  color: var(--text-muted);
}

.wc-sp-table {
  border: 1px solid var(--border-default);
  border-radius: 10px;
  overflow: hidden;
}

.wc-sp-thead {
  display: grid;
  grid-template-columns: 1fr 80px 80px 100px;
  gap: 8px;
  padding: 8px 14px;
  background: var(--surface-page);
  font-size: 11px;
  font-weight: 700;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.wc-sp-row {
  display: grid;
  grid-template-columns: 1fr 80px 80px 100px;
  gap: 8px;
  padding: 10px 14px;
  border-top: 1px solid var(--border-subtle);
  font-size: 13px;
  align-items: center;
}

.wc-sp-name {
  font-weight: 600;
  color: var(--text-primary);
}

.wc-sp-balance {
  font-weight: 600;
  tabular-nums: true;
}

.wc-sp-dir {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 6px;
  font-size: 11px;
  font-weight: 700;
  text-align: center;
}

.wc-dir--pay {
  background: rgba(22, 163, 74, 0.1);
  color: #16a34a;
}

.wc-dir--collect {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.wc-dir--even {
  background: rgba(100, 116, 139, 0.1);
  color: #64748b;
}

.wc-sp-amount {
  font-weight: 600;
  tabular-nums: true;
  text-align: right;
  color: var(--text-primary);
}

.wc-sp-actions {
  display: flex;
  gap: 8px;
}
</style>
