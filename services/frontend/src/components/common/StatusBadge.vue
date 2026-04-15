<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  status: string
  size?: 'sm' | 'md'
}>()

const colorMap: Record<string, string> = {
  ORDER_STATUS_PENDING: 'bg-yellow-100 text-yellow-800',
  ORDER_STATUS_CONFIRMED: 'bg-blue-100 text-blue-800',
  ORDER_STATUS_PROCESSING: 'bg-indigo-100 text-indigo-800',
  ORDER_STATUS_SHIPPED: 'bg-purple-100 text-purple-800',
  ORDER_STATUS_IN_TRANSIT: 'bg-violet-100 text-violet-800',
  ORDER_STATUS_DELIVERED: 'bg-green-100 text-green-800',
  ORDER_STATUS_CANCELLED: 'bg-red-100 text-red-800',
  ORDER_STATUS_EXPIRED: 'bg-gray-100 text-gray-800',
  ORDER_STATUS_READY_FOR_PICKUP: 'bg-teal-100 text-teal-800',
  ORDER_STATUS_PICKED_UP: 'bg-emerald-100 text-emerald-800',
  ORDER_STATUS_FAILED_DELIVERY: 'bg-orange-100 text-orange-800',
  ORDER_STATUS_RETURNED: 'bg-pink-100 text-pink-800',
  PAYMENT_STATUS_PENDING: 'bg-yellow-100 text-yellow-800',
  PAYMENT_STATUS_PROCESSING: 'bg-blue-100 text-blue-800',
  PAYMENT_STATUS_COMPLETED: 'bg-green-100 text-green-800',
  PAYMENT_STATUS_FAILED: 'bg-red-100 text-red-800',
  PAYMENT_STATUS_CANCELLED: 'bg-gray-100 text-gray-800',
  PAYMENT_STATUS_REFUNDED: 'bg-purple-100 text-purple-800',
  SHIPMENT_STATUS_PREPARING: 'bg-yellow-100 text-yellow-800',
  SHIPMENT_STATUS_SHIPPED: 'bg-blue-100 text-blue-800',
  SHIPMENT_STATUS_IN_TRANSIT: 'bg-purple-100 text-purple-800',
  SHIPMENT_STATUS_DELIVERED: 'bg-green-100 text-green-800',
  SHIPMENT_STATUS_CANCELLED: 'bg-red-100 text-red-800',
  SHIPMENT_STATUS_FAILED_DELIVERY: 'bg-orange-100 text-orange-800',
}

const statusLabelMap: Record<string, string> = {
  ORDER_STATUS_PENDING: 'order.pending',
  ORDER_STATUS_CONFIRMED: 'order.confirmed',
  ORDER_STATUS_PROCESSING: 'order.processing',
  ORDER_STATUS_SHIPPED: 'order.shipped',
  ORDER_STATUS_IN_TRANSIT: 'order.inTransit',
  ORDER_STATUS_DELIVERED: 'order.delivered',
  ORDER_STATUS_CANCELLED: 'order.cancelled',
  ORDER_STATUS_EXPIRED: 'order.expired',
  ORDER_STATUS_READY_FOR_PICKUP: 'order.readyForPickup',
  ORDER_STATUS_PICKED_UP: 'order.pickedUp',
  ORDER_STATUS_FAILED_DELIVERY: 'order.failedDelivery',
  ORDER_STATUS_RETURNED: 'order.returned',
  PAYMENT_STATUS_PENDING: 'payment.pending',
  PAYMENT_STATUS_PROCESSING: 'payment.processing',
  PAYMENT_STATUS_COMPLETED: 'payment.completed',
  PAYMENT_STATUS_FAILED: 'payment.failed',
  PAYMENT_STATUS_CANCELLED: 'payment.cancelled',
  PAYMENT_STATUS_REFUNDED: 'payment.refunded',
  SHIPMENT_STATUS_PREPARING: 'shipment.preparing',
  SHIPMENT_STATUS_SHIPPED: 'shipment.shipped',
  SHIPMENT_STATUS_IN_TRANSIT: 'shipment.inTransit',
  SHIPMENT_STATUS_DELIVERED: 'shipment.delivered',
  SHIPMENT_STATUS_CANCELLED: 'order.cancelled',
  SHIPMENT_STATUS_FAILED_DELIVERY: 'shipment.failedDelivery',
}

const colors = computed(() => colorMap[props.status] || 'bg-gray-100 text-gray-800')
const label = computed(() => statusLabelMap[props.status] ? (window as any).__vueI18n?.global?.t(statusLabelMap[props.status]) || props.status.replace(/_/g, ' ').replace(/\w+STATUS\s*/, '') : props.status)
const sizeClass = computed(() => props.size === 'sm' ? 'px-2 py-0.5 text-xs' : 'px-3 py-1 text-sm')
</script>

<template>
  <span :class="[colors, sizeClass, 'inline-flex items-center rounded-full font-medium']">
    {{ label }}
  </span>
</template>
