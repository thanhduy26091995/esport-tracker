export interface ApiError {
  code: string
  message: string
}

export interface ApiResponse<T = any> {
  data?: T
  error?: ApiError
}

export interface PaginationParams {
  page?: number
  limit?: number
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  limit: number
  total_pages: number
}
