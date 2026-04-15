<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { ProductVariant } from '@/types'
import { formatPrice } from '@/utils/format'

const { t } = useI18n()
const props = defineProps<{
  variants: ProductVariant[]
  selectedId: string | null
}>()

const emit = defineEmits<{
  (e: 'select', variant: ProductVariant): void
}>()

function getAttributeString(attrs: Record<string, string>): string {
  return Object.entries(attrs).map(([k, v]) => `${k}: ${v}`).join(', ')
}
</script>

<template>
  <div v-if="variants.length > 0">
    <h3 class="text-sm font-medium text-gray-900 mb-2">{{ t('product.variants') }}</h3>
    <div class="grid grid-cols-2 sm:grid-cols-3 gap-2">
      <button v-for="variant in variants" :key="variant.id" @click="emit('select', variant)"
        :class="[
          selectedId === variant.id ? 'ring-2 ring-shinkansen-600 bg-shinkansen-50' : 'ring-1 ring-gray-200 hover:ring-shinkansen-300',
          'rounded-lg p-3 text-left transition-colors'
        ]">
        <p class="text-sm font-medium text-gray-900">{{ variant.name }}</p>
        <p v-if="variant.attributes" class="text-xs text-gray-500 mt-0.5">{{ getAttributeString(variant.attributes) }}</p>
        <p class="text-sm font-semibold text-shinkansen-600 mt-1">{{ formatPrice(variant.price) }}</p>
        <p class="text-xs mt-0.5" :class="variant.stock_quantity > 0 ? 'text-green-600' : 'text-red-600'">
          {{ variant.stock_quantity > 0 ? t('common.inStock') : t('common.outOfStock') }}
        </p>
      </button>
    </div>
  </div>
</template>
