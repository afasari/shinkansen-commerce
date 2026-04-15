<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { Address, ShippingAddress } from '@/types'

const { t } = useI18n() as any

const props = defineProps<{
  addresses: Address[]
  selectedId: string | null
}>()

const emit = defineEmits<{
  (e: 'select', address: Address): void
}>()
</script>

<template>
  <div class="space-y-3">
    <button v-for="addr in addresses" :key="addr.id" @click="emit('select', addr)"
      :class="[
        selectedId === addr.id ? 'ring-2 ring-shinkansen-600 bg-shinkansen-50' : 'ring-1 ring-gray-200 hover:ring-shinkansen-300',
        'w-full text-left rounded-lg p-4 transition-colors'
      ]">
      <div class="flex items-center gap-2">
        <span class="font-medium text-sm text-gray-900">{{ addr.name }}</span>
        <span v-if="addr.is_default" class="px-2 py-0.5 text-xs bg-shinkansen-100 text-shinkansen-700 rounded-full">{{ t('address.default') }}</span>
      </div>
      <p class="text-sm text-gray-600 mt-1">{{ addr.phone }}</p>
      <p class="text-sm text-gray-600">{{ addr.postal_code }} {{ addr.prefecture }} {{ addr.city }} {{ addr.address_line1 }}</p>
      <p v-if="addr.address_line2" class="text-sm text-gray-600">{{ addr.address_line2 }}</p>
    </button>
  </div>
</template>
