<template>
  <div class="wc-pnl">
    <div class="wc-pnl-header">
      <span class="wc-pnl-title">House P&amp;L</span>
      <el-button size="small" text :loading="loading" @click="load">
        <el-icon><Refresh /></el-icon>
      </el-button>
    </div>

    <div v-if="loading && !data" class="wc-pnl-loading">
      <el-icon class="is-loading"><Loading /></el-icon>
    </div>

    <template v-else-if="data">
      <!-- Summary chips -->
      <div class="wc-pnl-summary">
        <div class="wc-pnl-chip">
          <span class="wc-pnl-chip-label">Stake thu</span>
          <span class="wc-pnl-chip-val">{{ fmt(data.total_stake_settled) }}</span>
        </div>
        <div class="wc-pnl-chip">
          <span class="wc-pnl-chip-label">Payout trả</span>
          <span class="wc-pnl-chip-val">{{ fmt(data.total_payout_settled) }}</span>
        </div>
        <div class="wc-pnl-chip" :class="data.house_profit >= 0 ? 'wc-pnl-chip--green' : 'wc-pnl-chip--red'">
          <span class="wc-pnl-chip-label">{{ data.house_profit >= 0 ? 'Lời' : 'Lỗ' }}</span>
          <span class="wc-pnl-chip-val wc-pnl-profit">
            {{ data.house_profit >= 0 ? '+' : '' }}{{ fmt(data.house_profit) }}
          </span>
        </div>
      </div>

      <!-- Pending row -->
      <div v-if="data.pending_bet_count > 0" class="wc-pnl-pending">
        <el-icon><Warning /></el-icon>
        <span>{{ data.pending_bet_count }} cược chờ settle — Stake: {{ fmt(data.total_stake_pending) }}</span>
      </div>

      <!-- Void row -->
      <div v-if="data.total_stake_void > 0" class="wc-pnl-void">
        <span>Void (hoàn cược): {{ fmt(data.total_stake_void) }}</span>
      </div>

      <!-- Match breakdown -->
      <div v-if="data.match_breakdown.length > 0" class="wc-pnl-breakdown">
        <div class="wc-pnl-breakdown-header" @click="showBreakdown = !showBreakdown">
          <span>Chi tiết theo trận ({{ data.match_breakdown.length }} trận)</span>
          <el-icon :class="{ rotated: showBreakdown }"><ArrowDown /></el-icon>
        </div>
        <div v-if="showBreakdown" class="wc-pnl-table">
          <div class="wc-pnl-thead">
            <span>Trận</span>
            <span>Cược</span>
            <span>Stake</span>
            <span>Payout</span>
            <span>Lời/Lỗ</span>
          </div>
          <div v-for="m in data.match_breakdown" :key="m.match_id" class="wc-pnl-trow">
            <span class="wc-pnl-td-match">{{ m.home_team }} vs {{ m.away_team }}</span>
            <span>{{ m.bet_count }}</span>
            <span>{{ fmt(m.stake) }}</span>
            <span>{{ fmt(m.payout) }}</span>
            <span :class="m.profit >= 0 ? 'wc-profit--pos' : 'wc-profit--neg'">
              {{ m.profit >= 0 ? '+' : '' }}{{ fmt(m.profit) }}
            </span>
          </div>
        </div>
      </div>

      <div class="wc-pnl-footer">
        Cập nhật lúc {{ fmtTime(data.generated_at) }} · {{ data.settled_bet_count }} cược đã settle
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Loading, Refresh, ArrowDown, Warning } from '@element-plus/icons-vue'
import { wcService } from '@/services/wcService'
import type { HousePnLResponse } from '@/types/wc'

const loading = ref(false)
const data = ref<HousePnLResponse | null>(null)
const showBreakdown = ref(false)

async function load() {
  loading.value = true
  try {
    data.value = await wcService.getHousePnL()
  } finally {
    loading.value = false
  }
}

onMounted(load)
defineExpose({ load })

function fmt(n: number) {
  return new Intl.NumberFormat('vi-VN').format(n)
}

function fmtTime(iso: string) {
  return new Date(iso).toLocaleTimeString('vi-VN', { hour: '2-digit', minute: '2-digit' })
}
</script>

<style scoped>
.wc-pnl {
  background: var(--surface-card);
  border: 1px solid var(--border-default);
  border-radius: 14px;
  padding: 14px 16px;
  margin-bottom: 16px;
}

.wc-pnl-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.wc-pnl-title {
  font-size: 13px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--text-muted);
}

.wc-pnl-loading {
  display: flex;
  justify-content: center;
  padding: 20px;
  color: var(--text-muted);
}

.wc-pnl-summary {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 10px;
}

.wc-pnl-chip {
  flex: 1;
  min-width: 100px;
  background: var(--surface-page);
  border: 1px solid var(--border-default);
  border-radius: 10px;
  padding: 10px 12px;
}

.wc-pnl-chip--green {
  background: rgba(22, 163, 74, 0.08);
  border-color: rgba(22, 163, 74, 0.2);
}

.wc-pnl-chip--red {
  background: rgba(239, 68, 68, 0.08);
  border-color: rgba(239, 68, 68, 0.2);
}

.wc-pnl-chip-label {
  display: block;
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
  margin-bottom: 4px;
}

.wc-pnl-chip-val {
  display: block;
  font-size: 16px;
  font-weight: 800;
  color: var(--text-primary);
  font-variant-numeric: tabular-nums;
}

.wc-pnl-chip--green .wc-pnl-profit { color: #16a34a; }
.wc-pnl-chip--red .wc-pnl-profit { color: #ef4444; }

.wc-pnl-pending {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #d97706;
  background: rgba(217, 119, 6, 0.08);
  border: 1px solid rgba(217, 119, 6, 0.2);
  border-radius: 8px;
  padding: 6px 10px;
  margin-bottom: 8px;
}

.wc-pnl-void {
  font-size: 11px;
  color: var(--text-muted);
  margin-bottom: 8px;
}

.wc-pnl-breakdown-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 6px 0;
  border-top: 1px solid var(--border-subtle);
}

.wc-pnl-breakdown-header .el-icon {
  transition: transform 0.2s;
}

.wc-pnl-breakdown-header .el-icon.rotated {
  transform: rotate(180deg);
}

.wc-pnl-table {
  margin-top: 6px;
  border: 1px solid var(--border-default);
  border-radius: 8px;
  overflow: hidden;
}

.wc-pnl-thead {
  display: grid;
  grid-template-columns: 2fr 50px 1fr 1fr 1fr;
  gap: 6px;
  padding: 6px 10px;
  background: var(--surface-page);
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}

.wc-pnl-trow {
  display: grid;
  grid-template-columns: 2fr 50px 1fr 1fr 1fr;
  gap: 6px;
  padding: 7px 10px;
  font-size: 12px;
  border-top: 1px solid var(--border-subtle);
  align-items: center;
}

.wc-pnl-td-match {
  font-weight: 600;
  color: var(--text-primary);
  font-size: 11px;
}

.wc-profit--pos { color: #16a34a; font-weight: 700; }
.wc-profit--neg { color: #ef4444; font-weight: 700; }

.wc-pnl-footer {
  font-size: 10px;
  color: var(--text-muted);
  text-align: right;
  margin-top: 8px;
  padding-top: 6px;
  border-top: 1px solid var(--border-subtle);
}
</style>
