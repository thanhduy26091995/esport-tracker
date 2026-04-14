import axios from 'axios'
import { ElMessage } from 'element-plus'

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
      const message = error.response.data?.error?.message || 'An error occurred'
      ElMessage.error(message)
    } else if (error.request) {
      // Network error
      ElMessage.error('Network error. Please check your connection.')
    }
    return Promise.reject(error)
  }
)
