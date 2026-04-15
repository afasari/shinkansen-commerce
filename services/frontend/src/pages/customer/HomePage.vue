<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useProductStore } from '@/stores/product'
import { formatPrice } from '@/utils/format'
import { ShoppingBagIcon } from '@heroicons/vue/24/outline'

const { t } = useI18n()
const router = useRouter()
const productStore = useProductStore()
const searchQuery = ref('')

onMounted(async () => {
  await productStore.fetchProducts({ limit: 8 })
})

function handleSearch() {
  if (searchQuery.value.trim()) {
    router.push({ name: 'search', query: { q: searchQuery.value.trim() } })
  }
}
</script>

<template>
  <div>
    <section class="bg-gradient-to-r from-shinkansen-600 to-shinkansen-800 text-white">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16 sm:py-24">
        <h1 class="text-3xl sm:text-4xl font-bold">{{ t('app.name') }}</h1>
        <p class="mt-3 text-lg text-shinkansen-100">{{ t('app.tagline') }}</p>
        <form @submit.prevent="handleSearch" class="mt-8 max-w-lg">
          <div class="flex rounded-lg shadow-sm">
            <input v-model="searchQuery" type="text" :placeholder="t('product.searchPlaceholder')"
              class="flex-1 rounded-l-lg border-0 px-4 py-3 text-gray-900 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-white sm:text-sm" />
            <button type="submit" class="rounded-r-lg bg-white px-6 py-3 text-sm font-semibold text-shinkansen-600 hover:bg-gray-50">
              {{ t('common.search') }}
            </button>
          </div>
        </form>
      </div>
    </section>

    <section class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
      <div class="flex items-center justify-between mb-6">
        <h2 class="text-2xl font-bold text-gray-900">{{ t('product.featured') }}</h2>
        <router-link to="/products" class="text-sm font-medium text-shinkansen-600 hover:text-shinkansen-500">
          {{ t('product.viewAll') }} &rarr;
        </router-link>
      </div>

      <div v-if="productStore.loading" class="text-center py-12">
        <div class="h-8 w-8 animate-spin rounded-full border-4 border-shinkansen-200 border-t-shinkansen-600 mx-auto"></div>
      </div>

      <div v-else-if="productStore.products.length === 0" class="text-center py-12 text-gray-500">
        {{ t('product.noProducts') }}
      </div>

      <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
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
    </section>
  </div>
</template>
