import { ref, onMounted, onUnmounted } from 'vue'
import axios from 'axios'

export function useHealthCheck(intervalMs = 30000) {
  const isHealthy = ref(true)
  const checking = ref(false)
  let timer: ReturnType<typeof setInterval> | null = null

  async function check() {
    checking.value = true
    try {
      await axios.get('/health', { timeout: 5000 })
      isHealthy.value = true
    } catch {
      isHealthy.value = false
    } finally {
      checking.value = false
    }
  }

  onMounted(() => {
    check()
    timer = setInterval(check, intervalMs)
  })

  onUnmounted(() => {
    if (timer) clearInterval(timer)
  })

  return { isHealthy, checking, check }
}
