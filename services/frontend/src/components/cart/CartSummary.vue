<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { Money } from '@/types'
import { formatPrice } from '@/utils/format'

const { t } = useI18n() as any

defineProps<{
  itemCount: number
  subtotal: Money
  tax?: Money
  discount?: Money
  total?: Money
  showTax?: boolean
}>()
</script>

<template>
  <div class="space-y-2 text-sm">
    <div class="flex justify-between">
      <span class="text-gray-600">{{ t('cart.totalItems', { count: itemCount }) }}</span>
    </div>
    <div class="flex justify-between">
      <span class="text-gray-600">{{ t('common.subtotal') }}</span>
      <span class="font-medium">{{ formatPrice(subtotal) }}</span>
    </div>
    <div v-if="showTax && tax" class="flex justify-between">
      <span class="text-gray-600">{{ t('common.tax') }}</span>
      <span class="font-medium">{{ formatPrice(tax) }}</span>
    </div>
    <div v-if="discount" class="flex justify-between">
      <span class="text-gray-600">{{ t('common.discount') }}</span>
      <span class="font-medium text-red-600">-{{ formatPrice(discount) }}</span>
    </div>
    <div v-if="total" class="border-t pt-2 flex justify-between">
      <span class="font-semibold text-gray-900">{{ t('common.total') }}</span>
      <span class="font-bold text-shinkansen-600">{{ formatPrice(total) }}</span>
    </div>
  </div>
</template>
