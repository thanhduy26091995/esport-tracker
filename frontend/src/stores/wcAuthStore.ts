import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { wcAuthService } from '@/services/wcAuthService'
import type { WcAuthUser } from '@/types/wc'
import { ElMessage } from 'element-plus'

const TOKEN_KEY = 'wc_token'
const USER_KEY = 'wc_user'

function loadUser(): WcAuthUser | null {
  try {
    const s = localStorage.getItem(USER_KEY)
    return s ? JSON.parse(s) : null
  } catch {
    return null
  }
}

export const useWcAuthStore = defineStore('wcAuth', () => {
  const token = ref<string | null>(localStorage.getItem(TOKEN_KEY))
  const user = ref<WcAuthUser | null>(loadUser())
  const loading = ref(false)

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.isAdmin ?? false)
  const userName = computed(() => user.value?.name ?? '')

  function _setAuth(t: string, u: WcAuthUser) {
    token.value = t
    user.value = u
    localStorage.setItem(TOKEN_KEY, t)
    localStorage.setItem(USER_KEY, JSON.stringify(u))
  }

  function logout() {
    token.value = null
    user.value = null
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(USER_KEY)
  }

  async function login(name: string, password: string) {
    loading.value = true
    try {
      const data = await wcAuthService.login(name, password)
      _setAuth(data.token, { id: data.user_id, name: data.name, isAdmin: data.is_admin })
      ElMessage.success(`Xin chào, ${data.name}!`)
    } finally {
      loading.value = false
    }
  }

  async function register(name: string, password: string) {
    loading.value = true
    try {
      const data = await wcAuthService.register(name, password)
      _setAuth(data.token, { id: data.user_id, name: data.name, isAdmin: data.is_admin })
      ElMessage.success(`Đăng ký thành công! Xin chào, ${data.name}!`)
    } finally {
      loading.value = false
    }
  }

  async function resetPassword(name: string) {
    loading.value = true
    try {
      await wcAuthService.resetPassword(name)
      ElMessage.success(`Mật khẩu đã được đặt lại thành ${name}_@123`)
    } finally {
      loading.value = false
    }
  }

  return {
    token,
    user,
    loading,
    isLoggedIn,
    isAdmin,
    userName,
    login,
    register,
    resetPassword,
    logout,
  }
})
