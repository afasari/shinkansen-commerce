import { client } from './client'
import type {
  Order,
  CreateOrderRequest,
  CreateOrderResponse,
  ListOrdersParams,
  ListOrdersResponse,
  OrderStatus,
} from '@/types'

export async function listOrders(params?: ListOrdersParams): Promise<ListOrdersResponse> {
  const res = await client.get<ListOrdersResponse>('/v1/orders', { params })
  return res.data
}

export async function getOrder(orderId: string): Promise<Order> {
  const res = await client.get<{ order: Order }>(`/v1/orders/${orderId}`)
  return res.data.order
}

export async function createOrder(data: CreateOrderRequest): Promise<CreateOrderResponse> {
  const res = await client.post<CreateOrderResponse>('/v1/orders', data)
  return res.data
}

export async function updateOrderStatus(orderId: string, status: OrderStatus, reason?: string): Promise<void> {
  await client.post(`/v1/orders/${orderId}/status`, { status, reason })
}

export async function cancelOrder(orderId: string, reason?: string): Promise<void> {
  await client.post(`/v1/orders/${orderId}/cancel`, { reason })
}

export async function applyPoints(orderId: string, points: number): Promise<{ success: boolean; yen_value: { currency: string; units: number; nanos: number } }> {
  const res = await client.post('/v1/orders/apply-points', { order_id: orderId, points })
  return res.data
}
