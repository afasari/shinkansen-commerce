import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { CartItem, Money } from '@/types'
import { createMoney } from '@/utils/format'
import { generateSessionId } from '@/utils/constants'

export const useCartStore = defineStore('cart', () => {
  const items = ref<CartItem[]>([])
  const sessionId = ref<string>(localStorage.getItem('cart_session_id') || generateSessionId())

  const itemCount = computed(() => items.value.reduce((sum, item) => sum + item.quantity, 0))

  const subtotal = computed<Money>(() => {
    const totalUnits = items.value.reduce((sum, item) => {
      return sum + (item.unit_price.units * item.quantity)
    }, 0)
    return createMoney(totalUnits)
  })

  function initSession() {
    if (!localStorage.getItem('cart_session_id')) {
      localStorage.setItem('cart_session_id', sessionId.value)
    }
    const saved = localStorage.getItem('cart_items')
    if (saved) {
      try {
        items.value = JSON.parse(saved)
      } catch {
        items.value = []
      }
    }
  }

  function persist() {
    localStorage.setItem('cart_items', JSON.stringify(items.value))
    localStorage.setItem('cart_session_id', sessionId.value)
  }

  function addItem(item: CartItem) {
    const existing = items.value.find(
      (i) => i.product_id === item.product_id && i.variant_id === item.variant_id,
    )
    if (existing) {
      existing.quantity += item.quantity
    } else {
      items.value.push({ ...item })
    }
    persist()
  }

  function updateQuantity(productId: string, variantId: string, quantity: number) {
    const item = items.value.find(
      (i) => i.product_id === productId && i.variant_id === variantId,
    )
    if (item) {
      item.quantity = Math.max(0, quantity)
      if (item.quantity === 0) {
        removeItem(productId, variantId)
      } else {
        persist()
      }
    }
  }

  function removeItem(productId: string, variantId: string) {
    items.value = items.value.filter(
      (i) => !(i.product_id === productId && i.variant_id === variantId),
    )
    persist()
  }

  function clearCart() {
    items.value = []
    persist()
  }

  function mergeCart(newItems: CartItem[]) {
    for (const item of newItems) {
      const existing = items.value.find(
        (i) => i.product_id === item.product_id && i.variant_id === item.variant_id,
      )
      if (existing) {
        existing.quantity += item.quantity
      } else {
        items.value.push({ ...item })
      }
    }
    persist()
  }

  return {
    items, sessionId, itemCount, subtotal,
    initSession, addItem, updateQuantity, removeItem, clearCart, mergeCart,
  }
})
