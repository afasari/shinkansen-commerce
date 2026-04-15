import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { ShippingAddress, PaymentMethod, DeliverySlot } from '@/types'
import { useCartStore } from './cart'
import { useOrderStore } from './order'
import { createMoney } from '@/utils/format'

export const useCheckoutStore = defineStore('checkout', () => {
  const step = ref(1)
  const selectedAddress = ref<ShippingAddress | null>(null)
  const selectedAddressId = ref<string | null>(null)
  const selectedDeliverySlot = ref<DeliverySlot | null>(null)
  const selectedPaymentMethod = ref<PaymentMethod | null>(null)
  const pointsToApply = ref(0)
  const lastOrderId = ref<string | null>(null)
  const lastOrderNumber = ref<string | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  function setAddress(addr: ShippingAddress, addressId: string) {
    selectedAddress.value = addr
    selectedAddressId.value = addressId
    step.value = 2
  }

  function setDeliverySlot(slot: DeliverySlot) {
    selectedDeliverySlot.value = slot
    step.value = 3
  }

  function setPaymentMethod(method: PaymentMethod) {
    selectedPaymentMethod.value = method
    step.value = 4
  }

  function setPoints(points: number) {
    pointsToApply.value = points
  }

  async function placeOrder() {
    const cartStore = useCartStore()
    const orderStore = useOrderStore()

    if (!selectedAddress.value || !selectedPaymentMethod.value) {
      error.value = 'Missing shipping address or payment method'
      return null
    }

    loading.value = true
    error.value = null

    try {
      const items = cartStore.items.map((item) => ({
        product_id: item.product_id,
        variant_id: item.variant_id,
        quantity: item.quantity,
      }))

      const userId = localStorage.getItem('user_id') || ''

      const res = await orderStore.createOrder({
        user_id: userId,
        items,
        shipping_address: selectedAddress.value,
        payment_method: selectedPaymentMethod.value,
        points_to_apply: pointsToApply.value > 0 ? String(pointsToApply.value) : undefined,
        delivery_slot_id: selectedDeliverySlot.value?.id || undefined,
      })

      lastOrderId.value = res.order_id
      lastOrderNumber.value = res.order_number
      cartStore.clearCart()
      return res
    } catch (e: unknown) {
      error.value = (e as Error).message
      return null
    } finally {
      loading.value = false
    }
  }

  function reset() {
    step.value = 1
    selectedAddress.value = null
    selectedAddressId.value = null
    selectedDeliverySlot.value = null
    selectedPaymentMethod.value = null
    pointsToApply.value = 0
    lastOrderId.value = null
    lastOrderNumber.value = null
    error.value = null
  }

  return {
    step, selectedAddress, selectedAddressId, selectedDeliverySlot,
    selectedPaymentMethod, pointsToApply, lastOrderId, lastOrderNumber,
    loading, error,
    setAddress, setDeliverySlot, setPaymentMethod, setPoints, placeOrder, reset,
  }
})
