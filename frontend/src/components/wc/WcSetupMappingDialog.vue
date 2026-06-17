<template>
  <el-dialog v-model="visible" title="Setup StatsAPI Mapping" width="680px" @closed="reset">
    <div v-if="!result" class="wc-mapping-intro">
      <p class="wc-mapping-desc">
        Tự động ghép các trận đấu trong hệ thống với fixture ID trên TheStatsAPI bằng tên đội + ngày thi đấu.
      </p>
      <el-button type="primary" :loading="loading" @click="runPreview">
        Xem trước mapping
      </el-button>
    </div>

    <div v-else>
      <div class="wc-mapping-summary">
        <span class="wc-mapping-chip wc-mapping-chip--ok">✓ Khớp: {{ result.matched.length }}</span>
        <span class="wc-mapping-chip wc-mapping-chip--warn">Chưa khớp (local): {{ result.unmatched_local.length }}</span>
        <span class="wc-mapping-chip wc-mapping-chip--info">API fixtures: {{ result.total_api_fixtures }}</span>
      </div>

      <el-collapse class="wc-mapping-collapse">
        <el-collapse-item :title="`Đã khớp (${result.matched.length})`" name="matched">
          <el-table :data="result.matched" size="small" max-height="240">
            <el-table-column prop="home_team" label="Home" width="130" />
            <el-table-column prop="away_team" label="Away" width="130" />
            <el-table-column prop="statsapi_fixture_id" label="Fixture ID" />
            <el-table-column prop="confidence" label="Độ tin cậy" width="100" />
          </el-table>
        </el-collapse-item>
        <el-collapse-item v-if="result.unmatched_local.length" :title="`Chưa khớp (${result.unmatched_local.length})`" name="unmatched">
          <div v-for="m in result.unmatched_local" :key="m.id" class="wc-mapping-unmatched-row">
            {{ m.home_team }} vs {{ m.away_team }}
          </div>
        </el-collapse-item>
      </el-collapse>
    </div>

    <template #footer>
      <el-button @click="visible = false">Đóng</el-button>
      <el-button v-if="result && result.matched.length" type="primary" :loading="saving" @click="confirmSave">
        Lưu {{ result.matched.length }} mapping
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { wcService } from '@/services/wcService'
import type { MappingResult } from '@/types/wc'

const emit = defineEmits<{ mapped: [] }>()

const visible = ref(false)
const loading = ref(false)
const saving = ref(false)
const result = ref<MappingResult | null>(null)

function open() {
  visible.value = true
}

function reset() {
  result.value = null
}

async function runPreview() {
  loading.value = true
  try {
    result.value = await wcService.setupMapping(true)
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error ?? 'Không thể lấy dữ liệu từ StatsAPI')
  } finally {
    loading.value = false
  }
}

async function confirmSave() {
  saving.value = true
  try {
    await wcService.setupMapping(false)
    ElMessage.success(`Đã lưu ${result.value?.matched.length ?? 0} mapping`)
    visible.value = false
    emit('mapped')
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error ?? 'Lỗi khi lưu mapping')
  } finally {
    saving.value = false
  }
}

defineExpose({ open })
</script>

<style scoped>
.wc-mapping-intro {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.wc-mapping-desc {
  font-size: 13px;
  color: var(--text-secondary);
  margin: 0;
}

.wc-mapping-summary {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 12px;
}

.wc-mapping-chip {
  font-size: 12px;
  font-weight: 600;
  padding: 3px 10px;
  border-radius: 12px;
}

.wc-mapping-chip--ok {
  background: rgba(22, 163, 74, 0.1);
  color: #16a34a;
}

.wc-mapping-chip--warn {
  background: rgba(217, 119, 6, 0.1);
  color: #d97706;
}

.wc-mapping-chip--info {
  background: rgba(59, 130, 246, 0.1);
  color: #3b82f6;
}

.wc-mapping-collapse {
  margin-top: 8px;
}

.wc-mapping-unmatched-row {
  font-size: 13px;
  padding: 4px 0;
  color: var(--text-secondary);
}
</style>
