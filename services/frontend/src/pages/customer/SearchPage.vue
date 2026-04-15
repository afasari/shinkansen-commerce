<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useProductStore } from '@/stores/product'
import { formatPrice } from '@/utils/format'
import AppPagination from '@/components/common/AppPagination.vue'
import { ShoppingBagIcon, FunnelIcon } from '@heroicons/vue/24/outline'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const productStore = useProductStore()

const query = ref('')
const categoryId = ref('')
const minPrice = ref<number | undefined>()
const maxPrice = ref<number | undefined>()
const inStockOnly = ref(false)
const page = ref(1)
const showFilters = ref(false)
const totalPages = ref(1)

onMounted(() => {
  query.value = (route.query.q as string) || ''
  loadResults()
})

watch(() => route.query, () => {
  query.value = (route.query.q as string) || ''
  page.value = Number(route.query.page) || 1
  loadResults()
})

watch(() => productStore.pagination, (pag) => {
  totalPages.value = Math.ceil(pag.total / pag.limit) || 1
}, { deep: true })

function loadResults() {
  if (!query.value.trim()) return
  productStore.searchProducts({
    q: query.value,
    category_id: categoryId.value || undefined,
    min_price: minPrice.value,
    max_price: maxPrice.value,
    in_stock_only: inStockOnly.value,
    page: page.value,
    limit: 20,
  })
}

function handleSearch() {
  page.value = 1
  router.push({ query: { q: query.value, page: '1' } })
}

function applyFilters() {
  loadResults()
}

function goToPage(p: number) {
  page.value = p
  router.push({ query: { ...route.query, page: String(p) } })
}
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <div class="mb-6">
      <form @submit.prevent="handleSearch" class="flex gap-2 max-w-2xl">
        <input v-model="query" type="text" :placeholder="t('product.searchPlaceholder')" class="input-field flex-1" />
        <button type="submit" class="btn-primary">{{ t('common.search') }}</button>
        <button type="button" @click="showFilters = !showFilters" class="btn-secondary gap-1">
          <FunnelIcon class="h-4 w-4" />
        </button>
      </form>
    </div>

    <div v-if="showFilters" class="card p-4 mb-6">
      <div class="grid grid-cols-1 sm:grid-cols-4 gap-4">
        <div>
          <label class="label-field text-xs">{{ t('product.category') }} ID</label>
          <input v-model="categoryId" class="input-field mt-1 text-sm" />
        </div>
        <div>
          <label class="label-field text-xs">{{ t('product.minPrice') }}</label>
          <input v-model.number="minPrice" type="number" class="input-field mt-1 text-sm" />
        </div>
        <div>
          <label class="label-field text-xs">{{ t('product.maxPrice') }}</label>
          <input v-model.number="maxPrice" type="number" class="input-field mt-1 text-sm" />
        </div>
        <div class="flex flex-col justify-end gap-2">
          <label class="flex items-center gap-2 text-sm">
            <input type="checkbox" v-model="inStockOnly" class="rounded border-gray-300 text-shinkansen-600" />
            {{ t('product.inStockOnly') }}
          </label>
          <button @click="applyFilters" class="btn-primary text-sm">{{ t('common.filter') }}</button>
        </div>
      </div>
    </div>

    <div v-if="query && productStore.products.length > 0" class="text-sm text-gray-500 mb-4">
      {{ t('common.showing') }} {{ productStore.pagination.total }} {{ t('product.searchResults').toLowerCase() }}
    </div>

    <div v-if="productStore.loading" class="text-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-shinkansen-200 border-t-shinkansen-600 mx-auto"></div>
    </div>

    <div v-else-if="query && productStore.products.length === 0" class="text-center py-12 text-gray-500">
      {{ t('common.noResults') }}
    </div>

    <template v-else-if="productStore.products.length > 0">
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
        <router-link v-for="product in productStore.products" :key="product.id"
          :to="{ name: 'product-detail', params: { id: product.id } }"
          class="card group hover:shadow-md transition-shadow">
          <div class="aspect-square bg-gray-200 overflow-hidden">
            <img v-if="product.image_urls?.length" :src="product.image_urls[0]" :alt="product.name"
              class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300" />
            <div v-else class="w-full h-full flex items-center justify-center">
              <ShoppingBagIcon class="h-16 w-16 text-gray-300" />
            </div>
          </div>
          <div class="p-4">
            <h3 class="text-sm font-medium text-gray-900 line-clamp-2">{{ product.name }}</h3>
            <p class="mt-1 text-lg font-semibold text-shinkansen-600">{{ formatPrice(product.price) }}</p>
          </div>
        </router-link>
      </div>
      <AppPagination :current-page="page" :total-pages="totalPages" @page-change="goToPage" />
    </template>
  </div>
</template>
