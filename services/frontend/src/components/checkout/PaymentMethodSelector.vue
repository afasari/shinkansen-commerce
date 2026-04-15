<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { PaymentMethod } from '@/types'

const { t } = useI18n() as any

const props = defineProps<{
  modelValue: PaymentMethod | null
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', method: PaymentMethod): void
}>()

const methods = [
  { value: PaymentMethod.CREDIT_CARD, label: 'checkout.creditCard', icon: '💳' },
  { value: PaymentMethod.KONBINI_SEVENELEVEN, label: 'checkout.konbiniSevenEleven', icon: '🏪' },
  { value: PaymentMethod.KONBINI_LAWSON, label: 'checkout.konbiniLawson', icon: '🏪' },
  { value: PaymentMethod.KONBINI_FAMILYMART, label: 'checkout.konbiniFamilyMart', icon: '🏪' },
  { value: PaymentMethod.PAYPAY, label: 'checkout.paypay', icon: '📱' },
  { value: PaymentMethod.RAKUTEN_PAY, label: 'checkout.rakutenPay', icon: '📱' },
]
</script>

<template>
  <div class="grid grid-cols-2 sm:grid-cols-3 gap-3">
    <button v-for="m in methods" :key="m.value" @click="emit('update:modelValue', m.value)"
      :class="[
        modelValue === m.value ? 'ring-2 ring-shinkansen-600 bg-shinkansen-50' : 'ring-1 ring-gray-200 hover:ring-shinkansen-300',
        'rounded-lg p-4 text-left transition-colors'
      ]">
      <span class="text-lg mb-1 block">{{ m.icon }}</span>
      <p class="text-sm font-medium text-gray-900">{{ t(m.label) }}</p>
    </button>
  </div>
</template>
