<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useProductStore } from '@/stores/product'
import { useCartStore } from '@/stores/cart'
import { formatPrice } from '@/utils/format'
import type { ProductVariant, Money } from '@/types'
import { ShoppingBagIcon, PlusIcon, MinusIcon, CheckIcon } from '@heroicons/vue/24/outline'

const { t } = useI18n()
const route = useRoute()
const productStore = useProductStore()
const cartStore = useCartStore()

const selectedVariant = ref<ProductVariant | null>(null)
const quantity = ref(1)
const addedToCart = ref(false)

onMounted(async () => {
  const id = route.params.id as string
  await productStore.fetchProduct(id)
  await productStore.fetchVariants(id)
})

const product = computed(() => productStore.currentProduct)

function selectVariant(variant: ProductVariant) {
  selectedVariant.value = variant
  quantity.value = 1
}

function getVariantAttributes(variant: ProductVariant): string {
  return Object.entries(variant.attributes || {}).map(([k, v]) => `${k}: ${v}`).join(', ')
}

function addToCart() {
  if (!product.value) return
  const price: Money = selectedVariant.value?.price || product.value.price
  cartStore.addItem({
    product_id: product.value.id,
    variant_id: selectedVariant.value?.id || '',
    product_name: product.value.name,
    product_image: product.value.image_urls?.[0] || '',
    unit_price: price,
    quantity: quantity.value,
    stock_quantity: selectedVariant.value?.stock_quantity ?? product.value.stock_quantity,
  })
  addedToCart.value = true
  setTimeout(() => { addedToCart.value = false }, 2000)
}

const currentPrice = computed<Money>(() => selectedVariant.value?.price || product.value?.price || { currency: 'JPY', units: 0, nanos: 0 })
const currentStock = computed(() => selectedVariant.value?.stock_quantity ?? product.value?.stock_quantity ?? 0)
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <div v-if="productStore.loading" class="text-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-shinkansen-200 border-t-shinkansen-600 mx-auto"></div>
    </div>

    <template v-else-if="product">
      <div class="lg:grid lg:grid-cols-2 lg:gap-x-8">
        <div class="aspect-square bg-gray-200 rounded-xl overflow-hidden mb-6 lg:mb-0">
          <img v-if="product.image_urls?.length" :src="product.image_urls[0]" :alt="product.name"
            class="w-full h-full object-cover" />
          <div v-else class="w-full h-full flex items-center justify-center">
            <ShoppingBagIcon class="h-24 w-24 text-gray-300" />
          </div>
        </div>

        <div class="space-y-6">
          <div>
            <h1 class="text-2xl font-bold text-gray-900">{{ product.name }}</h1>
            <p class="mt-1 text-sm text-gray-500">SKU: {{ product.sku }}</p>
          </div>

          <p class="text-3xl font-bold text-shinkansen-600">{{ formatPrice(currentPrice) }}</p>

          <div>
            <h3 class="text-sm font-medium text-gray-900 mb-2">{{ t('product.description') }}</h3>
            <p class="text-sm text-gray-600 whitespace-pre-line">{{ product.description }}</p>
          </div>

          <div v-if="productStore.variants.length > 0">
            <h3 class="text-sm font-medium text-gray-900 mb-2">{{ t('product.variants') }}</h3>
            <div class="grid grid-cols-2 sm:grid-cols-3 gap-2">
              <button v-for="variant in productStore.variants" :key="variant.id" @click="selectVariant(variant)"
                :class="[
                  selectedVariant?.id === variant.id ? 'ring-2 ring-shinkansen-600 bg-shinkansen-50' : 'ring-1 ring-gray-200 hover:ring-shinkansen-300',
                  'rounded-lg p-3 text-left'
                ]">
                <p class="text-sm font-medium text-gray-900">{{ variant.name }}</p>
                <p class="text-xs text-gray-500 mt-0.5">{{ getVariantAttributes(variant) }}</p>
                <p class="text-sm font-semibold text-shinkansen-600 mt-1">{{ formatPrice(variant.price) }}</p>
              </button>
            </div>
          </div>

          <div class="flex items-center gap-1 text-sm" :class="currentStock > 0 ? 'text-green-600' : 'text-red-600'">
            {{ currentStock > 0 ? `${t('common.inStock')} (${currentStock})` : t('common.outOfStock') }}
          </div>

          <div class="flex items-center gap-4">
            <div class="flex items-center border border-gray-300 rounded-lg">
              <button @click="quantity = Math.max(1, quantity - 1)" class="px-3 py-2 text-gray-600 hover:text-gray-900">
                <MinusIcon class="h-4 w-4" />
              </button>
              <span class="px-4 py-2 text-sm font-medium min-w-[3rem] text-center">{{ quantity }}</span>
              <button @click="quantity = Math.min(currentStock, quantity + 1)" class="px-3 py-2 text-gray-600 hover:text-gray-900">
                <PlusIcon class="h-4 w-4" />
              </button>
            </div>
            <button @click="addToCart" :disabled="currentStock === 0"
              :class="[addedToCart ? 'bg-green-600' : 'bg-shinkansen-600 hover:bg-shinkansen-700', 'btn-primary flex-1 justify-center gap-2']">
              <CheckIcon v-if="addedToCart" class="h-5 w-5" />
              {{ addedToCart ? t('common.success') + '!' : t('product.addToCart') }}
            </button>
          </div>

          <div v-if="product.image_urls?.length > 1" class="mt-6">
            <h3 class="text-sm font-medium text-gray-900 mb-2">Images</h3>
            <div class="flex gap-2 overflow-x-auto">
              <img v-for="(url, idx) in product.image_urls" :key="idx" :src="url"
                class="h-20 w-20 rounded-lg object-cover flex-shrink-0 ring-1 ring-gray-200" />
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
