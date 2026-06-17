<template>
  <el-dialog v-model="visible" title="Import kèo Tài/Xỉu từ StatsAPI" width="480px" @closed="reset">
    <div v-if="match" class="wc-import-match-name">
      {{ match.home_team }} vs {{ match.away_team }}
    </div>

    <div v-if="!preview" class="wc-import-intro">
      <p class="wc-import-desc">
        Lấy kèo Tài/Xỉu (Over/Under) từ TheStatsAPI và ghi đè lên dữ liệu hiện tại.
      </p>
      <div v-if="match && !match.statsapi_fixture_id" class="wc-import-warn">
        Trận này chưa có Fixture ID — hãy chạy Setup Mapping trước.
      </div>
    </div>

    <div v-if="preview" class="wc-import-preview">
      <div class="wc-import-compare">
        <div class="wc-import-compare-col">
          <div class="wc-import-compare-label">Hiện tại</div>
          <div class="wc-import-compare-row">
            <span>Mức Tài/Xỉu:</span>
            <span>{{ preview.current.ou_line ?? '—' }}</span>
          </div>
          <div class="wc-import-compare-row">
            <span>Kèo Tài:</span>
            <span>{{ preview.current.odds_over ?? '—' }}</span>
          </div>
          <div class="wc-import-compare-row">
            <span>Kèo Xỉu:</span>
            <span>{{ preview.current.odds_under ?? '—' }}</span>
          </div>
        </div>
        <div class="wc-import-compare-arrow">→</div>
        <div class="wc-import-compare-col wc-import-compare-col--new">
          <div class="wc-import-compare-label">Mới ({{ preview.source }})</div>
          <div class="wc-import-compare-row">
            <span>Mức Tài/Xỉu:</span>
            <strong>{{ preview.proposed.ou_line ?? '—' }}</strong>
          </div>
          <div class="wc-import-compare-row">
            <span>Kèo Tài:</span>
            <strong>{{ preview.proposed.odds_over ?? '—' }}</strong>
          </div>
          <div class="wc-import-compare-row">
            <span>Kèo Xỉu:</span>
            <strong>{{ preview.proposed.odds_under ?? '—' }}</strong>
          </div>
        </div>
      </div>
      <div class="wc-import-fetched">Fetched at: {{ preview.fetched_at }}</div>
    </div>

    <template #footer>
      <el-button @click="visible = false">Đóng</el-button>
      <el-button v-if="!preview" type="primary" :loading="loading" :disabled="!match?.statsapi_fixture_id" @click="runPreview">
        Xem trước
      </el-button>
      <el-button v-if="preview" type="warning" plain :loading="saving" @click="confirmImport">
        Ghi đè và lưu
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { wcService } from '@/services/wcService'
import type { WcMatch, ImportOUPreview } from '@/types/wc'

const emit = defineEmits<{ (e: 'imported'): void }>()

const visible = ref(false)
const loading = ref(false)
const saving = ref(false)
const match = ref<WcMatch | null>(null)
const preview = ref<ImportOUPreview | null>(null)

async function open(m: WcMatch) {
  match.value = m
  preview.value = null
  try {
    match.value = await wcService.getMatch(m.id)
  } catch { /* fall back to passed object */ }
  visible.value = true
}

function reset() {
  preview.value = null
}

async function runPreview() {
  if (!match.value) return
  loading.value = true
  try {
    const res = await wcService.importOU(match.value.id, true)
    preview.value = res as ImportOUPreview
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error ?? 'Không thể lấy dữ liệu')
  } finally {
    loading.value = false
  }
}

async function confirmImport() {
  if (!match.value) return
  saving.value = true
  try {
    await wcService.importOU(match.value.id, false)
    ElMessage.success('Đã import kèo Tài/Xỉu')
    visible.value = false
    emit('imported')
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error ?? 'Lỗi khi import')
  } finally {
    saving.value = false
  }
}

defineExpose({ open })
</script>

<style scoped>
.wc-import-match-name {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-primary);
  text-align: center;
  margin-bottom: 12px;
}

.wc-import-intro {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.wc-import-desc {
  font-size: 13px;
  color: var(--text-secondary);
  margin: 0;
}

.wc-import-warn {
  font-size: 13px;
  color: #d97706;
  background: rgba(217, 119, 6, 0.08);
  padding: 8px 12px;
  border-radius: 6px;
}

.wc-import-preview {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.wc-import-compare {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

.wc-import-compare-col {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
  background: var(--surface-page);
  border-radius: 8px;
  padding: 10px 12px;
}

.wc-import-compare-col--new {
  background: rgba(22, 163, 74, 0.06);
  border: 1px solid rgba(22, 163, 74, 0.2);
}

.wc-import-compare-label {
  font-size: 11px;
  font-weight: 700;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 2px;
}

.wc-import-compare-row {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  color: var(--text-secondary);
}

.wc-import-compare-row strong {
  color: var(--text-primary);
}

.wc-import-compare-arrow {
  font-size: 18px;
  color: var(--text-muted);
  padding-top: 28px;
}

.wc-import-fetched {
  font-size: 11px;
  color: var(--text-muted);
  text-align: right;
}
</style>
