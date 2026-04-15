import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User } from '@/types'
import * as authApi from '@/api/auth'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const accessToken = ref<string | null>(localStorage.getItem('access_token'))
  const refreshToken = ref<string | null>(localStorage.getItem('refresh_token'))

  const isAuthenticated = computed(() => !!accessToken.value)
  const userId = computed(() => localStorage.getItem('user_id'))
  const role = computed(() => user.value?.role || localStorage.getItem('user_role') || 'customer')
  const isAdmin = computed(() => role.value === 'admin')

  async function login(email: string, password: string) {
    const res = await authApi.login({ email, password })
    accessToken.value = res.access_token
    refreshToken.value = res.refresh_token
    if (res.role) {
      localStorage.setItem('user_role', res.role)
    }
    user.value = await authApi.getCurrentUser()
  }

  async function register(data: { email: string; password: string; name: string; phone: string }) {
    const res = await authApi.register(data)
    accessToken.value = res.access_token
    refreshToken.value = res.refresh_token
    if (res.role) {
      localStorage.setItem('user_role', res.role)
    }
    user.value = await authApi.getCurrentUser()
  }

  async function fetchUser() {
    if (!accessToken.value) return
    try {
      user.value = await authApi.getCurrentUser()
    } catch {
      user.value = null
    }
  }

  async function updateProfile(data: { name?: string; phone?: string }) {
    user.value = await authApi.updateCurrentUser(data)
  }

  function logout() {
    authApi.logout()
    user.value = null
    accessToken.value = null
    refreshToken.value = null
  }

  return { user, accessToken, refreshToken, isAuthenticated, userId, role, isAdmin, login, register, fetchUser, updateProfile, logout }
})
