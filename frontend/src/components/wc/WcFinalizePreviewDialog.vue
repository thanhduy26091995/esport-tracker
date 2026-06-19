<template>
  <el-dialog
    :model-value="modelValue"
    :title="title"
    width="680px"
    @update:model-value="$emit('update:modelValue', $event)"
    @close="$emit('cancel')"
  >
    <!-- Loading skeleton -->
    <div v-if="loading" class="wc-fp-skeleton">
      <el-skeleton :rows="5" animated />
    </div>

    <!-- No predictions -->
    <div v-else-if="preview && preview.matches.length === 0" class="wc-fp-empty">
      Không có dự đoán nào cần tính.
    </div>

    <!-- Preview content -->
    <div v-else-if="preview" class="wc-fp-content">
      <!-- Bulk: collapse per match -->
      <el-collapse v-if="preview.matches.length > 1" class="wc-fp-collapse">
        <el-collapse-item
          v-for="match in preview.matches"
          :key="match.match_id"
          :name="match.match_id"
        >
          <template #title>
            <span class="wc-fp-match-title">
              {{ match.home_team }} {{ match.home_score }}–{{ match.away_score }} {{ match.away_team }}
              <span class="wc-fp-stage">· {{ match.stage }}</span>
              <el-tag v-if="match.already_settled" size="small" type="warning" class="wc-fp-settled-tag">đã tính</el-tag>
            </span>
          </template>
          <div v-if="match.predictions.length === 0" class="wc-fp-no-pred">Không có dự đoán nào.</div>
          <table v-else class="wc-fp-table">
            <thead>
              <tr>
                <th>Tên</th>
                <th>Loại kèo</th>
                <th>Kết quả</th>
                <th class="wc-fp-th-r">Δ</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(row, i) in match.predictions" :key="i">
                <td class="wc-fp-td-name">{{ row.user_name }}</td>
                <td class="wc-fp-td-type">{{ predTypeLabel(row.prediction_type) }} ×{{ row.multiplier }}  {{ row.points }}đ</td>
                <td>{{ resultLabel(row.new_result) }}</td>
                <td class="wc-fp-td-delta" :class="row.net_delta >= 0 ? 'wc-fp-pos' : 'wc-fp-neg'">
                  {{ row.net_delta >= 0 ? '+' : '' }}{{ fmtPts(row.net_delta) }}
                </td>
              </tr>
            </tbody>
          </table>
        </el-collapse-item>
      </el-collapse>

      <!-- Single match: flat -->
      <div v-else class="wc-fp-single">
        <div class="wc-fp-match-header">
          {{ preview.matches[0].home_team }} {{ preview.matches[0].home_score }}–{{ preview.matches[0].away_score }} {{ preview.matches[0].away_team }}
          <span class="wc-fp-stage">· {{ preview.matches[0].stage }}</span>
        </div>
        <div v-if="preview.matches[0].predictions.length === 0" class="wc-fp-no-pred">Không có dự đoán nào.</div>
        <table v-else class="wc-fp-table">
          <thead>
            <tr>
              <th>Tên</th>
              <th>Loại kèo</th>
              <th>Kết quả</th>
              <th class="wc-fp-th-r">Δ</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(row, i) in preview.matches[0].predictions" :key="i">
              <td class="wc-fp-td-name">{{ row.user_name }}</td>
              <td class="wc-fp-td-type">{{ predTypeLabel(row.prediction_type) }} ×{{ row.multiplier }}  {{ row.points }}đ</td>
              <td>{{ resultLabel(row.new_result) }}</td>
              <td class="wc-fp-td-delta" :class="row.net_delta >= 0 ? 'wc-fp-pos' : 'wc-fp-neg'">
                {{ row.net_delta >= 0 ? '+' : '' }}{{ fmtPts(row.net_delta) }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- House summary -->
      <div class="wc-fp-summary">
        <span class="wc-fp-summary-item">
          Thu vào: <strong>{{ fmtPts(preview.house_summary.total_staked) }}</strong> đ
        </span>
        <span class="wc-fp-summary-sep">·</span>
        <span class="wc-fp-summary-item">
          Trả ra: <strong>{{ fmtPts(preview.house_summary.total_paid_out) }}</strong> đ
        </span>
        <span class="wc-fp-summary-sep">·</span>
        <span
          class="wc-fp-summary-item wc-fp-summary-net"
          :class="preview.house_summary.house_net >= 0 ? 'wc-fp-pos' : 'wc-fp-neg'"
        >
          {{ preview.house_summary.house_net >= 0 ? 'Lời' : 'Lỗ' }}:
          <strong>{{ preview.house_summary.house_net >= 0 ? '+' : '' }}{{ fmtPts(preview.house_summary.house_net) }}</strong> đ
        </span>
      </div>
    </div>

    <template #footer>
      <el-button @click="$emit('cancel')">Hủy</el-button>
      <el-button
        type="primary"
        :loading="confirming"
        :disabled="loading || !preview || preview.matches.length === 0"
        @click="$emit('confirm')"
      >
        Xác nhận &amp; Tính điểm
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import type { FinalizePreviewResult } from '@/types/wc'

defineProps<{
  modelValue: boolean
  title: string
  preview: FinalizePreviewResult | null
  loading: boolean
  confirming: boolean
}>()

defineEmits<{
  'update:modelValue': [value: boolean]
  confirm: []
  cancel: []
}>()

function fmtPts(v: number): string {
  return parseFloat(v.toFixed(2)).toString()
}

function resultLabel(r: string): string {
  switch (r) {
    case 'correct':   return '✅ Đúng'
    case 'incorrect': return '❌ Sai'
    case 'win_half':  return '⬆️ Thắng nửa'
    case 'lose_half': return '⬇️ Thua nửa'
    case 'void':      return '➡️ Hủy'
    default:          return r
  }
}

function predTypeLabel(t: string): string {
  switch (t) {
    case 'handicap':    return 'Chấp'
    case 'exact_score': return 'Tỉ số'
    case 'over_under':  return 'Tài/Xỉu'
    default:            return t
  }
}
</script>

<style scoped>
.wc-fp-skeleton {
  padding: 12px 0;
}

.wc-fp-empty,
.wc-fp-no-pred {
  text-align: center;
  color: var(--text-muted);
  font-size: 13px;
  padding: 16px 0;
}

.wc-fp-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.wc-fp-collapse {
  border: none;
}

.wc-fp-match-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 700;
}

.wc-fp-settled-tag {
  margin-left: 4px;
}

.wc-fp-single {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.wc-fp-match-header {
  font-size: 14px;
  font-weight: 700;
  color: var(--text-primary);
  text-align: center;
  padding: 4px 0 4px;
}

.wc-fp-stage {
  font-size: 12px;
  font-weight: 400;
  color: var(--text-muted);
}

.wc-fp-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.wc-fp-table th {
  text-align: left;
  font-size: 11px;
  font-weight: 700;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  padding: 6px 8px;
  border-bottom: 1px solid var(--border-subtle);
}

.wc-fp-th-r {
  text-align: right;
}

.wc-fp-table td {
  padding: 7px 8px;
  border-bottom: 1px solid var(--border-subtle);
  color: var(--text-primary);
}

.wc-fp-table tr:last-child td {
  border-bottom: none;
}

.wc-fp-td-name {
  font-weight: 600;
  min-width: 80px;
}

.wc-fp-td-type {
  color: var(--text-secondary);
  font-size: 12px;
}

.wc-fp-td-delta {
  font-weight: 700;
  text-align: right;
  min-width: 60px;
}

.wc-fp-summary {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  background: var(--surface-page);
  border-radius: 8px;
  font-size: 13px;
  flex-wrap: wrap;
}

.wc-fp-summary-item {
  color: var(--text-secondary);
}

.wc-fp-summary-sep {
  color: var(--text-muted);
}

.wc-fp-summary-net {
  font-weight: 700;
}

.wc-fp-pos {
  color: #16a34a;
}

.wc-fp-neg {
  color: #ef4444;
}
</style>
