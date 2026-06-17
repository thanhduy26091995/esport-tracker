<template>
  <el-dialog v-model="visible" title="Tạo kèo tỉ số (Poisson)" width="600px" @closed="reset">
    <div v-if="match" class="wc-poisson-match-name">
      {{ match.home_team }} vs {{ match.away_team }}
    </div>

    <el-form :model="form" label-position="top" size="small" class="wc-poisson-form">
      <div class="wc-poisson-params-row">
        <el-form-item label="λ Home (tỉ lệ ghi bàn nhà)">
          <el-input-number v-model="form.home_lambda" :min="0.1" :max="6" :step="0.1" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="λ Away (tỉ lệ ghi bàn khách)">
          <el-input-number v-model="form.away_lambda" :min="0.1" :max="6" :step="0.1" :precision="2" style="width: 100%" />
        </el-form-item>
      </div>
      <div class="wc-poisson-params-row">
        <el-form-item label="House Margin (vd: 0.10 = 10%)">
          <el-input-number v-model="form.house_margin" :min="0" :max="0.5" :step="0.01" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="Min Probability">
          <el-input-number v-model="form.min_prob" :min="0.001" :max="0.1" :step="0.005" :precision="3" style="width: 100%" />
        </el-form-item>
      </div>
      <div class="wc-poisson-actions">
        <el-button type="primary" :loading="loading" @click="runPreview">Xem trước</el-button>
      </div>
    </el-form>

    <div v-if="preview" class="wc-poisson-preview">
      <div class="wc-poisson-preview-header">
        <span>{{ preview.count }} tỉ số — Margin: {{ (preview.house_margin * 100).toFixed(0) }}%</span>
      </div>
      <el-table :data="preview.score_odds" size="small" max-height="280" :default-sort="{ prop: 'probability', order: 'descending' }">
        <el-table-column label="Tỉ số" width="80">
          <template #default="{ row }">
            <strong>{{ row.home_score }}–{{ row.away_score }}</strong>
          </template>
        </el-table-column>
        <el-table-column prop="probability" label="Xác suất" width="100" sortable>
          <template #default="{ row }">{{ (row.probability * 100).toFixed(2) }}%</template>
        </el-table-column>
        <el-table-column prop="odds" label="Kèo" width="80" sortable />
      </el-table>
    </div>

    <template #footer>
      <el-button @click="visible = false">Đóng</el-button>
      <el-button v-if="preview" type="warning" plain :loading="saving" @click="confirmSave">
        Lưu {{ preview.count }} kèo tỉ số
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { wcService } from '@/services/wcService'
import type { WcMatch, GeneratePoissonPreview } from '@/types/wc'

const emit = defineEmits<{ (e: 'saved'): void }>()

const visible = ref(false)
const loading = ref(false)
const saving = ref(false)
const match = ref<WcMatch | null>(null)
const preview = ref<GeneratePoissonPreview | null>(null)

const form = ref({
  home_lambda: 1.5,
  away_lambda: 1.2,
  house_margin: 0.1,
  min_prob: 0.01,
})

function open(m: WcMatch) {
  match.value = m
  preview.value = null
  visible.value = true
}

function reset() {
  preview.value = null
}

async function runPreview() {
  if (!match.value) return
  loading.value = true
  try {
    const res = await wcService.generatePoisson(match.value.id, {
      home_lambda: form.value.home_lambda,
      away_lambda: form.value.away_lambda,
      house_margin: form.value.house_margin,
      min_prob: form.value.min_prob,
    }, true)
    preview.value = res as GeneratePoissonPreview
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error ?? 'Lỗi tạo Poisson')
  } finally {
    loading.value = false
  }
}

async function confirmSave() {
  if (!match.value) return
  saving.value = true
  try {
    await wcService.generatePoisson(match.value.id, {
      home_lambda: form.value.home_lambda,
      away_lambda: form.value.away_lambda,
      house_margin: form.value.house_margin,
      min_prob: form.value.min_prob,
    }, false)
    ElMessage.success(`Đã lưu ${preview.value?.count ?? 0} kèo tỉ số`)
    visible.value = false
    emit('saved')
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error ?? 'Lỗi khi lưu')
  } finally {
    saving.value = false
  }
}

defineExpose({ open })
</script>

<style scoped>
.wc-poisson-match-name {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-primary);
  text-align: center;
  margin-bottom: 12px;
}

.wc-poisson-form {
  margin-bottom: 0;
}

.wc-poisson-params-row {
  display: flex;
  gap: 16px;
}

.wc-poisson-params-row .el-form-item {
  flex: 1;
}

.wc-poisson-actions {
  margin-top: 4px;
}

.wc-poisson-preview {
  margin-top: 16px;
}

.wc-poisson-preview-header {
  font-size: 12px;
  color: var(--text-muted);
  margin-bottom: 8px;
}
</style>
