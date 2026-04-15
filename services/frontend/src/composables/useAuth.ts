import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

export function useAuth() {
  const authStore = useAuthStore()
  const router = useRouter()

  const isAuthenticated = computed(() => authStore.isAuthenticated)
  const user = computed(() => authStore.user)
  const userName = computed(() => authStore.user?.name || '')

  async function login(email: string, password: string) {
    await authStore.login(email, password)
  }

  async function register(data: { email: string; password: string; name: string; phone: string }) {
    await authStore.register(data)
  }

  function logout() {
    authStore.logout()
    router.push('/')
  }

  async function fetchUser() {
    await authStore.fetchUser()
  }

  return { isAuthenticated, user, userName, login, register, logout, fetchUser }
}
