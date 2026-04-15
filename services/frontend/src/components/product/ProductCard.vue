<script setup lang="ts">
import { formatPrice } from '@/utils/format'
import { useI18n } from 'vue-i18n'
import { ShoppingBagIcon } from '@heroicons/vue/24/outline'
import type { Product } from '@/types'

defineProps<{ product: Product }>()
const { t } = useI18n()
</script>

<template>
  <router-link :to="{ name: 'product-detail', params: { id: product.id } }" class="card group hover:shadow-md transition-shadow">
    <div class="aspect-square bg-gray-200 overflow-hidden">
      <img v-if="product.image_urls?.length" :src="product.image_urls[0]" :alt="product.name"
        class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300" loading="lazy" />
      <div v-else class="w-full h-full flex items-center justify-center">
        <ShoppingBagIcon class="h-16 w-16 text-gray-300" />
      </div>
    </div>
    <div class="p-4">
      <h3 class="text-sm font-medium text-gray-900 line-clamp-2">{{ product.name }}</h3>
      <p class="mt-1 text-lg font-semibold text-shinkansen-600">{{ formatPrice(product.price) }}</p>
      <div class="mt-1 flex items-center gap-1">
        <span class="text-xs" :class="product.stock_quantity > 0 ? 'text-green-600' : 'text-red-600'">
          {{ product.stock_quantity > 0 ? `${t('common.inStock')} (${product.stock_quantity})` : t('common.outOfStock') }}
        </span>
      </div>
    </div>
  </router-link>
</template>
