<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import * as paymentApi from '@/api/payments'
import { PaymentStatus, type Payment } from '@/types'
import { formatPrice, formatDateTime } from '@/utils/format'
import StatusBadge from '@/components/common/StatusBadge.vue'

const { t } = useI18n()

const paymentId = ref('')
const payment = ref<Payment | null>(null)
const loading = ref(false)
const refunding = ref(false)

async function lookupPayment() {
  if (!paymentId.value) return
  loading.value = true
  try {
    payment.value = await paymentApi.getPayment(paymentId.value)
  } catch (e: unknown) {
    payment.value = null
    alert((e as Error).message)
  } finally {
    loading.value = false
  }
}

async function handleRefund() {
  if (!payment.value || !confirm(t('payment.refundConfirm'))) return
  refunding.value = true
  try {
    await paymentApi.refundPayment(payment.value.id, { amount: payment.value.amount })
    await lookupPayment()
  } catch (e: unknown) {
    alert((e as Error).message)
  } finally {
    refunding.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-6">{{ t('nav.payments') }} - {{ t('nav.management') }}</h1>

    <div class="card p-4 mb-6">
      <form @submit.prevent="lookupPayment" class="flex gap-3 items-end">
        <div class="flex-1">
          <label class="label-field">Payment ID</label>
          <input v-model="paymentId" required class="input-field mt-1" />
        </div>
        <button type="submit" :disabled="loading" class="btn-primary text-sm">{{ t('common.search') }}</button>
      </form>
    </div>

    <div v-if="loading" class="text-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-shinkansen-200 border-t-shinkansen-600 mx-auto"></div>
    </div>

    <template v-else-if="payment">
      <div class="card p-6 max-w-2xl">
        <div class="space-y-3 text-sm">
          <div class="flex justify-between">
            <span class="text-gray-600">ID</span>
            <span class="font-mono text-xs">{{ payment.id }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-600">{{ t('order.orderNumber') }}</span>
            <span class="font-mono text-xs">{{ payment.order_id }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-600">{{ t('common.price') }}</span>
            <span class="font-semibold">{{ formatPrice(payment.amount) }}</span>
          </div>
          <div class="flex justify-between items-center">
            <span class="text-gray-600">{{ t('payment.paymentStatus') }}</span>
            <StatusBadge :status="payment.status" size="sm" />
          </div>
          <div class="flex justify-between">
            <span class="text-gray-600">{{ t('payment.transactionId') }}</span>
            <span class="font-mono text-xs">{{ payment.transaction_id || '-' }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-600">{{ t('common.date') }}</span>
            <span>{{ formatDateTime(payment.created_at) }}</span>
          </div>
        </div>

        <div v-if="payment.status === PaymentStatus.COMPLETED" class="mt-6 pt-4 border-t">
          <button @click="handleRefund" :disabled="refunding" class="btn-danger text-sm">
            {{ t('payment.refund') }}
          </button>
        </div>
      </div>
    </template>
  </div>
</template>
