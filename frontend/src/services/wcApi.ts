import axios from 'axios'
import { ElMessage } from 'element-plus'

const BASE = (import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1') + '/wc'

export const wcApi = axios.create({
  baseURL: BASE,
  timeout: 10000,
  headers: { 'Content-Type': 'application/json' },
})

wcApi.interceptors.request.use((config) => {
  const token = localStorage.getItem('wc_token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

wcApi.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('wc_token')
      localStorage.removeItem('wc_user')
      window.location.href = '/world-cup/login'
    } else if (error.response?.status === 503) {
      ElMessage.warning('Tính năng World Cup hiện đang tắt.')
    } else if (error.response) {
      const msg = error.response.data?.error || error.response.data?.message || 'Đã xảy ra lỗi'
      ElMessage.error(msg)
    } else if (error.request) {
      ElMessage.error('Lỗi kết nối mạng.')
    }
    return Promise.reject(error)
  },
)
