<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { DeliverySlot } from '@/types'
import { formatTime } from '@/utils/format'

const { t } = useI18n() as any

const props = defineProps<{
  slots: DeliverySlot[]
  selectedId: string | null
}>()

const emit = defineEmits<{
  (e: 'select', slot: DeliverySlot): void
}>()
</script>

<template>
  <div v-if="slots.length === 0" class="text-center py-8 text-gray-500">
    {{ t('checkout.noSlotsAvailable') }}
  </div>
  <div v-else class="grid grid-cols-2 sm:grid-cols-3 gap-3">
    <button v-for="slot in slots" :key="slot.id" @click="emit('select', slot)"
      :class="[
        selectedId === slot.id ? 'ring-2 ring-shinkansen-600 bg-shinkansen-50' : 'ring-1 ring-gray-200 hover:ring-shinkansen-300',
        'rounded-lg p-4 text-left transition-colors'
      ]">
      <p class="text-sm font-medium text-gray-900">{{ formatTime(slot.start_time) }} - {{ formatTime(slot.end_time) }}</p>
      <p class="text-xs text-gray-500 mt-1">{{ slot.available }} / {{ slot.capacity }} {{ t('inventory.available') }}</p>
      <div class="mt-2 w-full bg-gray-200 rounded-full h-1">
        <div class="bg-shinkansen-600 h-1 rounded-full transition-all" :style="{ width: `${Math.min(100, (slot.reserved / slot.capacity) * 100)}%` }"></div>
      </div>
    </button>
  </div>
</template>
