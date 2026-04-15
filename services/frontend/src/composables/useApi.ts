import { ref } from 'vue'
import { useI18n } from 'vue-i18n'

export function useApi() {
  const { t } = useI18n()
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function execute<T>(fn: () => Promise<T>): Promise<T | null> {
    loading.value = true
    error.value = null
    try {
      return await fn()
    } catch (e: unknown) {
      const err = e as any
      error.value = err.response?.data?.message || err.message || t('common.error')
      return null
    } finally {
      loading.value = false
    }
  }

  function clearError() {
    error.value = null
  }

  return { loading, error, execute, clearError }
}
