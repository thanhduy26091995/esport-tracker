<template>
  <div class="page-wrapper">
    <div class="link-container">
      <div class="link-card">
        <div class="link-icon">
          <svg width="40" height="40" viewBox="0 0 24 24" fill="none">
            <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"/>
            <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/>
            <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l3.66-2.84z" fill="#FBBC05"/>
            <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/>
          </svg>
        </div>

        <h1 class="link-title">Liên kết tài khoản Google</h1>
        <p class="link-desc">
          Hệ thống WC 2026 yêu cầu xác thực danh tính qua Google để ngăn chặn mạo danh.
          Vui lòng liên kết tài khoản Google của bạn để tiếp tục.
        </p>

        <div class="link-info">
          <div class="link-info-row">
            <span class="link-info-icon">✓</span>
            <span>Dữ liệu hiện tại (điểm, cược, ví) được giữ nguyên</span>
          </div>
          <div class="link-info-row">
            <span class="link-info-icon">✓</span>
            <span>Chỉ liên kết một lần duy nhất</span>
          </div>
          <div class="link-info-row">
            <span class="link-info-icon">✓</span>
            <span>Bạn vẫn có thể đăng nhập bằng mật khẩu sau khi liên kết</span>
          </div>
        </div>

        <div id="link-google-btn" class="link-google-btn-wrapper"></div>

        <p v-if="errorMsg" class="link-error">{{ errorMsg }}</p>

        <button class="link-logout" @click="handleLogout">
          Đăng xuất và dùng tài khoản khác
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useWcAuthStore } from '@/stores/wcAuthStore'

const router = useRouter()
const authStore = useWcAuthStore()
const errorMsg = ref('')

onMounted(() => {
  const clientId = import.meta.env.VITE_GOOGLE_CLIENT_ID
  if (!clientId || !window.google) return

  window.google.accounts.id.initialize({
    client_id: clientId,
    callback: handleGoogleCredential,
    auto_select: false,
  })
  window.google.accounts.id.renderButton(
    document.getElementById('link-google-btn')!,
    { theme: 'filled_blue', size: 'large', text: 'continue_with', locale: 'vi', width: 320 }
  )
})

async function handleGoogleCredential(response: { credential: string }) {
  errorMsg.value = ''
  const ok = await authStore.linkGoogle(response.credential)
  if (ok) {
    router.push('/world-cup/predict')
  } else {
    errorMsg.value = 'Tài khoản Google này đã được liên kết với người chơi khác. Vui lòng dùng tài khoản Google khác.'
  }
}

function handleLogout() {
  authStore.logout()
  router.push('/world-cup/login')
}
</script>

<style scoped>
.link-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: calc(100vh - 52px);
  padding: 24px 16px;
  background: linear-gradient(160deg, #0f2027 0%, #1a3a2a 50%, #0f2027 100%);
}

.link-card {
  width: 100%;
  max-width: 440px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 20px;
  padding: 40px 32px;
  backdrop-filter: blur(12px);
  box-shadow: 0 24px 64px rgba(0, 0, 0, 0.4);
  text-align: center;
}

.link-icon {
  width: 72px;
  height: 72px;
  background: rgba(255, 255, 255, 0.06);
  border-radius: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 20px;
}

.link-title {
  font-size: 22px;
  font-weight: 800;
  color: #ffffff;
  margin: 0 0 12px;
}

.link-desc {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.6);
  line-height: 1.6;
  margin: 0 0 24px;
}

.link-info {
  background: rgba(22, 163, 74, 0.08);
  border: 1px solid rgba(22, 163, 74, 0.2);
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 24px;
  text-align: left;
}

.link-info-row {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.7);
  margin-bottom: 8px;
}

.link-info-row:last-child {
  margin-bottom: 0;
}

.link-info-icon {
  color: #4ade80;
  font-weight: 700;
  flex-shrink: 0;
}

.link-google-btn-wrapper {
  display: flex;
  justify-content: center;
  margin-bottom: 16px;
}

.link-error {
  font-size: 12px;
  color: #f87171;
  margin: 0 0 16px;
  padding: 10px;
  background: rgba(248, 113, 113, 0.1);
  border-radius: 8px;
}

.link-logout {
  background: none;
  border: none;
  color: rgba(255, 255, 255, 0.3);
  font-size: 12px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  transition: color 0.15s;
}

.link-logout:hover {
  color: rgba(255, 255, 255, 0.55);
}
</style>
