import axios from 'axios'
import { ElMessage } from 'element-plus'
import { translate, translateError } from '@/utils/i18n'

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 10000,
})

// Response interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
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
