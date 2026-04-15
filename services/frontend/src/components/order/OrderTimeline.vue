<script setup lang="ts">
import { OrderStatus, CANCELLABLE_STATUSES } from '@/utils/constants'
import StatusBadge from '@/components/common/StatusBadge.vue'

defineProps<{
  currentStatus: OrderStatus
}>()

const emit = defineEmits<{
  (e: 'cancel'): void
}>()

const timeline = [
  { status: OrderStatus.PENDING, label: 'order.pending' },
  { status: OrderStatus.CONFIRMED, label: 'order.confirmed' },
  { status: OrderStatus.PROCESSING, label: 'order.processing' },
  { status: OrderStatus.SHIPPED, label: 'order.shipped' },
  { status: OrderStatus.IN_TRANSIT, label: 'order.inTransit' },
  { status: OrderStatus.DELIVERED, label: 'order.delivered' },
]

const statusOrder = [
  OrderStatus.PENDING, OrderStatus.CONFIRMED, OrderStatus.PROCESSING,
  OrderStatus.SHIPPED, OrderStatus.IN_TRANSIT, OrderStatus.DELIVERED,
]

function isActive(status: OrderStatus, currentStatus: OrderStatus): boolean {
  return statusOrder.indexOf(status) <= statusOrder.indexOf(currentStatus)
}

function isTerminal(status: OrderStatus): boolean {
  return [OrderStatus.CANCELLED, OrderStatus.EXPIRED, OrderStatus.RETURNED, OrderStatus.FAILED_DELIVERY].includes(status)
}
</script>

<template>
  <div>
    <div v-if="isTerminal(currentStatus)" class="flex items-center gap-2 mb-4">
      <StatusBadge :status="currentStatus" />
    </div>

    <div v-else class="flex items-center gap-0 w-full">
      <template v-for="(step, idx) in timeline" :key="step.status">
        <div class="flex flex-col items-center flex-1">
          <div :class="[
            isActive(step.status, currentStatus) ? 'bg-shinkansen-600 text-white' : 'bg-gray-200 text-gray-400',
            'h-8 w-8 rounded-full flex items-center justify-center text-xs font-semibold'
          ]">{{ idx + 1 }}</div>
          <span :class="[isActive(step.status, currentStatus) ? 'text-shinkansen-600' : 'text-gray-400', 'text-xs mt-1 text-center']">
            {{ step.label.split('.').pop() }}
          </span>
        </div>
        <div v-if="idx < timeline.length - 1" class="h-0.5 flex-1 -mt-4"
          :class="isActive(timeline[idx + 1].status, currentStatus) ? 'bg-shinkansen-600' : 'bg-gray-200'"></div>
      </template>
    </div>
  </div>
</template>
