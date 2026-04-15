export interface Money {
  currency: string
  units: number
  nanos?: number
}

export interface Pagination {
  page: number
  limit: number
  total: number
}

export interface ApiError {
  code: string
  message: string
  details?: Record<string, string>
}

export interface PaginatedResponse<T> {
  data: T[]
  pagination: Pagination
}
