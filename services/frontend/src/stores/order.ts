import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Order, ListOrdersParams, ListOrdersResponse, OrderStatus, CreateOrderRequest, CreateOrderResponse } from '@/types'
import * as ordersApi from '@/api/orders'

export const useOrderStore = defineStore('order', () => {
  const orders = ref<Order[]>([])
  const currentOrder = ref<Order | null>(null)
  const pagination = ref({ page: 1, limit: 20, total: 0 })
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchOrders(params?: ListOrdersParams) {
    loading.value = true
    error.value = null
    try {
      const res = await ordersApi.listOrders(params)
      orders.value = res.orders
      pagination.value = res.pagination
    } catch (e: unknown) {
      error.value = (e as Error).message
    } finally {
      loading.value = false
    }
  }

  async function fetchOrder(orderId: string) {
    loading.value = true
    error.value = null
    try {
      currentOrder.value = await ordersApi.getOrder(orderId)
    } catch (e: unknown) {
      error.value = (e as Error).message
    } finally {
      loading.value = false
    }
  }

  async function createOrder(data: CreateOrderRequest): Promise<CreateOrderResponse> {
    const res = await ordersApi.createOrder(data)
    return res
  }

  async function updateStatus(orderId: string, status: OrderStatus, reason?: string) {
    await ordersApi.updateOrderStatus(orderId, status, reason)
    if (currentOrder.value && currentOrder.value.id === orderId) {
      currentOrder.value.status = status
    }
  }

  async function cancelOrder(orderId: string, reason?: string) {
    await ordersApi.cancelOrder(orderId, reason)
    if (currentOrder.value && currentOrder.value.id === orderId) {
      currentOrder.value.status = 'ORDER_STATUS_CANCELLED' as OrderStatus
    }
  }

  async function applyPoints(orderId: string, points: number) {
    return await ordersApi.applyPoints(orderId, points)
  }

  return {
    orders, currentOrder, pagination, loading, error,
    fetchOrders, fetchOrder, createOrder, updateStatus, cancelOrder, applyPoints,
  }
})
