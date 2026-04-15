<template>
  <el-dialog
    :model-value="modelValue"
    title="Settlement Details"
    @update:model-value="$emit('update:modelValue', $event)"
    width="90%"
    style="max-width: 680px"
    destroy-on-close
  >
    <div v-if="settlement">
      <!-- Header -->
      <div class="sd-header">
        <div class="sd-header-icon">
          <el-icon :size="22" color="white"><Warning /></el-icon>
        </div>
        <div class="sd-header-info">
          <p class="sd-header-name">{{ settlement.debtor.name }}</p>
          <p class="sd-header-date">{{ formatDateTime(settlement.settlement_date) }}</p>
        </div>
        <div class="sd-header-total">{{ formatVND(settlement.money_amount) }}</div>
      </div>

      <!-- Summary row -->
      <div class="sd-summary">
        <div class="sd-stat">
          <dt>Original Debt</dt>
          <dd class="sd-stat-val sd-stat-val--red">{{ settlement.original_debt_points }} pts</dd>
        </div>
        <div class="sd-stat-divider" />
        <div class="sd-stat">
          <dt>Total Paid</dt>
          <dd class="sd-stat-val">{{ formatVND(settlement.money_amount) }}</dd>
        </div>
        <div class="sd-stat-divider" />
        <div class="sd-stat">
          <dt>To Fund</dt>
          <dd class="sd-stat-val sd-stat-val--green">{{ formatVND(settlement.fund_amount) }}</dd>
        </div>
        <div class="sd-stat-divider" />
        <div class="sd-stat">
          <dt>To Winners</dt>
          <dd class="sd-stat-val sd-stat-val--blue">{{ formatVND(settlement.winner_distribution) }}</dd>
        </div>
      </div>

      <!-- Winners table -->
      <div v-if="settlement.winners?.length" class="sd-section">
        <p class="sd-section-title">Winner Payouts ({{ settlement.winners.length }})</p>
        <div class="sd-winners">
          <div class="sd-winner-header">
            <span>Player</span>
            <span>Cash Received</span>
            <span>Points Deducted</span>
          </div>
          <div v-for="w in settlement.winners" :key="w.id" class="sd-winner-row">
            <div class="sd-winner-name">
              <div class="sd-winner-avatar">{{ w.winner.name.charAt(0).toUpperCase() }}</div>
              {{ w.winner.name }}
            </div>
            <span class="sd-winner-cash">+{{ formatVND(w.money_amount) }}</span>
            <span class="sd-winner-pts">−{{ w.points_deducted }}</span>
          </div>
        </div>
      </div>

      <!-- Settlement process -->
      <div class="sd-process">
        <p class="sd-process-title">
          <el-icon style="margin-right:6px;vertical-align:middle"><InfoFilled /></el-icon>
          Settlement Process
        </p>
        <ol class="sd-process-list">
          <li>{{ settlement.debtor.name }} had {{ settlement.original_debt_points }} points debt (below threshold)</li>
          <li>{{ settlement.debtor.name }} paid {{ formatVND(settlement.money_amount) }} in real cash</li>
          <li>{{ fundSplitPercent }}% ({{ formatVND(settlement.fund_amount) }}) added to fund</li>
          <li>{{ 100 - fundSplitPercent }}% ({{ formatVND(settlement.winner_distribution) }}) split among {{ settlement.winners?.length ?? 0 }} winner(s)</li>
          <li>Winners received cash and had points deducted proportionally</li>
          <li>{{ settlement.debtor.name }}'s score reset to 0</li>
        </ol>
      </div>
    </div>

    <template #footer>
      <el-button size="large" @click="$emit('update:modelValue', false)">Close</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { Warning, InfoFilled } from '@element-plus/icons-vue'
import type { DebtSettlement } from '@/types/settlement'
import { formatVND } from '@/utils/formatters'
import { formatDateTime } from '@/utils/date'

interface Props {
  modelValue: boolean
  settlement: DebtSettlement | null
  fundSplitPercent?: number
}

withDefaults(defineProps<Props>(), { fundSplitPercent: 50 })
defineEmits<{ 'update:modelValue': [value: boolean] }>()
</script>

<style scoped>
.sd-header {
  display: flex; align-items: center; gap: 12px;
  padding: 16px; border-radius: 14px;
  background: var(--color-danger-bg);
  border: 1px solid var(--color-danger-border);
  margin-bottom: 16px;
}

.sd-header-icon {
  width: 44px; height: 44px; border-radius: 12px; flex-shrink: 0;
  background: linear-gradient(135deg, #ef4444, #b91c1c);
  display: flex; align-items: center; justify-content: center;
  box-shadow: 0 4px 10px rgba(239,68,68,0.3);
}

.sd-header-info { flex: 1; min-width: 0; }
.sd-header-name  { font-size: 16px; font-weight: 700; color: var(--color-danger); }
.sd-header-date  { font-size: 11px; color: var(--text-muted); margin-top: 2px; }
.sd-header-total { font-size: 18px; font-weight: 800; color: var(--color-danger); flex-shrink: 0; }

/* Summary strip */
.sd-summary {
  display: flex; align-items: center;
  background: var(--surface-page); border-radius: 12px;
  padding: 14px 16px; margin-bottom: 16px; gap: 0;
}
.sd-stat { flex: 1; text-align: center; }
.sd-stat dt { font-size: 10px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.06em; color: var(--text-muted); margin-bottom: 4px; }
.sd-stat-val { font-size: 14px; font-weight: 700; color: var(--text-primary); }
.sd-stat-val--red   { color: var(--color-danger); }
.sd-stat-val--green { color: var(--color-success); }
.sd-stat-val--blue  { color: var(--color-info); }

.sd-stat-divider { width: 1px; height: 32px; background: var(--border-default); flex-shrink: 0; }

/* Winners */
.sd-section { margin-bottom: 16px; }
.sd-section-title { font-size: 12px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.06em; color: var(--text-muted); margin-bottom: 8px; }

.sd-winners { border: 1px solid var(--border-default); border-radius: 10px; overflow: hidden; }

.sd-winner-header {
  display: grid; grid-template-columns: 1fr 1fr 1fr;
  background: var(--surface-page); padding: 8px 14px;
  font-size: 10px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.06em;
  color: var(--text-muted); border-bottom: 1px solid var(--border-default);
}

.sd-winner-row {
  display: grid; grid-template-columns: 1fr 1fr 1fr;
  align-items: center; padding: 10px 14px;
  border-bottom: 1px solid var(--border-subtle);
}
.sd-winner-row:last-child { border-bottom: none; }

.sd-winner-name {
  display: flex; align-items: center; gap: 8px;
  font-size: 13px; font-weight: 600; color: var(--text-primary);
}
.sd-winner-avatar {
  width: 26px; height: 26px; border-radius: 50%; flex-shrink: 0;
  background: linear-gradient(135deg, #dbeafe, #bfdbfe);
  color: #1d4ed8; font-size: 11px; font-weight: 700;
  display: flex; align-items: center; justify-content: center;
}
.sd-winner-cash { font-size: 13px; font-weight: 700; color: var(--color-success); text-align: center; }
.sd-winner-pts  { font-size: 12px; font-weight: 600; color: var(--color-danger); text-align: center; }

/* Process */
.sd-process {
  background: var(--color-info-bg); border: 1px solid var(--color-info-border);
  border-radius: 12px; padding: 14px 16px;
}
.sd-process-title { font-size: 12px; font-weight: 700; color: var(--color-info); margin-bottom: 8px; }
.sd-process-list {
  padding-left: 18px; margin: 0;
  font-size: 12px; color: var(--text-secondary); line-height: 1.7;
}
</style>
