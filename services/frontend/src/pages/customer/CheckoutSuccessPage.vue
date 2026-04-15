<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useCheckoutStore } from '@/stores/checkout'
import { useRouter } from 'vue-router'
import { CheckCircleIcon } from '@heroicons/vue/24/outline'

const { t } = useI18n()
const checkout = useCheckoutStore()
const router = useRouter()
</script>

<template>
  <div class="max-w-lg mx-auto px-4 py-16 text-center">
    <CheckCircleIcon class="h-16 w-16 text-green-500 mx-auto mb-4" />
    <h1 class="text-2xl font-bold text-gray-900">{{ t('checkout.orderPlaced') }}</h1>
    <p class="mt-2 text-gray-600">{{ t('checkout.orderPlacedMessage', { orderNumber: checkout.lastOrderNumber || '' }) }}</p>
    <div class="mt-8 flex items-center justify-center gap-4">
      <button v-if="checkout.lastOrderId" @click="router.push({ name: 'order-detail', params: { id: checkout.lastOrderId } })" class="btn-primary">
        {{ t('checkout.viewOrder') }}
      </button>
      <button @click="checkout.reset(); router.push('/')" class="btn-secondary">
        {{ t('checkout.continueShopping') }}
      </button>
    </div>
  </div>
</template>
