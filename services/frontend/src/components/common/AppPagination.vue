<script setup lang="ts">
defineProps<{
  currentPage: number
  totalPages: number
}>()

const emit = defineEmits<{
  (e: 'page-change', page: number): void
}>()

function getPageNumbers(current: number, total: number): (number | string)[] {
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)
  const pages: (number | string)[] = [1]
  if (current > 3) pages.push('...')
  for (let i = Math.max(2, current - 1); i <= Math.min(total - 1, current + 1); i++) {
    pages.push(i)
  }
  if (current < total - 2) pages.push('...')
  pages.push(total)
  return pages
}
</script>

<template>
  <div v-if="totalPages > 1" class="flex items-center justify-center gap-1 mt-6">
    <button :disabled="currentPage <= 1" @click="emit('page-change', currentPage - 1)"
      class="px-3 py-1.5 text-sm rounded-md border border-gray-300 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50">
      &laquo;
    </button>
    <template v-for="page in getPageNumbers(currentPage, totalPages)" :key="page">
      <button v-if="typeof page === 'number'" @click="emit('page-change', page)"
        :class="[page === currentPage ? 'bg-shinkansen-600 text-white' : 'border border-gray-300 hover:bg-gray-50', 'px-3 py-1.5 text-sm rounded-md']">
        {{ page }}
      </button>
      <span v-else class="px-2 py-1.5 text-sm text-gray-400">...</span>
    </template>
    <button :disabled="currentPage >= totalPages" @click="emit('page-change', currentPage + 1)"
      class="px-3 py-1.5 text-sm rounded-md border border-gray-300 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50">
      &raquo;
    </button>
  </div>
</template>
