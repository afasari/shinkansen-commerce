<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { CartItem, Money } from '@/types'
import { formatPrice } from '@/utils/format'
import { PlusIcon, MinusIcon, TrashIcon, ShoppingBagIcon } from '@heroicons/vue/24/outline'

const { t } = useI18n() as any

const props = defineProps<{
  item: CartItem
}>()

const emit = defineEmits<{
  (e: 'update', quantity: number): void
  (e: 'remove'): void
}>()

function lineTotal(): Money {
  return {
    currency: props.item.unit_price.currency,
    units: props.item.unit_price.units * props.item.quantity,
    nanos: 0,
  }
}
</script>

<template>
  <div class="flex gap-4 items-start">
    <div class="h-20 w-20 rounded-lg bg-gray-200 overflow-hidden flex-shrink-0">
      <img v-if="item.product_image" :src="item.product_image" :alt="item.product_name" class="w-full h-full object-cover" />
      <div v-else class="w-full h-full flex items-center justify-center">
        <ShoppingBagIcon class="h-6 w-6 text-gray-300" />
      </div>
    </div>
    <div class="flex-1 min-w-0">
      <p class="text-sm font-medium text-gray-900 line-clamp-1">{{ item.product_name }}</p>
      <p class="text-sm text-shinkansen-600 font-medium mt-0.5">{{ formatPrice(item.unit_price) }}</p>
    </div>
    <div class="flex items-center gap-2">
      <div class="flex items-center border border-gray-300 rounded-md">
        <button @click="emit('update', item.quantity - 1)" class="px-2 py-1 text-gray-500 hover:text-gray-700">
          <MinusIcon class="h-3.5 w-3.5" />
        </button>
        <span class="px-3 py-1 text-sm font-medium">{{ item.quantity }}</span>
        <button @click="emit('update', item.quantity + 1)" class="px-2 py-1 text-gray-500 hover:text-gray-700">
          <PlusIcon class="h-3.5 w-3.5" />
        </button>
      </div>
      <button @click="emit('remove')" class="p-1 text-gray-400 hover:text-red-500">
        <TrashIcon class="h-4 w-4" />
      </button>
    </div>
    <div class="text-right min-w-[4.5rem]">
      <p class="text-sm font-semibold text-gray-900">{{ formatPrice(lineTotal()) }}</p>
    </div>
  </div>
</template>
