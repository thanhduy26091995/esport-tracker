<template>
  <div class="page-wrapper">
    <div class="wc-auth-container">
      <div class="wc-auth-card">
        <div class="wc-auth-header">
          <div class="wc-trophy-badge">
            <el-icon :size="28" color="#ffffff"><Trophy /></el-icon>
          </div>
          <h1 class="wc-auth-title">{{ t('wc.registerTitle') }}</h1>
          <p class="wc-auth-subtitle">{{ t('wc.registerSubtitle') }}</p>
        </div>

        <el-form :model="form" @submit.prevent="handleRegister" class="wc-auth-form">
          <el-form-item>
            <el-input
              v-model="form.name"
              :placeholder="t('wc.namePlaceholder')"
              size="large"
              :prefix-icon="User"
              clearable
            />
          </el-form-item>
          <el-form-item>
            <el-input
              v-model="form.password"
              type="password"
              :placeholder="t('wc.passwordPlaceholder')"
              size="large"
              :prefix-icon="Lock"
              show-password
            />
          </el-form-item>
          <el-form-item>
            <el-input
              v-model="form.confirmPassword"
              type="password"
              :placeholder="t('wc.confirmPasswordPlaceholder')"
              size="large"
              :prefix-icon="Lock"
              show-password
              @keyup.enter="handleRegister"
            />
          </el-form-item>
          <p v-if="mismatch" class="wc-error-msg">{{ t('wc.passwordMismatch') }}</p>

          <el-button
            type="primary"
            size="large"
            :loading="authStore.loading"
            @click="handleRegister"
            class="wc-auth-btn"
          >
            {{ t('wc.registerBtn') }}
          </el-button>
        </el-form>

        <div class="wc-auth-footer">
          <span class="wc-auth-footer-text">{{ t('wc.hasAccount') }}</span>
          <router-link to="/world-cup/login" class="wc-auth-link">{{ t('wc.loginBtn') }}</router-link>
        </div>
      </div>

      <div class="wc-auth-back">
        <router-link to="/world-cup" class="wc-back-link">
          <el-icon><ArrowLeft /></el-icon>
          {{ t('wc.schedule') }}
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Trophy, User, Lock, ArrowLeft } from '@element-plus/icons-vue'
import { useWcAuthStore } from '@/stores/wcAuthStore'

const { t } = useI18n()
const router = useRouter()
const authStore = useWcAuthStore()

const form = ref({ name: '', password: '', confirmPassword: '' })
const mismatch = computed(
  () => form.value.confirmPassword.length > 0 && form.value.password !== form.value.confirmPassword,
)

async function handleRegister() {
  if (!form.value.name || !form.value.password) return
  if (form.value.password !== form.value.confirmPassword) return
  try {
    await authStore.register(form.value.name, form.value.password)
    router.push('/world-cup/predict')
  } catch { /* error shown by store */ }
}
</script>

<style scoped>
.wc-auth-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: calc(100vh - 52px);
  padding: 24px 16px;
  background: linear-gradient(160deg, #0f2027 0%, #1a3a2a 50%, #0f2027 100%);
}

.wc-auth-card {
  width: 100%;
  max-width: 420px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 20px;
  padding: 36px 32px;
  backdrop-filter: blur(12px);
  box-shadow: 0 24px 64px rgba(0, 0, 0, 0.4);
}

.wc-auth-header {
  text-align: center;
  margin-bottom: 28px;
}

.wc-trophy-badge {
  width: 64px;
  height: 64px;
  background: linear-gradient(135deg, #d97706, #f59e0b);
  border-radius: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 16px;
  box-shadow: 0 8px 24px rgba(217, 119, 6, 0.4);
}

.wc-auth-title {
  font-size: 22px;
  font-weight: 800;
  color: #ffffff;
  letter-spacing: -0.02em;
  margin: 0 0 6px;
}

.wc-auth-subtitle {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.5);
  margin: 0;
}

.wc-auth-form {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.wc-auth-form :deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.06) !important;
  box-shadow: 0 0 0 1px rgba(255, 255, 255, 0.12) !important;
}

.wc-auth-form :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px rgba(255, 255, 255, 0.24) !important;
}

.wc-auth-form :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 2px #16a34a !important;
}

.wc-auth-form :deep(.el-input__inner) {
  color: #ffffff !important;
}

.wc-auth-form :deep(.el-input__prefix-icon) {
  color: rgba(255, 255, 255, 0.4) !important;
}

.wc-error-msg {
  font-size: 12px;
  color: #f87171;
  margin: -8px 0 4px 2px;
}

.wc-auth-btn {
  width: 100%;
  background: linear-gradient(135deg, #16a34a, #15803d) !important;
  border-color: #15803d !important;
  height: 44px;
  font-size: 15px;
  font-weight: 600;
  margin-top: 8px;
  box-shadow: 0 4px 16px rgba(22, 163, 74, 0.35) !important;
}

.wc-auth-btn:hover {
  background: linear-gradient(135deg, #15803d, #166534) !important;
}

.wc-auth-footer {
  text-align: center;
  margin-top: 20px;
  font-size: 13px;
}

.wc-auth-footer-text {
  color: rgba(255, 255, 255, 0.5);
  margin-right: 6px;
}

.wc-auth-link {
  color: #4ade80;
  text-decoration: none;
  font-weight: 600;
}

.wc-auth-link:hover {
  color: #86efac;
  text-decoration: underline;
}

.wc-auth-back {
  margin-top: 20px;
}

.wc-back-link {
  display: flex;
  align-items: center;
  gap: 6px;
  color: rgba(255, 255, 255, 0.4);
  text-decoration: none;
  font-size: 13px;
  transition: color 0.15s;
}

.wc-back-link:hover {
  color: rgba(255, 255, 255, 0.7);
}
</style>
