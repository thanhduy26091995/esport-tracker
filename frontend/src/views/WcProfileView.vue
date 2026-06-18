<template>
  <div class="page-wrapper">
    <div class="profile-container">
      <div class="profile-card">
        <div class="profile-header">
          <h1 class="profile-title">Hồ sơ cá nhân</h1>
          <p class="profile-subtitle">Chỉnh sửa tên hiển thị và ảnh đại diện</p>
        </div>

        <div v-if="loading" class="profile-loading">
          <el-skeleton :rows="4" animated />
        </div>

        <template v-else>
          <!-- Avatar preview -->
          <div class="avatar-section">
            <img
              :src="previewAvatar || DEFAULT_AVATAR"
              class="avatar-img"
              :alt="form.name"
              @error="(e: Event) => ((e.target as HTMLImageElement).src = DEFAULT_AVATAR)"
            />
            <div class="avatar-badge">
              <el-icon color="#16a34a"><Picture /></el-icon>
            </div>
          </div>

          <el-form :model="form" class="profile-form" label-position="top">
            <el-form-item label="Tên hiển thị">
              <el-input
                v-model="form.name"
                placeholder="Nhập tên hiển thị (tối thiểu 2 ký tự)"
                size="large"
                clearable
              />
            </el-form-item>

            <el-form-item label="URL ảnh đại diện">
              <el-input
                v-model="form.avatarUrl"
                placeholder="https://..."
                size="large"
                clearable
                @input="debouncedPreview"
              />
              <div class="avatar-hint">Dán URL ảnh bất kỳ. Ảnh Google của bạn được hỗ trợ.</div>
            </el-form-item>

            <el-button
              type="primary"
              size="large"
              :loading="saving"
              @click="handleSave"
              class="save-btn"
            >
              Lưu thay đổi
            </el-button>
          </el-form>

          <div class="profile-meta">
            <span class="profile-meta-item">
              <el-icon><Check /></el-icon>
              Google đã liên kết
            </span>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Picture, Check } from '@element-plus/icons-vue'
import { wcProfileService } from '@/services/wcProfileService'
import { useWcAuthStore } from '@/stores/wcAuthStore'

const DEFAULT_AVATAR = 'data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" width="80" height="80" viewBox="0 0 80 80"%3E%3Ccircle cx="40" cy="40" r="40" fill="%23374151"/%3E%3Ccircle cx="40" cy="32" r="14" fill="%236b7280"/%3E%3Cellipse cx="40" cy="72" rx="22" ry="16" fill="%236b7280"/%3E%3C/svg%3E'

const authStore = useWcAuthStore()
const loading = ref(true)
const saving = ref(false)

const form = ref({ name: '', avatarUrl: '' })
let previewTimer: ReturnType<typeof setTimeout> | null = null
const previewAvatar = ref<string>('')

onMounted(async () => {
  try {
    const profile = await wcProfileService.getProfile()
    form.value.name = profile.name
    form.value.avatarUrl = profile.avatar_url ?? ''
    previewAvatar.value = profile.avatar_url ?? ''
  } finally {
    loading.value = false
  }
})

function debouncedPreview() {
  if (previewTimer) clearTimeout(previewTimer)
  previewTimer = setTimeout(() => {
    previewAvatar.value = form.value.avatarUrl
  }, 600)
}

async function handleSave() {
  const name = form.value.name.trim()
  if (name.length < 2) {
    ElMessage.error('Tên phải có ít nhất 2 ký tự')
    return
  }
  saving.value = true
  try {
    const updated = await wcProfileService.updateProfile({
      name,
      avatar_url: form.value.avatarUrl || undefined,
    })
    // Sync store so navbar updates immediately
    if (authStore.user) {
      authStore.user = {
        ...authStore.user,
        name: updated.name,
        avatarUrl: updated.avatar_url,
      }
      localStorage.setItem('wc_user', JSON.stringify(authStore.user))
    }
    ElMessage.success('Đã lưu hồ sơ!')
  } catch (err: unknown) {
    const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error
    if (msg?.includes('taken')) {
      ElMessage.error('Tên này đã được sử dụng bởi người chơi khác.')
    } else {
      ElMessage.error(msg ?? 'Lưu thất bại. Vui lòng thử lại.')
    }
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.profile-container {
  display: flex;
  align-items: flex-start;
  justify-content: center;
  min-height: calc(100vh - 52px);
  padding: 32px 16px;
  background: linear-gradient(160deg, #0f2027 0%, #1a3a2a 50%, #0f2027 100%);
}

.profile-card {
  width: 100%;
  max-width: 480px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 20px;
  padding: 36px 32px;
  backdrop-filter: blur(12px);
  box-shadow: 0 24px 64px rgba(0, 0, 0, 0.4);
}

.profile-header {
  margin-bottom: 28px;
}

.profile-title {
  font-size: 22px;
  font-weight: 800;
  color: #ffffff;
  margin: 0 0 6px;
}

.profile-subtitle {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.5);
  margin: 0;
}

.avatar-section {
  display: flex;
  justify-content: center;
  position: relative;
  margin-bottom: 28px;
}

.avatar-img {
  width: 88px;
  height: 88px;
  border-radius: 50%;
  object-fit: cover;
  border: 3px solid rgba(22, 163, 74, 0.5);
  background: #374151;
}

.avatar-badge {
  position: absolute;
  bottom: 0;
  right: calc(50% - 56px);
  width: 24px;
  height: 24px;
  background: #16a34a;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid rgba(15, 32, 39, 0.9);
}

.profile-form :deep(.el-form-item__label) {
  color: rgba(255, 255, 255, 0.7);
  font-size: 13px;
  font-weight: 600;
}

.profile-form :deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.06) !important;
  box-shadow: 0 0 0 1px rgba(255, 255, 255, 0.12) !important;
}

.profile-form :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px rgba(255, 255, 255, 0.24) !important;
}

.profile-form :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 2px #16a34a !important;
}

.profile-form :deep(.el-input__inner) {
  color: #ffffff !important;
}

.avatar-hint {
  font-size: 11px;
  color: rgba(255, 255, 255, 0.35);
  margin-top: 4px;
}

.save-btn {
  width: 100%;
  height: 44px;
  font-size: 15px;
  font-weight: 600;
  background: linear-gradient(135deg, #16a34a, #15803d) !important;
  border-color: #15803d !important;
  box-shadow: 0 4px 16px rgba(22, 163, 74, 0.35) !important;
  margin-top: 8px;
}

.profile-meta {
  margin-top: 20px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.35);
}

.profile-meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #4ade80;
}

.profile-loading {
  padding: 20px 0;
}
</style>
