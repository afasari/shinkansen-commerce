import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { ShippingAddress, DeliverySlot, Money } from '@/types'
import { OrderStatus, PaymentMethod, PaymentStatus } from '@/types'
import { useCartStore } from './cart'
import { useAuthStore } from './auth'
import { DEFAULT_WAREHOUSE_ID } from '@/utils/constants'
import { createMoney } from '@/utils/format'
import * as ordersApi from '@/api/orders'
import * as inventoryApi from '@/api/inventory'
import * as deliveryApi from '@/api/delivery'
import * as paymentsApi from '@/api/payments'

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
  const loadingStep = ref<string | null>(null)
  const error = ref<string | null>(null)

  const TAX_RATE = 0.10

  const subtotal = computed<Money>(() => {
    const cartStore = useCartStore()
    return cartStore.subtotal
  })

  const pointsDiscount = computed<number>(() => {
    return pointsToApply.value > 0 ? pointsToApply.value * 10 : 0
  })

  const tax = computed<number>(() => {
    const taxable = (subtotal.value?.units ?? 0) - pointsDiscount.value
    return Math.max(0, Math.round(taxable * TAX_RATE))
  })

  const total = computed<number>(() => {
    return Math.max(0, (subtotal.value?.units ?? 0) - pointsDiscount.value + tax.value)
  })

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

  async function placeOrder(cardData?: { cardNumber: string; cardExpiry: string; cardCvv: string }) {
    const cartStore = useCartStore()
    const authStore = useAuthStore()

    if (!selectedAddress.value || !selectedPaymentMethod.value) {
      error.value = 'Missing shipping address or payment method'
      return null
    }

    if (cartStore.items.length === 0) {
      error.value = 'Cart is empty'
      return null
    }

    loading.value = true
    error.value = null

    let orderId: string | null = null
    let orderNumber: string | null = null
    let reservationId: string | null = null
    let finalStatus = OrderStatus.PENDING

    try {
      const userId = authStore.userId || ''
      const items = cartStore.items.map((item) => ({
        product_id: item.product_id,
        variant_id: item.variant_id,
        quantity: item.quantity,
      }))

      loadingStep.value = 'Creating order...'
      const orderRes = await ordersApi.createOrder({
        user_id: userId,
        items,
        shipping_address: selectedAddress.value,
        payment_method: selectedPaymentMethod.value,
        points_to_apply: pointsToApply.value > 0 ? String(pointsToApply.value) : undefined,
        delivery_slot_id: selectedDeliverySlot.value?.id || undefined,
      })
      orderId = orderRes.order_id
      orderNumber = orderRes.order_number

      loadingStep.value = 'Reserving inventory...'
      try {
        const reserveRes = await inventoryApi.reserveStock({
          order_id: orderId,
          items: cartStore.items.map((item) => ({
            product_id: item.product_id,
            variant_id: item.variant_id || '',
            warehouse_id: DEFAULT_WAREHOUSE_ID,
            quantity: item.quantity,
          })),
          expires_at: new Date(Date.now() + 30 * 60 * 1000).toISOString(),
        })
        reservationId = reserveRes.reservation_id
        if (reserveRes.success === false || (reserveRes.failed_items && reserveRes.failed_items.length > 0)) {
          await rollbackOrder(orderId, reservationId, null)
          error.value = `Some items are out of stock: ${(reserveRes.failed_items || []).join(', ')}`
          return null
        }
      } catch (invErr) {
        await rollbackOrder(orderId, null, null)
        error.value = `Inventory reservation failed: ${(invErr as Error).message}`
        return null
      }

      if (selectedDeliverySlot.value) {
        loadingStep.value = 'Reserving delivery slot...'
        try {
          await deliveryApi.reserveDeliverySlotByPath(
            selectedDeliverySlot.value.id,
            orderId,
          )
        } catch (delErr) {
          await rollbackOrder(orderId, reservationId, null)
          error.value = `Delivery slot reservation failed: ${(delErr as Error).message}`
          return null
        }
      }

      loadingStep.value = 'Creating payment...'
      let paymentId: string | null = null
      try {
        const paymentRes = await paymentsApi.createPayment({
          order_id: orderId,
          method: selectedPaymentMethod.value,
          amount: { currency: 'JPY', units: total.value, nanos: 0 },
        })
        paymentId = paymentRes.payment_id
      } catch (payErr) {
        await rollbackOrder(orderId, reservationId, null)
        error.value = `Payment creation failed: ${(payErr as Error).message}`
        return null
      }

      loadingStep.value = 'Processing payment...'
      try {
        const paymentData: Record<string, string> = {}
        if (cardData && selectedPaymentMethod.value === PaymentMethod.CREDIT_CARD) {
          paymentData.card_number = cardData.cardNumber
          paymentData.expiry = cardData.cardExpiry
          paymentData.cvv = cardData.cardCvv
        }

        const processRes = await paymentsApi.processPayment(paymentId, { payment_data: paymentData })
        const paymentStatus = typeof processRes.status === 'number'
          ? processRes.status
          : paymentStatusFromString(processRes.status)

        if (paymentStatus === PaymentStatus.COMPLETED) {
          loadingStep.value = 'Confirming order...'
          finalStatus = OrderStatus.CONFIRMED
          try {
            await ordersApi.updateOrderStatus(orderId, OrderStatus.CONFIRMED)
          } catch {
          }
        } else if (paymentStatus === PaymentStatus.PROCESSING) {
          finalStatus = OrderStatus.PENDING
        } else if (paymentStatus === PaymentStatus.FAILED) {
          await rollbackOrder(orderId, reservationId, paymentId)
          error.value = 'Payment was declined. Please try a different payment method.'
          return null
        }
      } catch (procErr) {
        await rollbackOrder(orderId, reservationId, paymentId)
        error.value = `Payment processing failed: ${(procErr as Error).message}`
        return null
      }

      lastOrderId.value = orderId
      lastOrderNumber.value = orderNumber
      cartStore.clearCart()
      return { order_id: orderId, order_number: orderNumber, status: finalStatus }
    } catch (e: unknown) {
      if (orderId) {
        await rollbackOrder(orderId, reservationId, null)
      }
      error.value = (e as Error).message
      return null
    } finally {
      loading.value = false
      loadingStep.value = null
    }
  }

  async function rollbackOrder(orderId: string, reservationId: string | null, paymentId: string | null) {
    try {
      if (reservationId) {
        await inventoryApi.releaseStock({ reservation_id: reservationId })
      }
    } catch {
      // Best-effort rollback
    }
    try {
      await ordersApi.cancelOrder(orderId, 'Checkout flow failed')
    } catch {
      // Best-effort rollback
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
    loading.value = false
    loadingStep.value = null
    error.value = null
  }

  return {
    step, selectedAddress, selectedAddressId, selectedDeliverySlot,
    selectedPaymentMethod, pointsToApply, lastOrderId, lastOrderNumber,
    loading, loadingStep, error, subtotal, pointsDiscount, tax, total,
    setAddress, setDeliverySlot, setPaymentMethod, setPoints, placeOrder, reset,
  }
})

function paymentStatusFromString(s: string): PaymentStatus {
  const map: Record<string, PaymentStatus> = {
    PAYMENT_STATUS_UNSPECIFIED: PaymentStatus.UNSPECIFIED,
    PAYMENT_STATUS_PENDING: PaymentStatus.PENDING,
    PAYMENT_STATUS_PROCESSING: PaymentStatus.PROCESSING,
    PAYMENT_STATUS_COMPLETED: PaymentStatus.COMPLETED,
    PAYMENT_STATUS_FAILED: PaymentStatus.FAILED,
    PAYMENT_STATUS_CANCELLED: PaymentStatus.CANCELLED,
    PAYMENT_STATUS_REFUNDED: PaymentStatus.REFUNDED,
  }
  return map[s] ?? PaymentStatus.UNSPECIFIED
}
