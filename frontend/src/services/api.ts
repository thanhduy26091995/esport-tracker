import axios from 'axios'
import { ElMessage } from 'element-plus'
import { translate, translateError } from '@/utils/i18n'

const TOKEN_KEY = 'site_access_token'

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1',
  headers: {
    'Content-Type': 'application/json',
    'X-Internal-Key': import.meta.env.VITE_INTERNAL_API_KEY || '',
  },
  timeout: 10000,
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem(TOKEN_KEY)
  if (token) config.headers['X-Site-Token'] = token
  return config
})

// Response interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 403) {
      const { useSiteAccessStore } = await import('@/stores/siteAccessStore')
      useSiteAccessStore().invalidate()
      return Promise.reject(error)
    }
    if (error.response) {
      // Server error
      const message = translateError(
        error.response.data?.error?.code || error.response.data?.code,
        error.response.data?.error?.message || error.response.data?.message,
      )
      ElMessage.error(message)
    } else if (error.request) {
      // Network error
      ElMessage.error(translate('errors.network'))
    }
    return Promise.reject(error)
  }
)
