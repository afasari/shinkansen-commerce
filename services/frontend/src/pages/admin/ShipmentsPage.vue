<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useDeliveryStore } from '@/stores/delivery'
import { ShipmentStatus, SHIPMENT_STATUS_LIST } from '@/utils/constants'
import { formatDateTime } from '@/utils/format'
import StatusBadge from '@/components/common/StatusBadge.vue'

const { t } = useI18n()
const deliveryStore = useDeliveryStore()

const shipmentId = ref('')
const statusUpdate = ref<ShipmentStatus | ''>('')
const statusDescription = ref('')
const updating = ref(false)

async function lookupShipment() {
  if (!shipmentId.value) return
  await deliveryStore.fetchShipment(shipmentId.value)
  statusUpdate.value = ''
  statusDescription.value = ''
}

async function handleStatusUpdate() {
  if (!statusUpdate.value || !shipmentId.value) return
  updating.value = true
  try {
    await deliveryStore.updateShipmentStatus(shipmentId.value, statusUpdate.value as ShipmentStatus, statusDescription.value)
    await lookupShipment()
    statusUpdate.value = ''
    statusDescription.value = ''
  } catch (e: unknown) {
    alert((e as Error).message)
  } finally {
    updating.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-6">{{ t('shipment.title') }} - {{ t('nav.management') }}</h1>

    <div class="card p-4 mb-6">
      <form @submit.prevent="lookupShipment" class="flex gap-3 items-end">
        <div class="flex-1">
          <label class="label-field">Shipment ID</label>
          <input v-model="shipmentId" required class="input-field mt-1" />
        </div>
        <button type="submit" class="btn-primary text-sm">{{ t('common.search') }}</button>
      </form>
    </div>

    <div v-if="deliveryStore.loading" class="text-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-shinkansen-200 border-t-shinkansen-600 mx-auto"></div>
    </div>

    <template v-if="deliveryStore.currentShipment">
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div class="card p-6">
          <h2 class="font-semibold text-gray-900 mb-4">{{ t('shipment.title') }}</h2>
          <div class="space-y-2 text-sm">
            <div class="flex justify-between">
              <span class="text-gray-600">ID</span>
              <span class="font-mono text-xs">{{ deliveryStore.currentShipment.id }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-600">{{ t('order.orderNumber') }}</span>
              <span>{{ deliveryStore.currentShipment.order_id.substring(0, 8) }}...</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-600">{{ t('shipment.trackingNumber') }}</span>
              <span>{{ deliveryStore.currentShipment.tracking_number || '-' }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-gray-600">{{ t('common.status') }}</span>
              <StatusBadge :status="deliveryStore.currentShipment.status" size="sm" />
            </div>
            <div class="flex justify-between">
              <span class="text-gray-600">{{ t('shipment.carrier') }}</span>
              <span>{{ deliveryStore.currentShipment.carrier || '-' }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-600">{{ t('shipment.estimatedDelivery') }}</span>
              <span>{{ formatDateTime(deliveryStore.currentShipment.estimated_delivery_at) }}</span>
            </div>
          </div>
        </div>

        <div class="card p-6">
          <h2 class="font-semibold text-gray-900 mb-4">{{ t('shipment.updateStatus') }}</h2>
          <form @submit.prevent="handleStatusUpdate" class="space-y-3">
            <div>
              <label class="label-field">{{ t('common.status') }}</label>
              <select v-model="statusUpdate" class="input-field mt-1">
                <option value="" disabled>Select status...</option>
                <option v-for="s in SHIPMENT_STATUS_LIST" :key="s" :value="s">{{ s.replace('SHIPMENT_STATUS_', '').replace(/_/g, ' ') }}</option>
              </select>
            </div>
            <div>
              <label class="label-field">{{ t('shipment.description') }}</label>
              <input v-model="statusDescription" class="input-field mt-1" />
            </div>
            <button type="submit" :disabled="updating || !statusUpdate" class="btn-primary text-sm">{{ t('common.save') }}</button>
          </form>

          <div v-if="deliveryStore.currentShipment.tracking_events?.length" class="mt-6 pt-4 border-t">
            <h3 class="text-sm font-semibold text-gray-900 mb-2">{{ t('shipment.trackingEvents') }}</h3>
            <div class="space-y-2">
              <div v-for="event in deliveryStore.currentShipment.tracking_events" :key="event.id" class="text-sm">
                <div class="flex justify-between">
                  <span class="font-medium">{{ event.status }}</span>
                  <span class="text-gray-500 text-xs">{{ formatDateTime(event.timestamp) }}</span>
                </div>
                <p v-if="event.location" class="text-gray-500 text-xs">{{ event.location }}</p>
                <p v-if="event.description" class="text-gray-600 text-xs">{{ event.description }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
