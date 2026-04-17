<script setup lang="ts">
import { computed } from 'vue'
import { OrderStatus, PaymentStatus, ShipmentStatus } from '@/types'
import { ORDER_STATUS_COLORS, PAYMENT_STATUS_COLORS, SHIPMENT_STATUS_COLORS, ORDER_STATUS_LABELS, PAYMENT_STATUS_LABELS, SHIPMENT_STATUS_LABELS } from '@/utils/constants'
import { useI18n } from 'vue-i18n'

const props = defineProps<{
  status: string | number
  type?: 'order' | 'payment' | 'shipment'
  size?: 'sm' | 'md'
}>()

const { t } = useI18n()

const detectedType = computed(() => {
  if (props.type) return props.type
  const n = typeof props.status === 'number' ? props.status : Number(props.status)
  if (n in ORDER_STATUS_COLORS) return 'order'
  if (n in PAYMENT_STATUS_COLORS) return 'payment'
  if (n in SHIPMENT_STATUS_COLORS) return 'shipment'
  return 'order'
})

const colors = computed(() => {
  const n = typeof props.status === 'number' ? props.status : Number(props.status)
  const t = detectedType.value
  if (t === 'payment') return PAYMENT_STATUS_COLORS[n] || 'bg-gray-100 text-gray-800'
  if (t === 'shipment') return SHIPMENT_STATUS_COLORS[n] || 'bg-gray-100 text-gray-800'
  return ORDER_STATUS_COLORS[n] || 'bg-gray-100 text-gray-800'
})

const label = computed(() => {
  const n = typeof props.status === 'number' ? props.status : Number(props.status)
  const t = detectedType.value
  let key: string | undefined
  if (t === 'payment') key = PAYMENT_STATUS_LABELS[n]
  else if (t === 'shipment') key = SHIPMENT_STATUS_LABELS[n]
  else key = ORDER_STATUS_LABELS[n]
  return key || String(props.status)
})

const sizeClass = computed(() => props.size === 'sm' ? 'px-2 py-0.5 text-xs' : 'px-3 py-1 text-sm')
</script>

<template>
  <span :class="[colors, sizeClass, 'inline-flex items-center rounded-full font-medium']">
    {{ label }}
  </span>
</template>
