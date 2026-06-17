<template>
  <div class="wc-sync-logs">
    <div class="wc-sync-logs-header">
      <span class="wc-sync-logs-title">Lịch sử Sync</span>
      <el-button size="small" :loading="loading" @click="load">Refresh</el-button>
    </div>
    <div v-if="logs.length === 0 && !loading" class="wc-sync-logs-empty">
      Chưa có log nào.
    </div>
    <el-table v-else :data="logs" size="small" max-height="260">
      <el-table-column prop="created_at" label="Thời gian" width="140">
        <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column prop="trigger" label="Trigger" width="80" />
      <el-table-column prop="sync_type" label="Loại" width="120" />
      <el-table-column prop="matches_updated" label="Cập nhật" width="80" />
      <el-table-column prop="matches_failed" label="Lỗi" width="60">
        <template #default="{ row }">
          <span :class="row.matches_failed > 0 ? 'wc-sync-failed' : ''">{{ row.matches_failed }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="error_detail" label="Chi tiết lỗi" min-width="120">
        <template #default="{ row }">
          <span class="wc-sync-error-text">{{ row.error_detail ?? '—' }}</span>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { wcService } from '@/services/wcService'
import type { WcSyncLog } from '@/types/wc'

const logs = ref<WcSyncLog[]>([])
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    logs.value = await wcService.getSyncLogs()
  } finally {
    loading.value = false
  }
}

function formatTime(s: string) {
  return new Date(s).toLocaleString('vi-VN', {
    day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit',
  })
}

onMounted(load)
defineExpose({ load })
</script>

<style scoped>
.wc-sync-logs {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.wc-sync-logs-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.wc-sync-logs-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
}

.wc-sync-logs-empty {
  font-size: 13px;
  color: var(--text-muted);
  text-align: center;
  padding: 12px 0;
}

.wc-sync-failed {
  color: #ef4444;
  font-weight: 700;
}

.wc-sync-error-text {
  font-size: 11px;
  color: var(--text-muted);
  word-break: break-all;
}
</style>
