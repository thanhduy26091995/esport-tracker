<template>
  <div class="wc-settlement-history">
    <div v-if="settlements.length === 0" class="empty-state">
      <div class="empty-state-title">Chưa có tất toán nào</div>
    </div>
    <div v-else class="wc-sh-list">
      <div v-for="s in settlements" :key="s.id" class="wc-sh-card">
        <div class="wc-sh-card-header" @click="toggle(s.id)">
          <div class="wc-sh-info">
            <span class="wc-sh-name">{{ s.name }}</span>
            <span class="wc-sh-date">{{ formatDate(s.created_at) }}</span>
            <span v-if="s.note" class="wc-sh-note">{{ s.note }}</span>
          </div>
          <div class="wc-sh-rate">{{ s.point_rate?.toLocaleString() }} ₫/điểm</div>
          <el-icon class="wc-sh-chevron" :class="{ rotated: expanded === s.id }"><ArrowDown /></el-icon>
        </div>

        <div v-if="expanded === s.id" class="wc-sh-details">
          <div v-if="!detailMap[s.id]" class="wc-sh-loading">
            <el-icon class="is-loading"><Loading /></el-icon>
          </div>
          <template v-else>
            <div class="wc-sh-detail-list">
              <div
                v-for="detail in detailMap[s.id]"
                :key="detail.id"
                class="wc-sh-detail-row"
              >
                <span class="wc-shd-user">{{ detail.user_name }}</span>
                <span class="wc-shd-dir" :class="`wc-dir--${detail.direction}`">
                  {{ dirLabel(detail.direction) }}
                </span>
                <span class="wc-shd-amount">{{ formatMoney(detail.amount) }}</span>
                <span class="wc-shd-status" :class="`wc-status--${detail.status}`">
                  {{ detail.status === 'done' ? t('wc.statusDone') : t('wc.statusPending') }}
                </span>
                <el-button
                  v-if="detail.status !== 'done' && detail.direction !== 'even'"
                  size="small"
                  type="success"
                  text
                  @click="handleMarkDone(s.id, detail.wc_user_id)"
                >
                  {{ t('wc.markDone') }}
                </el-button>
              </div>
            </div>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ArrowDown, Loading } from '@element-plus/icons-vue'
import { useWcStore } from '@/stores/wcStore'
import type { WcSettlement, WcSettlementDetailWithUser, WcSettlementDirection } from '@/types/wc'

const { t } = useI18n()
const store = useWcStore()
defineProps<{ settlements: WcSettlement[] }>()

const expanded = ref<string | null>(null)
const detailMap = ref<Record<string, WcSettlementDetailWithUser[]>>({})

async function toggle(id: string) {
  if (expanded.value === id) {
    expanded.value = null
    return
  }
  expanded.value = id
  if (!detailMap.value[id]) {
    await store.fetchSettlement(id)
    if (store.currentSettlement?.id === id) {
      detailMap.value[id] = store.currentSettlement.details
    }
  }
}

async function handleMarkDone(settlementId: string, wcUserId: string) {
  await store.markSettlementDone(settlementId, wcUserId)
  await store.fetchSettlement(settlementId)
  if (store.currentSettlement) {
    detailMap.value[settlementId] = store.currentSettlement.details
  }
}

function dirLabel(d: WcSettlementDirection) {
  if (d === 'pay') return t('wc.directionPay')
  if (d === 'collect') return t('wc.directionCollect')
  return t('wc.directionEven')
}

function formatDate(s: string) {
  return new Date(s).toLocaleDateString('vi-VN', { day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit' })
}

function formatMoney(n: number) {
  return new Intl.NumberFormat('vi-VN').format(n) + ' ₫'
}
</script>

<style scoped>
.wc-sh-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.wc-sh-card {
  border: 1px solid var(--border-default);
  border-radius: 12px;
  overflow: hidden;
  background: var(--surface-card);
}

.wc-sh-card-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  cursor: pointer;
  transition: background 0.15s;
}

.wc-sh-card-header:hover {
  background: var(--surface-page);
}

.wc-sh-info {
  flex: 1;
}

.wc-sh-name {
  display: block;
  font-size: 14px;
  font-weight: 700;
  color: var(--text-primary);
}

.wc-sh-date {
  font-size: 11px;
  color: var(--text-muted);
}

.wc-sh-note {
  display: block;
  font-size: 11px;
  color: var(--text-muted);
  font-style: italic;
}

.wc-sh-rate {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  flex-shrink: 0;
}

.wc-sh-chevron {
  color: var(--text-muted);
  transition: transform 0.2s;
  flex-shrink: 0;
}

.wc-sh-chevron.rotated {
  transform: rotate(180deg);
}

.wc-sh-details {
  border-top: 1px solid var(--border-subtle);
  background: var(--surface-page);
  padding: 8px;
}

.wc-sh-loading {
  display: flex;
  justify-content: center;
  padding: 16px;
  color: var(--text-muted);
}

.wc-sh-detail-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.wc-sh-detail-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  background: var(--surface-card);
  border-radius: 8px;
  font-size: 13px;
}

.wc-shd-user {
  flex: 1;
  font-weight: 600;
  color: var(--text-primary);
}

.wc-shd-dir {
  padding: 2px 8px;
  border-radius: 6px;
  font-size: 11px;
  font-weight: 700;
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

.wc-shd-amount {
  font-weight: 600;
  tabular-nums: true;
  color: var(--text-primary);
}

.wc-shd-status {
  font-size: 11px;
  font-weight: 700;
  padding: 1px 7px;
  border-radius: 5px;
}

.wc-status--pending {
  background: rgba(217, 119, 6, 0.1);
  color: #d97706;
}

.wc-status--done {
  background: rgba(22, 163, 74, 0.1);
  color: #16a34a;
}
</style>
