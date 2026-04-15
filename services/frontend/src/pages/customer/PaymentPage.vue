<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useOrderStore } from '@/stores/order'
import * as paymentApi from '@/api/payments'
import { PaymentMethod, PaymentStatus, type Payment } from '@/types'
import { formatPrice, formatDateTime } from '@/utils/format'
import StatusBadge from '@/components/common/StatusBadge.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const orderStore = useOrderStore()

const payment = ref<Payment | null>(null)
const cardNumber = ref('')
const cardExpiry = ref('')
const cardCvv = ref('')
const loading = ref(false)
const processing = ref(false)
const error = ref('')
const paymentCreated = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    await orderStore.fetchOrder(route.params.id as string)
  } finally {
    loading.value = false
  }
})

const order = computed(() => orderStore.currentOrder)

async function createAndProcessPayment() {
  if (!order.value) return
  processing.value = true
  error.value = ''
  try {
    if (!paymentCreated.value) {
      const res = await paymentApi.createPayment({
        order_id: order.value.id,
        method: order.value.payment_method,
        amount: order.value.total_amount,
      })
      payment.value = await paymentApi.getPayment(res.payment_id)
      paymentCreated.value = true
    }

    if (order.value.payment_method === PaymentMethod.CREDIT_CARD) {
      const res = await paymentApi.processPayment(payment.value!.id, {
        payment_data: {
          card_number: cardNumber.value,
          expiry: cardExpiry.value,
          cvv: cardCvv.value,
        },
      })
      payment.value = await paymentApi.getPayment(payment.value!.id)
    } else {
      const res = await paymentApi.processPayment(payment.value!.id, {
        payment_data: {},
      })
      payment.value = await paymentApi.getPayment(payment.value!.id)
    }
  } catch (e: unknown) {
    error.value = (e as Error).message
  } finally {
    processing.value = false
  }
}

async function refreshPayment() {
  if (payment.value) {
    payment.value = await paymentApi.getPayment(payment.value.id)
  }
}
</script>

<template>
  <div class="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <button @click="router.back()" class="text-sm text-gray-500 hover:text-gray-700 mb-4">&larr; {{ t('common.back') }}</button>
    <h1 class="text-2xl font-bold text-gray-900 mb-6">{{ t('payment.title') }}</h1>

    <div v-if="loading" class="text-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-shinkansen-200 border-t-shinkansen-600 mx-auto"></div>
    </div>

    <template v-else-if="order">
      <div class="card p-6 mb-6">
        <div class="flex justify-between items-center mb-4">
          <h2 class="font-semibold">{{ order.order_number }}</h2>
          <span class="text-lg font-bold text-shinkansen-600">{{ formatPrice(order.total_amount) }}</span>
        </div>

        <div v-if="error" class="rounded-md bg-red-50 p-3 mb-4">
          <p class="text-sm text-red-700">{{ error }}</p>
        </div>

        <div v-if="payment">
          <div class="space-y-2 mb-4">
            <div class="flex justify-between text-sm">
              <span class="text-gray-600">{{ t('payment.paymentStatus') }}</span>
              <StatusBadge :status="payment.status" size="sm" />
            </div>
            <div v-if="payment.transaction_id" class="flex justify-between text-sm">
              <span class="text-gray-600">{{ t('payment.transactionId') }}</span>
              <span class="font-mono text-xs">{{ payment.transaction_id }}</span>
            </div>
          </div>

          <div v-if="payment.status === PaymentStatus.COMPLETED" class="rounded-md bg-green-50 p-4 text-center">
            <p class="text-green-700 font-medium">{{ t('payment.completed') }}!</p>
          </div>

          <div v-else-if="payment.status === PaymentStatus.PROCESSING" class="rounded-md bg-yellow-50 p-4">
            <p class="text-yellow-700 text-sm">{{ t('payment.processing') }}...</p>
            <button @click="refreshPayment" class="mt-2 text-sm text-yellow-700 underline">{{ t('common.view') }}</button>
          </div>
        </div>

        <div v-if="!payment || payment.status === PaymentStatus.PENDING">
          <div v-if="order.payment_method === PaymentMethod.CREDIT_CARD && !paymentCreated" class="space-y-3 mb-4">
            <div>
              <label class="label-field">{{ t('checkout.cardNumber') }}</label>
              <input v-model="cardNumber" class="input-field mt-1" placeholder="4242 4242 4242 4242" />
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="label-field">{{ t('checkout.expiry') }}</label>
                <input v-model="cardExpiry" class="input-field mt-1" placeholder="MM/YY" />
              </div>
              <div>
                <label class="label-field">{{ t('checkout.cvv') }}</label>
                <input v-model="cardCvv" class="input-field mt-1" placeholder="123" />
              </div>
            </div>
          </div>

          <button @click="createAndProcessPayment" :disabled="processing" class="btn-primary w-full">
            <span v-if="processing" class="animate-spin mr-2 inline-block h-4 w-4 border-2 border-white border-t-transparent rounded-full"></span>
            {{ t('payment.processPayment') }}
          </button>
        </div>
      </div>
    </template>
  </div>
</template>
