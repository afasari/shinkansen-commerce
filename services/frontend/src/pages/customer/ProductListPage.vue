<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useProductStore } from '@/stores/product'
import { formatPrice } from '@/utils/format'
import AppPagination from '@/components/common/AppPagination.vue'
import { ShoppingBagIcon } from '@heroicons/vue/24/outline'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const productStore = useProductStore()

const page = ref(1)
const activeOnly = ref(true)

onMounted(() => {
  page.value = Number(route.query.page) || 1
  loadProducts()
})

watch(() => route.query, () => {
  page.value = Number(route.query.page) || 1
  loadProducts()
})

function loadProducts() {
  productStore.fetchProducts({
    category_id: route.query.category_id as string || undefined,
    active_only: activeOnly.value,
    page: page.value,
    limit: 20,
  })
}

function goToPage(p: number) {
  router.push({ query: { ...route.query, page: String(p) } })
}

const totalPages = ref(1)
watch(() => productStore.pagination, (pag) => {
  totalPages.value = Math.ceil(pag.total / pag.limit) || 1
}, { deep: true })
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-gray-900">{{ t('product.title') }}</h1>
      <div class="flex items-center gap-4">
        <label class="flex items-center gap-2 text-sm text-gray-600">
          <input type="checkbox" v-model="activeOnly" @change="loadProducts" class="rounded border-gray-300 text-shinkansen-600 focus:ring-shinkansen-500" />
          {{ t('product.activeOnly') }}
        </label>
      </div>
    </div>

    <div v-if="productStore.loading" class="text-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-shinkansen-200 border-t-shinkansen-600 mx-auto"></div>
    </div>

    <div v-else-if="productStore.products.length === 0" class="text-center py-12 text-gray-500">
      {{ t('product.noProducts') }}
    </div>

    <template v-else>
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
            <p class="mt-1 text-xs" :class="product.stock_quantity > 0 ? 'text-green-600' : 'text-red-600'">
              {{ product.stock_quantity > 0 ? t('common.inStock') : t('common.outOfStock') }}
            </p>
          </div>
        </router-link>
      </div>
      <AppPagination :current-page="page" :total-pages="totalPages" @page-change="goToPage" />
    </template>
  </div>
</template>
