<template>
  <div class="page-wrapper">
    <div class="wc-auth-container">
      <div class="wc-auth-card">
        <div class="wc-auth-header">
          <div class="wc-trophy-badge">
            <el-icon :size="28" color="#ffffff"><Trophy /></el-icon>
          </div>
          <h1 class="wc-auth-title">{{ tournamentName }}</h1>
          <p class="wc-auth-subtitle">{{ t('wc.loginSubtitle') }}</p>
        </div>

        <el-form :model="form" @submit.prevent="handleLogin" class="wc-auth-form">
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
              @keyup.enter="handleLogin"
            />
          </el-form-item>

          <el-button
            type="primary"
            size="large"
            :loading="authStore.loading"
            @click="handleLogin"
            class="wc-auth-btn"
            native-type="submit"
          >
            {{ t('wc.loginBtn') }}
          </el-button>
        </el-form>

        <div class="wc-divider">
          <span>hoặc</span>
        </div>

        <!-- Google Sign-In button rendered by GSI -->
        <div id="google-signin-btn" class="wc-google-btn-wrapper"></div>

        <p class="wc-google-hint">
          Người chơi mới? Đăng nhập bằng Google để tạo tài khoản tự động.
        </p>
      </div>

      <div class="wc-auth-back">
        <router-link :to="scheduleRoute" class="wc-back-link">
          <el-icon><ArrowLeft /></el-icon>
          {{ t('wc.schedule') }}
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Trophy, User, Lock, ArrowLeft } from '@element-plus/icons-vue'
import { useWcAuthStore } from '@/stores/wcAuthStore'
import { useTournamentRoutes } from '@/composables/useTournamentRoutes'

const { t } = useI18n()
const router = useRouter()
const authStore = useWcAuthStore()
const { scheduleRoute, predictPath, tournamentName } = useTournamentRoutes()

const form = ref({ name: '', password: '' })

async function handleLogin() {
  if (!form.value.name || !form.value.password) return
  try {
    await authStore.login(form.value.name, form.value.password)
    router.push(predictPath.value)
  } catch { /* error shown by store/api interceptor */ }
}

onMounted(() => {
  if (localStorage.getItem('wc_token')) {
    router.replace(predictPath.value)
    return
  }

  const clientId = import.meta.env.VITE_GOOGLE_CLIENT_ID
  if (!clientId) return

  const initGsi = () => {
    window.google!.accounts.id.initialize({
      client_id: clientId,
      callback: handleGoogleCredential,
      auto_select: false,
    })
    window.google!.accounts.id.renderButton(
      document.getElementById('google-signin-btn')!,
      { theme: 'outline', size: 'large', text: 'signin_with', locale: 'vi', width: 356 }
    )
  }

  if (window.google) {
    initGsi()
  } else {
    window.onGoogleLibraryLoad = initGsi
  }
})

async function handleGoogleCredential(response: { credential: string }) {
  try {
    await authStore.loginWithGoogle(response.credential)
    router.push(predictPath.value)
  } catch {
    ElMessage.error('Đăng nhập Google thất bại. Vui lòng thử lại.')
  }
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

.wc-divider {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 20px 0;
  color: rgba(255, 255, 255, 0.3);
  font-size: 12px;
}

.wc-divider::before,
.wc-divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background: rgba(255, 255, 255, 0.12);
}

.wc-google-btn-wrapper {
  display: flex;
  justify-content: center;
}

.wc-google-hint {
  text-align: center;
  font-size: 11px;
  color: rgba(255, 255, 255, 0.35);
  margin: 12px 0 0;
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
