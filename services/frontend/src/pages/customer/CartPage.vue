<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useCartStore } from '@/stores/cart'
import { formatPrice } from '@/utils/format'
import { PlusIcon, MinusIcon, TrashIcon, ShoppingBagIcon } from '@heroicons/vue/24/outline'

const { t } = useI18n()
const router = useRouter()
const cartStore = useCartStore()

const isEmpty = computed(() => cartStore.items.length === 0)
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <h1 class="text-2xl font-bold text-gray-900 mb-6">{{ t('cart.title') }}</h1>

    <div v-if="isEmpty" class="text-center py-16">
      <ShoppingBagIcon class="h-16 w-16 text-gray-300 mx-auto mb-4" />
      <p class="text-gray-500 mb-4">{{ t('cart.empty') }}</p>
      <router-link to="/products" class="btn-primary">{{ t('cart.continueShopping') }}</router-link>
    </div>

    <div v-else class="lg:grid lg:grid-cols-12 lg:gap-8">
      <div class="lg:col-span-8">
        <div class="space-y-4">
          <div v-for="item in cartStore.items" :key="`${item.product_id}-${item.variant_id}`" class="card p-4 flex gap-4">
            <router-link :to="{ name: 'product-detail', params: { id: item.product_id } }" class="flex-shrink-0">
              <div class="h-24 w-24 rounded-lg bg-gray-200 overflow-hidden">
                <img v-if="item.product_image" :src="item.product_image" :alt="item.product_name" class="w-full h-full object-cover" />
                <div v-else class="w-full h-full flex items-center justify-center">
                  <ShoppingBagIcon class="h-8 w-8 text-gray-300" />
                </div>
              </div>
            </router-link>
            <div class="flex-1 min-w-0">
              <router-link :to="{ name: 'product-detail', params: { id: item.product_id } }"
                class="text-sm font-medium text-gray-900 hover:text-shinkansen-600 line-clamp-2">
                {{ item.product_name }}
              </router-link>
              <p v-if="item.variant_id" class="text-xs text-gray-500 mt-0.5">{{ item.variant_id }}</p>
              <p class="text-sm font-semibold text-shinkansen-600 mt-1">{{ formatPrice(item.unit_price) }}</p>
            </div>
            <div class="flex items-center gap-2">
              <div class="flex items-center border border-gray-300 rounded-md">
                <button @click="cartStore.updateQuantity(item.product_id, item.variant_id, item.quantity - 1)"
                  class="px-2 py-1 text-gray-500 hover:text-gray-700">
                  <MinusIcon class="h-3.5 w-3.5" />
                </button>
                <span class="px-3 py-1 text-sm font-medium">{{ item.quantity }}</span>
                <button @click="cartStore.updateQuantity(item.product_id, item.variant_id, item.quantity + 1)"
                  class="px-2 py-1 text-gray-500 hover:text-gray-700">
                  <PlusIcon class="h-3.5 w-3.5" />
                </button>
              </div>
              <button @click="cartStore.removeItem(item.product_id, item.variant_id)" class="p-1 text-gray-400 hover:text-red-500">
                <TrashIcon class="h-5 w-5" />
              </button>
            </div>
            <div class="text-right min-w-[5rem]">
              <p class="text-sm font-semibold text-gray-900">{{ formatPrice({ currency: item.unit_price.currency, units: item.unit_price.units * item.quantity, nanos: 0 }) }}</p>
            </div>
          </div>
        </div>
      </div>

      <div class="lg:col-span-4 mt-8 lg:mt-0">
        <div class="card p-6 sticky top-20">
          <h2 class="text-lg font-semibold text-gray-900">{{ t('cart.orderSummary') }}</h2>
          <div class="mt-4 space-y-2">
            <div class="flex justify-between text-sm">
              <span class="text-gray-600">{{ t('cart.totalItems', { count: cartStore.itemCount }) }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-600">{{ t('common.subtotal') }}</span>
              <span class="font-medium">{{ formatPrice(cartStore.subtotal) }}</span>
            </div>
          </div>
          <div class="mt-4 pt-4 border-t border-gray-200 flex justify-between">
            <span class="text-base font-semibold text-gray-900">{{ t('common.total') }}</span>
            <span class="text-base font-bold text-shinkansen-600">{{ formatPrice(cartStore.subtotal) }}</span>
          </div>
          <button @click="router.push('/checkout')" class="btn-primary w-full mt-6">{{ t('cart.checkout') }}</button>
          <router-link to="/products" class="block text-center mt-3 text-sm text-shinkansen-600 hover:text-shinkansen-500">
            {{ t('cart.continueShopping') }}
          </router-link>
        </div>
      </div>
    </div>
  </div>
</template>
