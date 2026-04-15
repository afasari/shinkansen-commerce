<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useOrderStore } from '@/stores/order'
import { formatPrice, formatDate, formatDateTime } from '@/utils/format'
import { OrderStatus, CANCELLABLE_STATUSES, ORDER_STATUS_TRANSITIONS } from '@/utils/constants'
import StatusBadge from '@/components/common/StatusBadge.vue'
import AppModal from '@/components/common/AppModal.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const orderStore = useOrderStore()

const showCancelModal = ref(false)
const cancelReason = ref('')

onMounted(() => {
  orderStore.fetchOrder(route.params.id as string)
})

const order = computed(() => orderStore.currentOrder)
const canCancel = computed(() => order.value ? CANCELLABLE_STATUSES.includes(order.value.status) : false)

async function handleCancel() {
  await orderStore.cancelOrder(route.params.id as string, cancelReason.value)
  showCancelModal.value = false
  cancelReason.value = ''
}

function getPaymentMethodLabel(method: string): string {
  const map: Record<string, string> = {
    PAYMENT_METHOD_CREDIT_CARD: t('checkout.creditCard'),
    PAYMENT_METHOD_KONBINI_SEVENELEVEN: t('checkout.konbiniSevenEleven'),
    PAYMENT_METHOD_KONBINI_LAWSON: t('checkout.konbiniLawson'),
    PAYMENT_METHOD_KONBINI_FAMILYMART: t('checkout.konbiniFamilyMart'),
    PAYMENT_METHOD_PAYPAY: t('checkout.paypay'),
    PAYMENT_METHOD_RAKUTEN_PAY: t('checkout.rakutenPay'),
  }
  return map[method] || method
}
</script>

<template>
  <div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <div v-if="orderStore.loading" class="text-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-shinkansen-200 border-t-shinkansen-600 mx-auto"></div>
    </div>

    <template v-else-if="order">
      <div class="flex items-center justify-between mb-6">
        <div>
          <button @click="router.back()" class="text-sm text-gray-500 hover:text-gray-700 mb-2">&larr; {{ t('common.back') }}</button>
          <h1 class="text-2xl font-bold text-gray-900">{{ t('order.orderDetails') }}</h1>
          <p class="text-sm text-gray-500 mt-1">{{ order.order_number }}</p>
        </div>
        <div class="flex items-center gap-3">
          <StatusBadge :status="order.status" />
          <button v-if="canCancel" @click="showCancelModal = true" class="btn-danger text-sm">{{ t('order.cancelOrder') }}</button>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div class="lg:col-span-2 space-y-6">
          <div class="card p-4">
            <h2 class="font-semibold text-gray-900 mb-3">{{ t('common.items') }}</h2>
            <div class="divide-y divide-gray-100">
              <div v-for="item in order.items" :key="item.id" class="flex items-center justify-between py-3">
                <div>
                  <p class="text-sm font-medium text-gray-900">{{ item.product_name }}</p>
                  <p class="text-xs text-gray-500">{{ t('common.quantity') }}: {{ item.quantity }}</p>
                </div>
                <p class="text-sm font-semibold">{{ formatPrice(item.total_price) }}</p>
              </div>
            </div>
          </div>

          <div class="card p-4">
            <h2 class="font-semibold text-gray-900 mb-3">{{ t('order.shippingAddress') }}</h2>
            <p class="text-sm text-gray-600">{{ order.shipping_address.name }}</p>
            <p class="text-sm text-gray-600">{{ order.shipping_address.phone }}</p>
            <p class="text-sm text-gray-600">{{ order.shipping_address.postal_code }} {{ order.shipping_address.prefecture }} {{ order.shipping_address.city }}</p>
            <p class="text-sm text-gray-600">{{ order.shipping_address.address_line1 }} {{ order.shipping_address.address_line2 }}</p>
          </div>

          <div class="card p-4">
            <h2 class="font-semibold text-gray-900 mb-3">{{ t('order.paymentMethod') }}</h2>
            <p class="text-sm text-gray-600">{{ getPaymentMethodLabel(order.payment_method) }}</p>
          </div>

          <div class="flex gap-3">
            <router-link :to="{ name: 'payment', params: { id: order.id } }" class="btn-secondary text-sm">{{ t('payment.title') }}</router-link>
            <router-link :to="{ name: 'shipment-tracking', params: { id: order.id } }" class="btn-secondary text-sm">{{ t('order.trackShipment') }}</router-link>
          </div>
        </div>

        <div>
          <div class="card p-4 sticky top-20">
            <h2 class="font-semibold text-gray-900 mb-3">{{ t('cart.orderSummary') }}</h2>
            <div class="space-y-2 text-sm">
              <div class="flex justify-between"><span class="text-gray-600">{{ t('common.subtotal') }}</span><span>{{ formatPrice(order.subtotal_amount) }}</span></div>
              <div class="flex justify-between"><span class="text-gray-600">{{ t('common.tax') }}</span><span>{{ formatPrice(order.tax_amount) }}</span></div>
              <div v-if="Number(order.points_applied) > 0" class="flex justify-between"><span class="text-gray-600">{{ t('common.discount') }} ({{ order.points_applied }} pts)</span><span class="text-red-600">-{{ formatPrice(order.discount_amount) }}</span></div>
              <div class="border-t pt-2 flex justify-between font-semibold">
                <span>{{ t('common.total') }}</span>
                <span class="text-shinkansen-600">{{ formatPrice(order.total_amount) }}</span>
              </div>
            </div>
            <div class="mt-4 pt-4 border-t text-xs text-gray-500 space-y-1">
              <p>{{ t('common.date') }}: {{ formatDateTime(order.created_at) }}</p>
              <p>{{ t('common.status') }}: <StatusBadge :status="order.status" size="sm" /></p>
            </div>
          </div>
        </div>
      </div>
    </template>

    <AppModal :open="showCancelModal" :message="t('order.cancelConfirm')" :danger="true"
      :confirm-label="t('order.cancelOrder')" @close="showCancelModal = false" @confirm="handleCancel">
      <textarea v-model="cancelReason" :placeholder="t('order.cancelReason')" class="input-field mt-2" rows="3"></textarea>
    </AppModal>
  </div>
</template>
