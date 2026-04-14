import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import type { Config, ConfigKey } from '@/types/config'
import { configService } from '@/services/configService'

export const useConfigStore = defineStore('config', () => {
  const configs = ref<Config[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Getters
  const debtThreshold = computed(() => {
    const config = configs.value.find((c) => c.key === 'debt_threshold')
    return config ? parseInt(config.value) : -6
  })

  const pointToVnd = computed(() => {
    const config = configs.value.find((c) => c.key === 'point_to_vnd')
    return config ? parseInt(config.value) : 22000
  })

  const fundSplitPercent = computed(() => {
    const config = configs.value.find((c) => c.key === 'fund_split_percent')
    return config ? parseInt(config.value) : 50
  })

  // Actions
  async function fetchConfigs() {
    loading.value = true
    error.value = null
    try {
      configs.value = await configService.getAll()
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch configs'
      if (error.value) ElMessage.error(error.value)
    } finally {
      loading.value = false
    }
  }

  async function updateConfig(key: ConfigKey, value: string) {
    loading.value = true
    error.value = null
    try {
      const updated = await configService.update(key, { value })
      const index = configs.value.findIndex((c) => c.key === key)
      if (index !== -1) {
        configs.value[index] = updated
      }
      ElMessage.success(`Configuration "${key}" updated successfully`)
      return updated
    } catch (err: any) {
      const errorMsg =
        err.response?.data?.error?.message || err.message || 'Failed to update config'
      error.value = errorMsg
      ElMessage.error(errorMsg)
      throw err
    } finally {
      loading.value = false
    }
  }

  return {
    configs,
    loading,
    error,
    debtThreshold,
    pointToVnd,
    fundSplitPercent,
    fetchConfigs,
    updateConfig,
  }
})
