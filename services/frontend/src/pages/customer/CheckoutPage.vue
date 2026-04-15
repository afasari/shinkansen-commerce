<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useCheckoutStore } from '@/stores/checkout'
import { useAddressStore } from '@/stores/address'
import { useDeliveryStore } from '@/stores/delivery'
import { useCartStore } from '@/stores/cart'
import { formatPrice, formatDate, formatTime } from '@/utils/format'
import { PaymentMethod, type Address, type ShippingAddress, type DeliverySlot } from '@/types'
import { PREFECTURES, DEFAULT_DELIVERY_ZONE_ID } from '@/utils/constants'

const { t } = useI18n()
const router = useRouter()
const checkout = useCheckoutStore()
const addressStore = useAddressStore()
const deliveryStore = useDeliveryStore()
const cartStore = useCartStore()

const deliveryDate = ref(new Date().toISOString().split('T')[0])
const deliveryZoneId = ref(DEFAULT_DELIVERY_ZONE_ID)
const selectedAddressRef = ref<Address | null>(null)
const selectedSlotRef = ref<DeliverySlot | null>(null)

const cardNumber = ref('')
const cardExpiry = ref('')
const cardCvv = ref('')

onMounted(async () => {
  if (cartStore.items.length === 0) {
    router.replace('/cart')
    return
  }
  await addressStore.fetchAddresses()
})

const steps = [
  { num: 1, label: t('checkout.step1') },
  { num: 2, label: t('checkout.step2') },
  { num: 3, label: t('checkout.step3') },
  { num: 4, label: t('checkout.step4') },
]

const paymentMethods = [
  { value: PaymentMethod.CREDIT_CARD, label: t('checkout.creditCard') },
  { value: PaymentMethod.KONBINI_SEVENELEVEN, label: t('checkout.konbiniSevenEleven') },
  { value: PaymentMethod.KONBINI_LAWSON, label: t('checkout.konbiniLawson') },
  { value: PaymentMethod.KONBINI_FAMILYMART, label: t('checkout.konbiniFamilyMart') },
  { value: PaymentMethod.PAYPAY, label: t('checkout.paypay') },
  { value: PaymentMethod.RAKUTEN_PAY, label: t('checkout.rakutenPay') },
]

function selectExistingAddress(addr: Address) {
  selectedAddressRef.value = addr
  const shippingAddr: ShippingAddress = {
    name: addr.name,
    phone: addr.phone,
    postal_code: addr.postal_code,
    prefecture: addr.prefecture,
    city: addr.city,
    address_line1: addr.address_line1,
    address_line2: addr.address_line2,
  }
  checkout.setAddress(shippingAddr, addr.id)
}

async function goToStep2() {
  if (selectedAddressRef.value) {
    await deliveryStore.fetchSlots(deliveryZoneId.value, deliveryDate.value)
    checkout.step = 2
  }
}

function selectSlot(slot: DeliverySlot) {
  selectedSlotRef.value = slot
  checkout.setDeliverySlot(slot)
}

function goToStep4() {
  if (checkout.selectedPaymentMethod) {
    checkout.step = 4
  }
}

async function handlePlaceOrder() {
  const res = await checkout.placeOrder()
  if (res) {
    router.push({ name: 'checkout-success' })
  }
}

const isCreditCard = computed(() => checkout.selectedPaymentMethod === PaymentMethod.CREDIT_CARD)
</script>

<template>
  <div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <h1 class="text-2xl font-bold text-gray-900 mb-6">{{ t('checkout.title') }}</h1>

    <nav class="flex items-center mb-8">
      <template v-for="(step, idx) in steps" :key="step.num">
        <div class="flex items-center">
          <div :class="[
            checkout.step >= step.num ? 'bg-shinkansen-600 text-white' : 'bg-gray-200 text-gray-500',
            'h-8 w-8 rounded-full flex items-center justify-center text-sm font-semibold'
          ]">{{ step.num }}</div>
          <span :class="[checkout.step >= step.num ? 'text-shinkansen-600' : 'text-gray-400', 'ml-2 text-sm font-medium hidden sm:block']">{{ step.label }}</span>
        </div>
        <div v-if="idx < steps.length - 1" class="flex-1 h-0.5 mx-3" :class="checkout.step > step.num ? 'bg-shinkansen-600' : 'bg-gray-200'"></div>
      </template>
    </nav>

    <div v-if="checkout.error" class="rounded-md bg-red-50 p-3 mb-6">
      <p class="text-sm text-red-700">{{ checkout.error }}</p>
    </div>

    <!-- Step 1: Address -->
    <div v-if="checkout.step === 1" class="space-y-4">
      <h2 class="text-lg font-semibold">{{ t('checkout.selectAddress') }}</h2>
      <div v-if="addressStore.addresses.length > 0" class="space-y-3">
        <button v-for="addr in addressStore.addresses" :key="addr.id" @click="selectExistingAddress(addr)"
          :class="[selectedAddressRef?.id === addr.id ? 'ring-2 ring-shinkansen-600 bg-shinkansen-50' : 'ring-1 ring-gray-200 hover:ring-shinkansen-300', 'w-full text-left rounded-lg p-4']">
          <div class="flex items-center gap-2">
            <span class="font-medium text-sm">{{ addr.name }}</span>
            <span v-if="addr.is_default" class="px-2 py-0.5 text-xs bg-shinkansen-100 text-shinkansen-700 rounded-full">{{ t('address.default') }}</span>
          </div>
          <p class="text-sm text-gray-600 mt-1">{{ addr.postal_code }} {{ addr.prefecture }} {{ addr.city }} {{ addr.address_line1 }}</p>
          <p class="text-sm text-gray-500">{{ addr.phone }}</p>
        </button>
      </div>
      <router-link to="/account/addresses/new" class="btn-secondary text-sm">{{ t('checkout.addNewAddress') }}</router-link>

      <div class="pt-4 border-t">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="label-field">{{ t('checkout.deliveryZone') }}</label>
            <input v-model="deliveryZoneId" class="input-field mt-1 text-sm" />
          </div>
          <div>
            <label class="label-field">{{ t('checkout.deliveryDate') }}</label>
            <input v-model="deliveryDate" type="date" class="input-field mt-1 text-sm" />
          </div>
        </div>
      </div>

      <div class="flex justify-end">
        <button @click="goToStep2" :disabled="!selectedAddressRef" class="btn-primary">{{ t('common.next') }}</button>
      </div>
    </div>

    <!-- Step 2: Delivery Slot -->
    <div v-if="checkout.step === 2" class="space-y-4">
      <h2 class="text-lg font-semibold">{{ t('checkout.selectDeliverySlot') }}</h2>

      <div v-if="deliveryStore.loading" class="text-center py-8">
        <div class="h-8 w-8 animate-spin rounded-full border-4 border-shinkansen-200 border-t-shinkansen-600 mx-auto"></div>
      </div>

      <div v-else-if="deliveryStore.slots.length === 0" class="text-center py-8 text-gray-500">
        {{ t('checkout.noSlotsAvailable') }}
      </div>

      <div v-else class="grid grid-cols-2 sm:grid-cols-3 gap-3">
        <button v-for="slot in deliveryStore.slots" :key="slot.id" @click="selectSlot(slot)"
          :class="[
            selectedSlotRef?.id === slot.id ? 'ring-2 ring-shinkansen-600 bg-shinkansen-50' : 'ring-1 ring-gray-200 hover:ring-shinkansen-300',
            'rounded-lg p-4 text-left'
          ]">
          <p class="text-sm font-medium">{{ formatTime(slot.start_time) }} - {{ formatTime(slot.end_time) }}</p>
          <p class="text-xs text-gray-500 mt-1">{{ slot.available }} / {{ slot.capacity }} {{ t('common.available') }}</p>
        </button>
      </div>

      <div class="flex justify-between">
        <button @click="checkout.step = 1" class="btn-secondary">{{ t('common.back') }}</button>
        <button @click="checkout.step = 3" :disabled="!selectedSlotRef" class="btn-primary">{{ t('common.next') }}</button>
      </div>
    </div>

    <!-- Step 3: Payment -->
    <div v-if="checkout.step === 3" class="space-y-4">
      <h2 class="text-lg font-semibold">{{ t('checkout.selectPaymentMethod') }}</h2>
      <div class="grid grid-cols-2 sm:grid-cols-3 gap-3">
        <button v-for="pm in paymentMethods" :key="pm.value" @click="checkout.setPaymentMethod(pm.value)"
          :class="[
            checkout.selectedPaymentMethod === pm.value ? 'ring-2 ring-shinkansen-600 bg-shinkansen-50' : 'ring-1 ring-gray-200 hover:ring-shinkansen-300',
            'rounded-lg p-4 text-left'
          ]">
          <p class="text-sm font-medium text-gray-900">{{ pm.label }}</p>
        </button>
      </div>

      <div v-if="isCreditCard" class="mt-4 space-y-3">
        <div>
          <label class="label-field">{{ t('checkout.cardNumber') }}</label>
          <input v-model="cardNumber" class="input-field mt-1" placeholder="4242 4242 4242 4242" />
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="label-field">{{ t('checkout.expiry') }}</label>
            <input v-model="cardExpiry" class="input-field mt-1" placeholder="MM/YY" />
          </div>
          <div>
            <label class="label-field">{{ t('checkout.cvv') }}</label>
            <input v-model="cardCvv" class="input-field mt-1" placeholder="123" />
          </div>
        </div>
      </div>

      <div class="pt-4 border-t">
        <label class="label-field">{{ t('checkout.applyPoints') }}</label>
        <input v-model.number="checkout.pointsToApply" type="number" min="0" class="input-field mt-1" />
        <p class="text-xs text-gray-500 mt-1">{{ t('checkout.pointsValue', { value: Math.floor(checkout.pointsToApply / 10) }) }}</p>
      </div>

      <div class="flex justify-between">
        <button @click="checkout.step = 2" class="btn-secondary">{{ t('common.back') }}</button>
        <button @click="goToStep4" :disabled="!checkout.selectedPaymentMethod" class="btn-primary">{{ t('common.next') }}</button>
      </div>
    </div>

    <!-- Step 4: Review -->
    <div v-if="checkout.step === 4" class="space-y-6">
      <h2 class="text-lg font-semibold">{{ t('checkout.step4') }}</h2>

      <div class="card p-4">
        <h3 class="text-sm font-semibold text-gray-900 mb-2">{{ t('checkout.shippingInfo') }}</h3>
        <p class="text-sm text-gray-600">{{ checkout.selectedAddress?.name }}</p>
        <p class="text-sm text-gray-600">{{ checkout.selectedAddress?.postal_code }} {{ checkout.selectedAddress?.prefecture }} {{ checkout.selectedAddress?.city }}</p>
        <p class="text-sm text-gray-600">{{ checkout.selectedAddress?.address_line1 }} {{ checkout.selectedAddress?.address_line2 }}</p>
      </div>

      <div class="card p-4">
        <h3 class="text-sm font-semibold text-gray-900 mb-2">{{ t('common.items') }}</h3>
        <div v-for="item in cartStore.items" :key="`${item.product_id}-${item.variant_id}`" class="flex justify-between py-2 text-sm">
          <span>{{ item.product_name }} x{{ item.quantity }}</span>
          <span class="font-medium">{{ formatPrice({ currency: item.unit_price.currency, units: item.unit_price.units * item.quantity, nanos: 0 }) }}</span>
        </div>
        <div class="border-t pt-2 mt-2 flex justify-between font-semibold">
          <span>{{ t('common.total') }}</span>
          <span>{{ formatPrice(cartStore.subtotal) }}</span>
        </div>
      </div>

      <div class="flex justify-between">
        <button @click="checkout.step = 3" class="btn-secondary">{{ t('common.back') }}</button>
        <button @click="handlePlaceOrder" :disabled="checkout.loading" class="btn-primary">
          <span v-if="checkout.loading" class="animate-spin mr-2 inline-block h-4 w-4 border-2 border-white border-t-transparent rounded-full"></span>
          {{ t('checkout.placeOrder') }}
        </button>
      </div>
    </div>
  </div>
</template>
