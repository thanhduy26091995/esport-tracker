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
  const avatarUrl = computed(() => user.value?.avatarUrl ?? null)
  const googleLinked = computed(() => user.value?.googleLinked ?? false)

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
      _setAuth(data.token, {
        id: data.user_id,
        name: data.name,
        isAdmin: data.is_admin,
        avatarUrl: data.avatar_url,
        googleLinked: data.google_linked,
      })
      ElMessage.success(`Xin chào, ${data.name}!`)
    } finally {
      loading.value = false
    }
  }

  async function loginWithGoogle(idToken: string) {
    loading.value = true
    try {
      const data = await wcAuthService.googleLogin(idToken)
      _setAuth(data.token, {
        id: data.user_id,
        name: data.name,
        isAdmin: data.is_admin,
        avatarUrl: data.avatar_url,
        googleLinked: data.google_linked,
      })
      ElMessage.success(`Xin chào, ${data.name}!`)
    } finally {
      loading.value = false
    }
  }

  async function linkGoogle(idToken: string): Promise<boolean> {
    loading.value = true
    try {
      const data = await wcAuthService.googleLink(idToken)
      if (user.value) {
        user.value = { ...user.value, googleLinked: data.google_linked, avatarUrl: data.avatar_url }
        localStorage.setItem(USER_KEY, JSON.stringify(user.value))
      }
      ElMessage.success('Liên kết Google thành công!')
      return true
    } catch {
      return false
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
    avatarUrl,
    googleLinked,
    login,
    loginWithGoogle,
    linkGoogle,
    logout,
  }
})
