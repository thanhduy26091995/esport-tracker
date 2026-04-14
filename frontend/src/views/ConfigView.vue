<template>
  <div class="min-h-screen bg-gray-50">
    <div class="max-w-4xl mx-auto py-6 sm:px-6 lg:px-8">
      <div class="px-4 py-6 sm:px-0">
        <!-- Header -->
        <div class="mb-6">
          <h1 class="text-3xl font-bold text-gray-900">Configuration</h1>
          <p class="mt-1 text-sm text-gray-500">
            Manage system settings and conversion rates
          </p>
        </div>

        <!-- Settings Card -->
        <div class="bg-white shadow rounded-lg">
          <div class="px-4 py-5 sm:p-6">
            <div v-if="configStore.loading" class="text-center py-12">
              <el-icon class="animate-spin" :size="40"><Loading /></el-icon>
              <p class="mt-4 text-gray-500">Loading configuration...</p>
            </div>

            <form v-else @submit.prevent="handleSubmit" class="space-y-8">
              <!-- Debt Threshold -->
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">
                  Debt Threshold (Points)
                  <span class="text-red-500">*</span>
                </label>
                <el-input
                  v-model.number="formData.debt_threshold"
                  type="number"
                  placeholder="e.g., -50"
                  :max="0"
                  size="large"
                  class="w-full"
                >
                  <template #prefix>
                    <el-icon><Warning /></el-icon>
                  </template>
                </el-input>
                <p class="mt-2 text-sm text-gray-500">
                  When a player's score reaches this threshold (≤ 0), automatic debt settlement is triggered.
                </p>
                <p v-if="formData.debt_threshold > 0" class="mt-1 text-sm text-red-600">
                  ⚠️ Value must be less than or equal to 0
                </p>
              </div>

              <!-- Point to VND Conversion -->
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">
                  Point to VND Conversion Rate
                  <span class="text-red-500">*</span>
                </label>
                <el-input
                  v-model.number="formData.point_to_vnd"
                  type="number"
                  placeholder="e.g., 22000"
                  :min="1"
                  size="large"
                  class="w-full"
                >
                  <template #prefix>
                    <el-icon><Money /></el-icon>
                  </template>
                  <template #suffix>
                    <span class="text-gray-500">VND / point</span>
                  </template>
                </el-input>
                <p class="mt-2 text-sm text-gray-500">
                  How much VND each point is worth. Used to calculate debt amounts.
                </p>
                <div v-if="formData.point_to_vnd > 0" class="mt-2 p-3 bg-blue-50 rounded-lg">
                  <p class="text-sm text-blue-800">
                    <strong>Example:</strong> -10 points = {{ formatVND(formData.point_to_vnd * -10) }} debt
                  </p>
                </div>
                <p v-if="formData.point_to_vnd <= 0" class="mt-1 text-sm text-red-600">
                  ⚠️ Value must be greater than 0
                </p>
              </div>

              <!-- Fund Split Percentage -->
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">
                  Fund Split Percentage
                  <span class="text-red-500">*</span>
                </label>
                <div class="flex items-center gap-4">
                  <el-slider
                    v-model="formData.fund_split_percent"
                    :min="0"
                    :max="100"
                    :step="1"
                    show-stops
                    :marks="marks"
                    class="flex-1"
                  />
                  <el-input
                    v-model.number="formData.fund_split_percent"
                    type="number"
                    :min="0"
                    :max="100"
                    size="large"
                    class="w-24"
                  >
                    <template #suffix>%</template>
                  </el-input>
                </div>
                <p class="mt-2 text-sm text-gray-500">
                  Percentage of debt settlement that goes to the fund. The rest is distributed to winners.
                </p>
                <div v-if="formData.fund_split_percent >= 0 && formData.fund_split_percent <= 100" 
                     class="mt-2 p-3 bg-green-50 rounded-lg">
                  <p class="text-sm text-green-800">
                    <strong>Example ({{ formatVND(220000) }} debt):</strong><br>
                    Fund receives: {{ formatVND(220000 * formData.fund_split_percent / 100) }} ({{ formData.fund_split_percent }}%)<br>
                    Winners share: {{ formatVND(220000 * (100 - formData.fund_split_percent) / 100) }} ({{ 100 - formData.fund_split_percent }}%)
                  </p>
                </div>
                <p v-if="formData.fund_split_percent < 0 || formData.fund_split_percent > 100" 
                   class="mt-1 text-sm text-red-600">
                  ⚠️ Value must be between 0 and 100
                </p>
              </div>

              <!-- Current Values Info -->
              <div class="border-t border-gray-200 pt-6">
                <h3 class="text-sm font-medium text-gray-900 mb-3">Current Configuration</h3>
                <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
                  <div class="bg-gray-50 rounded-lg p-4">
                    <p class="text-xs text-gray-500 uppercase tracking-wide">Debt Threshold</p>
                    <p class="mt-1 text-lg font-semibold text-gray-900">
                      {{ configStore.debtThreshold }} points
                    </p>
                  </div>
                  <div class="bg-gray-50 rounded-lg p-4">
                    <p class="text-xs text-gray-500 uppercase tracking-wide">Point Value</p>
                    <p class="mt-1 text-lg font-semibold text-gray-900">
                      {{ formatVND(configStore.pointToVnd) }}
                    </p>
                  </div>
                  <div class="bg-gray-50 rounded-lg p-4">
                    <p class="text-xs text-gray-500 uppercase tracking-wide">Fund Split</p>
                    <p class="mt-1 text-lg font-semibold text-gray-900">
                      {{ configStore.fundSplitPercent }}%
                    </p>
                  </div>
                </div>
              </div>

              <!-- Action Buttons -->
              <div class="flex justify-end gap-3 pt-6 border-t border-gray-200">
                <el-button @click="handleReset" :disabled="!hasChanges">
                  Reset
                </el-button>
                <el-button
                  type="primary"
                  native-type="submit"
                  :loading="configStore.loading"
                  :disabled="!isFormValid || !hasChanges"
                >
                  Save Changes
                </el-button>
              </div>
            </form>
          </div>
        </div>

        <!-- Warning Alert -->
        <div class="mt-6">
          <el-alert
            title="Important Notice"
            type="warning"
            :closable="false"
            show-icon
          >
            <p class="text-sm">
              Changes to these settings will affect all future matches and settlements. 
              Existing records will not be recalculated.
            </p>
          </el-alert>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Loading, Warning, Money } from '@element-plus/icons-vue'
import { useConfigStore } from '@/stores/configStore'
import { formatVND } from '@/utils/formatters'

const configStore = useConfigStore()

// Form data
const formData = ref({
  debt_threshold: 0,
  point_to_vnd: 0,
  fund_split_percent: 0
})

// Slider marks
const marks = {
  0: '0%',
  25: '25%',
  50: '50%',
  75: '75%',
  100: '100%'
}

// Load config on mount
onMounted(async () => {
  await configStore.fetchConfigs()
  resetFormData()
})

// Watch for config changes
watch(() => [configStore.debtThreshold, configStore.pointToVnd, configStore.fundSplitPercent], () => {
  resetFormData()
}, { deep: true })

// Reset form data to current config
const resetFormData = () => {
  formData.value = {
    debt_threshold: configStore.debtThreshold,
    point_to_vnd: configStore.pointToVnd,
    fund_split_percent: configStore.fundSplitPercent
  }
}

// Check if form is valid
const isFormValid = computed(() => {
  return (
    formData.value.debt_threshold <= 0 &&
    formData.value.point_to_vnd > 0 &&
    formData.value.fund_split_percent >= 0 &&
    formData.value.fund_split_percent <= 100
  )
})

// Check if form has changes
const hasChanges = computed(() => {
  return (
    formData.value.debt_threshold !== configStore.debtThreshold ||
    formData.value.point_to_vnd !== configStore.pointToVnd ||
    formData.value.fund_split_percent !== configStore.fundSplitPercent
  )
})

// Event handlers
const handleReset = () => {
  resetFormData()
  ElMessage.info('Form reset to current values')
}

const handleSubmit = async () => {
  if (!isFormValid.value) {
    ElMessage.error('Please fix validation errors before saving')
    return
  }

  try {
    // Update each config value
    await configStore.updateConfig('debt_threshold', formData.value.debt_threshold.toString())
    await configStore.updateConfig('point_to_vnd', formData.value.point_to_vnd.toString())
    await configStore.updateConfig('fund_split_percent', formData.value.fund_split_percent.toString())
    ElMessage.success('Configuration updated successfully')
  } catch (error) {
    // Error is already handled by the store
  }
}
</script>
