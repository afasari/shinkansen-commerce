<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useOrderStore } from '@/stores/order'
import { formatPrice, formatDate, formatDateTime } from '@/utils/format'
import { OrderStatus, ORDER_STATUS_TRANSITIONS, CANCELLABLE_STATUSES } from '@/utils/constants'
import StatusBadge from '@/components/common/StatusBadge.vue'
import AppModal from '@/components/common/AppModal.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const orderStore = useOrderStore()

const showCancelModal = ref(false)
const cancelReason = ref('')
const updating = ref(false)

onMounted(() => {
  orderStore.fetchOrder(route.params.id as string)
})

const order = computed(() => orderStore.currentOrder)
const canCancel = computed(() => order.value ? CANCELLABLE_STATUSES.includes(order.value.status) : false)
const availableTransitions = computed(() => {
  if (!order.value) return []
  return ORDER_STATUS_TRANSITIONS[order.value.status] || []
})

async function handleStatusUpdate(status: OrderStatus) {
  updating.value = true
  try {
    await orderStore.updateStatus(route.params.id as string, status)
  } catch (e: unknown) {
    alert((e as Error).message)
  } finally {
    updating.value = false
  }
}

async function handleCancel() {
  await orderStore.cancelOrder(route.params.id as string, cancelReason.value)
  showCancelModal.value = false
}
</script>

<template>
  <div>
    <button @click="router.back()" class="text-sm text-gray-500 hover:text-gray-700 mb-4">&larr; {{ t('common.back') }}</button>

    <div v-if="orderStore.loading" class="text-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-shinkansen-200 border-t-shinkansen-600 mx-auto"></div>
    </div>

    <template v-else-if="order">
      <div class="flex items-center justify-between mb-6">
        <div>
          <h1 class="text-2xl font-bold text-gray-900">{{ order.order_number }}</h1>
          <p class="text-sm text-gray-500">{{ formatDateTime(order.created_at) }}</p>
        </div>
        <StatusBadge :status="order.status" />
      </div>

      <div v-if="availableTransitions.length > 0 || canCancel" class="card p-4 mb-6">
        <h2 class="text-sm font-semibold text-gray-900 mb-3">{{ t('order.updateStatus') }}</h2>
        <div class="flex flex-wrap gap-2">
          <button v-for="status in availableTransitions" :key="status" @click="handleStatusUpdate(status)" :disabled="updating"
            class="btn-secondary text-sm">
            {{ status.replace('ORDER_STATUS_', '').replace(/_/g, ' ') }}
          </button>
          <button v-if="canCancel" @click="showCancelModal = true" class="btn-danger text-sm">{{ t('order.cancelOrder') }}</button>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div class="lg:col-span-2 space-y-6">
          <div class="card p-4">
            <h2 class="font-semibold text-gray-900 mb-3">{{ t('common.items') }}</h2>
            <div class="divide-y divide-gray-100">
              <div v-for="item in order.items" :key="item.id" class="flex justify-between py-2 text-sm">
                <div>
                  <span class="font-medium">{{ item.product_name }}</span>
                  <span class="text-gray-500 ml-2">x{{ item.quantity }}</span>
                </div>
                <span class="font-semibold">{{ formatPrice(item.total_price) }}</span>
              </div>
            </div>
          </div>

          <div class="card p-4">
            <h2 class="font-semibold text-gray-900 mb-2">{{ t('order.shippingAddress') }}</h2>
            <p class="text-sm text-gray-600">{{ order.shipping_address.name }} - {{ order.shipping_address.phone }}</p>
            <p class="text-sm text-gray-600">{{ order.shipping_address.postal_code }} {{ order.shipping_address.prefecture }} {{ order.shipping_address.city }}</p>
            <p class="text-sm text-gray-600">{{ order.shipping_address.address_line1 }} {{ order.shipping_address.address_line2 }}</p>
          </div>

          <div class="card p-4">
            <h2 class="font-semibold text-gray-900 mb-2">{{ t('order.paymentMethod') }}</h2>
            <p class="text-sm text-gray-600">{{ order.payment_method.replace('PAYMENT_METHOD_', '').replace(/_/g, ' ') }}</p>
          </div>
        </div>

        <div class="card p-4 h-fit">
          <h2 class="font-semibold text-gray-900 mb-3">{{ t('cart.orderSummary') }}</h2>
          <div class="space-y-2 text-sm">
            <div class="flex justify-between"><span class="text-gray-600">{{ t('common.subtotal') }}</span><span>{{ formatPrice(order.subtotal_amount) }}</span></div>
            <div class="flex justify-between"><span class="text-gray-600">{{ t('common.tax') }}</span><span>{{ formatPrice(order.tax_amount) }}</span></div>
            <div v-if="Number(order.points_applied) > 0" class="flex justify-between"><span class="text-gray-600">{{ t('common.discount') }}</span><span>-{{ formatPrice(order.discount_amount) }}</span></div>
            <div class="border-t pt-2 flex justify-between font-semibold">
              <span>{{ t('common.total') }}</span>
              <span>{{ formatPrice(order.total_amount) }}</span>
            </div>
          </div>
        </div>
      </div>
    </template>

    <AppModal :open="showCancelModal" :message="t('order.cancelConfirm')" :danger="true"
      :confirm-label="t('order.cancelOrder')" @close="showCancelModal = false" @confirm="handleCancel" />
  </div>
</template>
