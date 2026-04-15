import { ref, computed } from 'vue'

export function usePagination(defaultLimit = 20) {
  const page = ref(1)
  const limit = ref(defaultLimit)
  const total = ref(0)

  const totalPages = computed(() => Math.ceil(total.value / limit.value) || 1)
  const hasNext = computed(() => page.value < totalPages.value)
  const hasPrev = computed(() => page.value > 1)

  function setTotal(t: number) {
    total.value = t
  }

  function nextPage() {
    if (hasNext.value) page.value++
  }

  function prevPage() {
    if (hasPrev.value) page.value--
  }

  function goToPage(p: number) {
    page.value = Math.max(1, Math.min(p, totalPages.value))
  }

  function reset() {
    page.value = 1
  }

  return { page, limit, total, totalPages, hasNext, hasPrev, setTotal, nextPage, prevPage, goToPage, reset }
}
