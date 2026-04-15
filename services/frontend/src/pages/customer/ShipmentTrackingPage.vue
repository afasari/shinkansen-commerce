<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useOrderStore } from '@/stores/order'
import { useDeliveryStore } from '@/stores/delivery'
import { ShipmentStatus } from '@/types'
import { formatDateTime } from '@/utils/format'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { TruckIcon, MapPinIcon, CheckCircleIcon } from '@heroicons/vue/24/outline'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const orderStore = useOrderStore()
const deliveryStore = useDeliveryStore()

onMounted(async () => {
  await orderStore.fetchOrder(route.params.id as string)
})

const order = computed(() => orderStore.currentOrder)

async function loadShipment(shipmentId: string) {
  await deliveryStore.fetchShipment(shipmentId)
}
</script>

<template>
  <div class="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <button @click="router.back()" class="text-sm text-gray-500 hover:text-gray-700 mb-4">&larr; {{ t('common.back') }}</button>
    <h1 class="text-2xl font-bold text-gray-900 mb-6">{{ t('shipment.title') }}</h1>

    <div v-if="orderStore.loading" class="text-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-shinkansen-200 border-t-shinkansen-600 mx-auto"></div>
    </div>

    <template v-else-if="order">
      <div class="card p-6 mb-6">
        <div class="flex justify-between items-center mb-4">
          <h2 class="font-semibold">{{ order.order_number }}</h2>
          <StatusBadge :status="order.status" size="sm" />
        </div>

        <div v-if="order.estimated_delivery_at" class="text-sm text-gray-600">
          <span class="font-medium">{{ t('shipment.estimatedDelivery') }}:</span> {{ formatDateTime(order.estimated_delivery_at) }}
        </div>
      </div>

      <div v-if="order.status === 'ORDER_STATUS_PENDING' || order.status === 'ORDER_STATUS_CONFIRMED'" class="text-center py-12 text-gray-500">
        <p>{{ t('shipment.preparing') }}...</p>
      </div>

      <div v-else class="card p-6">
        <h3 class="font-semibold text-gray-900 mb-4">{{ t('shipment.trackingEvents') }}</h3>

        <div class="space-y-0">
          <div v-for="(event, idx) in [
            { status: t('shipment.preparing'), desc: t('shipment.preparing'), done: true },
            { status: t('shipment.shipped'), desc: t('shipment.shipped'), done: ['ORDER_STATUS_SHIPPED','ORDER_STATUS_IN_TRANSIT','ORDER_STATUS_DELIVERED'].includes(order.status) },
            { status: t('shipment.inTransit'), desc: t('shipment.inTransit'), done: ['ORDER_STATUS_IN_TRANSIT','ORDER_STATUS_DELIVERED'].includes(order.status) },
            { status: t('shipment.delivered'), desc: t('shipment.delivered'), done: order.status === 'ORDER_STATUS_DELIVERED' },
          ]" :key="idx" class="flex items-start gap-3">
            <div class="flex flex-col items-center">
              <div :class="[event.done ? 'bg-shinkansen-600 text-white' : 'bg-gray-200 text-gray-400', 'h-6 w-6 rounded-full flex items-center justify-center flex-shrink-0']">
                <CheckCircleIcon v-if="event.done" class="h-4 w-4" />
              </div>
              <div v-if="idx < 3" class="w-0.5 h-8" :class="event.done ? 'bg-shinkansen-600' : 'bg-gray-200'"></div>
            </div>
            <div class="pb-6">
              <p :class="[event.done ? 'text-gray-900 font-medium' : 'text-gray-400', 'text-sm']">{{ event.status }}</p>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
