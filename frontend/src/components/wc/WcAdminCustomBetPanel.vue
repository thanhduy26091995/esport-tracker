<template>
  <el-dialog
    v-model="visible"
    :title="`Kèo phụ — ${matchName}`"
    width="640px"
    destroy-on-close
    @open="loadBets"
  >
    <!-- Create form -->
    <div class="wc-cb-admin-section">
      <div class="wc-cb-admin-section-title">Tạo kèo mới</div>
      <el-input
        v-model="createForm.title"
        placeholder="Tiêu đề kèo (vd: Tài/Xỉu Góc H1)"
        maxlength="300"
        show-word-limit
        style="margin-bottom: 8px"
      />
      <div class="wc-cb-line-row">
        <span class="wc-cb-line-label">Line (tùy chọn)</span>
        <el-input-number
          v-model="createForm.line"
          :min="0"
          :step="0.5"
          :precision="1"
          :controls="false"
          placeholder="vd: 4.5"
          style="width: 100px"
        />
      </div>
      <div
        v-for="(opt, i) in createForm.options"
        :key="i"
        class="wc-cb-opt-row"
      >
        <el-input
          v-model="opt.label"
          placeholder="Lựa chọn"
          style="flex: 1"
        />
        <el-input-number
          v-model="opt.odds"
          :min="1.01"
          :step="0.1"
          :precision="2"
          controls-position="right"
          style="width: 110px"
          placeholder="Odds"
        />
        <el-button
          text
          type="danger"
          :disabled="createForm.options.length <= 2"
          @click="removeOption(i)"
        >✕</el-button>
      </div>
      <div class="wc-cb-opt-actions">
        <el-button
          text
          type="primary"
          :disabled="createForm.options.length >= 10"
          @click="addOption"
        >+ Thêm lựa chọn</el-button>
        <el-button
          type="primary"
          :loading="creating"
          :disabled="!createForm.title.trim()"
          @click="handleCreate"
        >Tạo kèo</el-button>
      </div>
    </div>

    <!-- Existing bets -->
    <div class="wc-cb-admin-section">
      <div class="wc-cb-admin-section-title">Danh sách kèo</div>
      <div v-if="loading" class="wc-cb-loading">Đang tải...</div>
      <div v-else-if="bets.length === 0" class="wc-cb-empty">Chưa có kèo phụ nào.</div>
      <div v-else class="wc-cb-list">
        <div v-for="bet in bets" :key="bet.id" class="wc-cb-admin-row">
          <div class="wc-cb-admin-row-header">
            <span class="wc-cb-admin-title">
              {{ bet.title }}
              <span v-if="bet.line != null" class="wc-cb-line-chip">@{{ bet.line }}</span>
            </span>
            <el-tag :type="statusTagType(bet.status)" size="small">{{ statusLabel(bet.status) }}</el-tag>
            <span class="wc-cb-entry-count">{{ bet.entry_count }} cược</span>
          </div>
          <div class="wc-cb-admin-opts">
            <span v-for="opt in bet.options" :key="opt.id" class="wc-cb-admin-opt-chip">
              {{ opt.label }} @{{ opt.odds.toFixed(2) }}
              <el-icon v-if="opt.is_winner" class="wc-winner-icon"><CircleCheck /></el-icon>
            </span>
          </div>
          <div class="wc-cb-admin-actions">
            <template v-if="bet.status === 'open'">
              <el-button plain size="small" type="warning" @click="handleCloseStatus(bet.id)">Đóng cược</el-button>
              <el-button plain size="small" type="success" @click="openSettle(bet)">Tất toán</el-button>
              <el-button size="small" type="danger" plain @click="handleVoid(bet.id)">Huỷ kèo</el-button>
            </template>
            <template v-else-if="bet.status === 'closed'">
              <el-button plain size="small" type="info" @click="handleOpenStatus(bet.id)">Mở lại</el-button>
              <el-button plain size="small" type="success" @click="openSettle(bet)">Tất toán</el-button>
              <el-button size="small" type="danger" plain @click="handleVoid(bet.id)">Huỷ kèo</el-button>
            </template>
            <template v-else>
              <span class="wc-cb-done-label">{{ statusLabel(bet.status) }}</span>
            </template>
          </div>
        </div>
      </div>
    </div>

    <!-- Settle dialog -->
    <el-dialog
      v-model="settleDialogVisible"
      title="Chọn kết quả thắng"
      width="400px"
      append-to-body
    >
      <div v-if="settlingBet" class="wc-settle-options">
        <div
          v-for="opt in settlingBet.options"
          :key="opt.id"
          class="wc-settle-option"
          :class="{ 'wc-settle-option--selected': selectedWinner === opt.id }"
          @click="selectedWinner = opt.id"
        >
          <span>{{ opt.label }}</span>
          <span class="wc-settle-odds">@{{ opt.odds.toFixed(2) }}</span>
        </div>
      </div>
      <template #footer>
        <el-button @click="settleDialogVisible = false">Hủy</el-button>
        <el-button
          plain
          type="success"
          :disabled="!selectedWinner"
          :loading="settling"
          @click="handleSettle"
        >Xác nhận tất toán</el-button>
      </template>
    </el-dialog>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { CircleCheck } from '@element-plus/icons-vue'
import { wcService } from '@/services/wcService'
import type { WcCustomBetWithOptions } from '@/types/wc'

const visible = ref(false)
const matchId = ref('')
const matchName = ref('')

const loading = ref(false)
const creating = ref(false)
const settling = ref(false)
const bets = ref<WcCustomBetWithOptions[]>([])

const createForm = reactive({
  title: '',
  line: null as number | null,
  options: [
    { label: '', odds: 1.9, display_order: 0 },
    { label: '', odds: 1.9, display_order: 1 },
  ],
})

const settleDialogVisible = ref(false)
const settlingBet = ref<WcCustomBetWithOptions | null>(null)
const selectedWinner = ref<string | null>(null)

function open(mid: string, name: string) {
  matchId.value = mid
  matchName.value = name
  visible.value = true
}

defineExpose({ open })

async function loadBets() {
  loading.value = true
  try {
    bets.value = await wcService.adminListCustomBets(matchId.value)
  } finally {
    loading.value = false
  }
}

function addOption() {
  createForm.options.push({ label: '', odds: 1.9, display_order: createForm.options.length })
}

function removeOption(i: number) {
  createForm.options.splice(i, 1)
}

async function handleCreate() {
  if (!createForm.title.trim()) return
  for (const opt of createForm.options) {
    if (!opt.label.trim()) {
      ElMessage.warning('Vui lòng điền đầy đủ tên lựa chọn')
      return
    }
    if (opt.odds <= 1) {
      ElMessage.warning('Odds phải > 1')
      return
    }
  }
  creating.value = true
  try {
    await wcService.adminCreateCustomBet(
      matchId.value,
      createForm.title.trim(),
      createForm.line,
      createForm.options.map((o, i) => ({ label: o.label.trim(), odds: o.odds, display_order: i })),
    )
    createForm.title = ''
    createForm.line = null
    createForm.options = [
      { label: '', odds: 1.9, display_order: 0 },
      { label: '', odds: 1.9, display_order: 1 },
    ]
    ElMessage.success('Tạo kèo thành công')
    await loadBets()
  } catch {
    // error shown by wcApi interceptor
  } finally {
    creating.value = false
  }
}

async function handleCloseStatus(betId: string) {
  try {
    await wcService.adminUpdateCustomBet(betId, { status: 'closed' })
    await loadBets()
  } catch {
    // error shown by wcApi interceptor
  }
}

async function handleOpenStatus(betId: string) {
  try {
    await wcService.adminUpdateCustomBet(betId, { status: 'open' })
    await loadBets()
  } catch {
    // error shown by wcApi interceptor
  }
}

function openSettle(bet: WcCustomBetWithOptions) {
  settlingBet.value = bet
  selectedWinner.value = null
  settleDialogVisible.value = true
}

async function handleSettle() {
  if (!settlingBet.value || !selectedWinner.value) return
  settling.value = true
  try {
    await wcService.adminSettleCustomBet(settlingBet.value.id, selectedWinner.value)
    settleDialogVisible.value = false
    ElMessage.success('Tất toán thành công')
    await loadBets()
  } catch {
    // error shown by wcApi interceptor
  } finally {
    settling.value = false
  }
}

async function handleVoid(betId: string) {
  await ElMessageBox.confirm('Huỷ kèo này sẽ hoàn tiền cho tất cả người đã cược. Xác nhận?', 'Huỷ kèo', {
    confirmButtonText: 'Huỷ kèo',
    cancelButtonText: 'Không',
    type: 'warning',
  })
  try {
    await wcService.adminVoidCustomBet(betId)
    ElMessage.success('Đã huỷ kèo')
    await loadBets()
  } catch {
    // error shown by wcApi interceptor
  }
}

function statusTagType(status: string) {
  switch (status) {
    case 'open': return 'success'
    case 'closed': return 'warning'
    case 'settled': return 'info'
    case 'void': return 'danger'
    default: return 'info'
  }
}

function statusLabel(status: string) {
  switch (status) {
    case 'open': return 'Đang mở'
    case 'closed': return 'Đã đóng'
    case 'settled': return 'Đã tất toán'
    case 'void': return 'Đã huỷ'
    default: return status
  }
}
</script>

<style scoped>
.wc-cb-admin-section {
  margin-bottom: 20px;
}

.wc-cb-admin-section + .wc-cb-admin-section {
  border-top: 1px solid var(--border-default);
  padding-top: 16px;
}

.wc-cb-admin-section-title {
  font-size: 13px;
  font-weight: 700;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 10px;
}

.wc-cb-line-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
}

.wc-cb-line-label {
  font-size: 13px;
  color: var(--text-secondary);
  white-space: nowrap;
}

.wc-cb-line-chip {
  font-size: 11px;
  font-weight: 700;
  color: var(--el-color-warning);
  background: rgba(var(--el-color-warning-rgb, 230, 162, 60), 0.1);
  padding: 1px 5px;
  border-radius: 4px;
  margin-left: 2px;
}

.wc-cb-opt-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.wc-cb-opt-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 8px;
}

.wc-cb-loading,
.wc-cb-empty {
  color: var(--text-muted);
  font-size: 13px;
  text-align: center;
  padding: 16px 0;
}

.wc-cb-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.wc-cb-admin-row {
  border: 1px solid var(--border-default);
  border-radius: 8px;
  padding: 12px 14px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.wc-cb-admin-row-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.wc-cb-admin-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  flex: 1;
}

.wc-cb-entry-count {
  font-size: 12px;
  color: var(--text-muted);
}

.wc-cb-admin-opts {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.wc-cb-admin-opt-chip {
  font-size: 12px;
  background: var(--surface-page);
  border: 1px solid var(--border-default);
  border-radius: 6px;
  padding: 2px 8px;
  color: var(--text-secondary);
  display: flex;
  align-items: center;
  gap: 4px;
}

.wc-winner-icon {
  color: #16a34a;
}

.wc-cb-admin-actions {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.wc-cb-done-label {
  font-size: 12px;
  color: var(--text-muted);
}

/* Settle dialog styles */
.wc-settle-options {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.wc-settle-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  border: 1.5px solid var(--border-default);
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  transition: border-color 0.15s, background 0.15s;
}

.wc-settle-option:hover {
  border-color: var(--el-color-primary-light-5);
}

.wc-settle-option--selected {
  border-color: var(--el-color-primary);
  background: rgba(var(--el-color-primary-rgb, 64, 158, 255), 0.08);
  font-weight: 600;
}

.wc-settle-odds {
  font-size: 12px;
  color: var(--el-color-primary);
  font-weight: 700;
}
</style>
