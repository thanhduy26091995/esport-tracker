<template>
  <div v-if="show" class="gate-overlay">
    <div class="gate-card">
      <div class="gate-icon">🔒</div>
      <h2 class="gate-question">{{ store.question || 'Xác thực truy cập' }}</h2>
      <el-input
        v-model="answer"
        :placeholder="t('gate.placeholder')"
        size="large"
        class="gate-input"
        @keyup.enter="submit"
        :disabled="loading"
      />
      <p v-if="error" class="gate-error">{{ error }}</p>
      <el-button
        type="primary"
        size="large"
        class="gate-btn"
        :loading="loading"
        @click="submit"
      >
        {{ t('gate.submit') }}
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteAccessStore } from '@/stores/siteAccessStore'

const { t } = useI18n()
const store = useSiteAccessStore()

const answer = ref('')
const loading = ref(false)
const error = ref('')

const show = computed(() => store.checked && store.enabled && !store.token)

async function submit() {
  if (!answer.value.trim()) return
  loading.value = true
  error.value = ''
  try {
    await store.submit(answer.value.trim())
  } catch {
    error.value = t('gate.error')
    answer.value = ''
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.gate-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(6px);
}

.gate-card {
  background: #fff;
  border-radius: 16px;
  padding: 40px 36px;
  width: 100%;
  max-width: 420px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.35);
}

.gate-icon {
  font-size: 36px;
}

.gate-question {
  font-size: 18px;
  font-weight: 600;
  text-align: center;
  color: #1a1a2e;
  margin: 0;
}

.gate-input {
  width: 100%;
}

.gate-error {
  color: #f56c6c;
  font-size: 14px;
  margin: 0;
}

.gate-btn {
  width: 100%;
}
</style>
