import axios from 'axios'
import type { AuthResponse } from '@/types'

const client = axios.create({
  baseURL: '',
  headers: {
    'Content-Type': 'application/json',
  },
})

function getAccessToken(): string | null {
  return localStorage.getItem('access_token')
}

function getRefreshToken(): string | null {
  return localStorage.getItem('refresh_token')
}

function setTokens(response: AuthResponse) {
  localStorage.setItem('access_token', response.access_token)
  localStorage.setItem('refresh_token', response.refresh_token)
  localStorage.setItem('user_id', response.user_id)
  if (response.role) {
    localStorage.setItem('user_role', response.role)
  }
}

function clearTokens() {
  localStorage.removeItem('access_token')
  localStorage.removeItem('refresh_token')
  localStorage.removeItem('user_id')
  localStorage.removeItem('user_role')
}

client.interceptors.request.use((config) => {
  const token = getAccessToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

let isRefreshing = false
let failedQueue: Array<{
  resolve: (value: unknown) => void
  reject: (reason?: unknown) => void
}> = []

function processQueue(error: unknown, token: string | null = null) {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error)
    } else {
      prom.resolve(token)
    }
  })
  failedQueue = []
}

client.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config

    if (error.response?.status === 401 && !originalRequest._retry) {
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject })
        }).then((token) => {
          originalRequest.headers.Authorization = `Bearer ${token}`
          return client(originalRequest)
        })
      }

      originalRequest._retry = true
      isRefreshing = true

      const refreshToken = getRefreshToken()
      if (!refreshToken) {
        clearTokens()
        window.location.href = '/login'
        return Promise.reject(error)
      }

      try {
        const { data } = await axios.post('/v1/users/login', {
          grant_type: 'refresh_token',
          refresh_token: refreshToken,
        })

        const authData = data as AuthResponse
        setTokens(authData)
        processQueue(null, authData.access_token)

        originalRequest.headers.Authorization = `Bearer ${authData.access_token}`
        return client(originalRequest)
      } catch (refreshError) {
        processQueue(refreshError, null)
        clearTokens()
        window.location.href = '/login'
        return Promise.reject(refreshError)
      } finally {
        isRefreshing = false
      }
    }

    return Promise.reject(error)
  },
)

export { client, getAccessToken, getRefreshToken, setTokens, clearTokens }
