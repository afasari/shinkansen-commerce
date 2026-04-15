<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useDeliveryStore } from '@/stores/delivery'
import { DEFAULT_DELIVERY_ZONE_ID } from '@/utils/constants'
import { formatTime } from '@/utils/format'

const { t } = useI18n()
const deliveryStore = useDeliveryStore()

const deliveryZoneId = ref(DEFAULT_DELIVERY_ZONE_ID)
const date = ref(new Date().toISOString().split('T')[0])
const loading = ref(false)

async function loadSlots() {
  loading.value = true
  await deliveryStore.fetchSlots(deliveryZoneId.value, date.value)
  loading.value = false
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-6">{{ t('nav.delivery') }} - {{ t('nav.management') }}</h1>

    <div class="card p-4 mb-6">
      <form @submit.prevent="loadSlots" class="flex gap-3 items-end">
        <div class="flex-1">
          <label class="label-field">{{ t('checkout.deliveryZone') }} ID</label>
          <input v-model="deliveryZoneId" required class="input-field mt-1" />
        </div>
        <div>
          <label class="label-field">{{ t('checkout.deliveryDate') }}</label>
          <input v-model="date" type="date" class="input-field mt-1" />
        </div>
        <button type="submit" :disabled="loading" class="btn-primary text-sm">{{ t('common.search') }}</button>
      </form>
    </div>

    <div v-if="loading || deliveryStore.loading" class="text-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-shinkansen-200 border-t-shinkansen-600 mx-auto"></div>
    </div>

    <div v-else-if="deliveryStore.slots.length > 0" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      <div v-for="slot in deliveryStore.slots" :key="slot.id" class="card p-4">
        <div class="flex justify-between items-start">
          <div>
            <p class="text-sm font-semibold text-gray-900">{{ formatTime(slot.start_time) }} - {{ formatTime(slot.end_time) }}</p>
            <p class="text-xs text-gray-500 mt-1">Zone: {{ slot.delivery_zone_id.substring(0, 8) }}...</p>
          </div>
          <div class="text-right">
            <p class="text-sm font-medium text-green-600">{{ slot.available }} {{ t('inventory.available') }}</p>
            <p class="text-xs text-gray-500">{{ slot.reserved }}/{{ slot.capacity }}</p>
          </div>
        </div>
        <div class="mt-2">
          <div class="w-full bg-gray-200 rounded-full h-1.5">
            <div class="bg-shinkansen-600 h-1.5 rounded-full" :style="{ width: `${(slot.reserved / slot.capacity) * 100}%` }"></div>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="text-center py-12 text-gray-500">
      {{ t('checkout.noSlotsAvailable') }}
    </div>
  </div>
</template>
