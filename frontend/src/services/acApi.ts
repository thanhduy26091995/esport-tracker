import axios from 'axios'
import { ElMessage } from 'element-plus'

const BASE = (import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1') + '/ac'

export const acApi = axios.create({
  baseURL: BASE,
  timeout: 10000,
  headers: { 'Content-Type': 'application/json' },
})

const ERROR_MAP: [string, string][] = [
  ['invalid name or password',                   'Email hoặc mật khẩu không đúng'],
  ['already taken',                              'Tài khoản này đã tồn tại'],
  ['name cannot be empty',                       'Vui lòng nhập email'],
  ['name must be at least',                      'Email phải có ít nhất 2 ký tự'],
  ['user is blocked from placing predictions',   'Tài khoản của bạn đang bị khoá, không thể dự đoán. Vui lòng liên hệ admin.'],
  ['user is blocked from placing bets',          'Tài khoản của bạn đang bị khoá, không thể đặt cược. Vui lòng liên hệ admin.'],
  ['predictions are closed for this match',      'Trận đấu hiện không mở dự đoán'],
  ['cannot modify prediction: match is locked',  'Trận đấu đã khóa, không thể chỉnh sửa'],
  ['cannot delete a finalized prediction',       'Không thể xóa dự đoán đã có kết quả'],
  ['cannot modify a finalized prediction',       'Không thể sửa dự đoán đã có kết quả'],
  ['points must be greater than 0',             'Số điểm phải lớn hơn 0'],
  ['failed to submit prediction',               'Dự đoán thất bại (có thể đã tồn tại)'],
  ['scoreline',                                  'Tỉ số này chưa được cấu hình để dự đoán'],
  ['handicap odds not set',                      'Trận đấu chưa có chấp điểm'],
  ['prediction_choice must be',                  'Vui lòng chọn đội để dự đoán'],
  ['predicted_home_score and predicted_away_score are required', 'Vui lòng nhập tỉ số dự đoán'],
  ['prediction not found',                       'Không tìm thấy dự đoán'],
  ['match score not set',                        'Chưa có tỉ số trận đấu'],
  ['match is already completed or cancelled',    'Trận đấu đã kết thúc hoặc bị hủy'],
  ['match not found',                            'Không tìm thấy trận đấu'],
  ['score odds not found',                       'Không tìm thấy hệ số tỉ số'],
  ['invalid match id',                           'Mã trận đấu không hợp lệ'],
  ['invalid prediction id',                      'Mã dự đoán không hợp lệ'],
  ['not found',                                  'Không tìm thấy'],
  ['delta cannot be 0',                          'Số điểm thay đổi không được bằng 0'],
  ['wallet not found',                           'Không tìm thấy ví người dùng'],
  ['failed to create settlement',                'Tạo tất toán thất bại, vui lòng thử lại'],
  ['unauthorized',                               'Bạn không có quyền thực hiện thao tác này'],
]

function friendlyError(raw: string): string {
  const lower = raw.toLowerCase()
  for (const [key, msg] of ERROR_MAP) {
    if (lower.includes(key.toLowerCase())) return msg
  }
  return 'Đã xảy ra lỗi, vui lòng thử lại'
}

acApi.interceptors.request.use((config) => {
  const token = localStorage.getItem('wc_token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

acApi.interceptors.response.use(
  (response) => response,
  (error) => {
    const url = error.config?.url ?? ''
    if (error.response?.status === 401 && !url.includes('/auth/')) {
      localStorage.removeItem('wc_token')
      localStorage.removeItem('wc_user')
      window.location.href = '/asean-cup/login'
    } else if (error.response?.status === 503) {
      ElMessage.warning('Tính năng ASEAN Cup hiện đang tắt.')
    } else if (error.response) {
      const raw = error.response.data?.error || error.response.data?.message || ''
      ElMessage.error(friendlyError(raw))
    } else if (error.request) {
      ElMessage.error('Lỗi kết nối mạng.')
    }
    return Promise.reject(error)
  },
)
