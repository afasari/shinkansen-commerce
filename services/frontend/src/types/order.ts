import type { Money, Pagination } from './common'

export enum OrderStatus {
  UNSPECIFIED = 0,
  PENDING = 1,
  CONFIRMED = 2,
  PROCESSING = 3,
  SHIPPED = 4,
  IN_TRANSIT = 5,
  DELIVERED = 6,
  CANCELLED = 7,
  EXPIRED = 8,
  READY_FOR_PICKUP = 9,
  PICKED_UP = 10,
  FAILED_DELIVERY = 11,
  RETURNED = 12,
}

export enum PaymentMethod {
  UNSPECIFIED = 0,
  CREDIT_CARD = 1,
  KONBINI_SEVENELEVEN = 2,
  KONBINI_LAWSON = 3,
  KONBINI_FAMILYMART = 4,
  PAYPAY = 5,
  RAKUTEN_PAY = 6,
}

export interface ShippingAddress {
  name: string
  phone: string
  postal_code: string
  prefecture: string
  city: string
  address_line1: string
  address_line2: string
}

export interface OrderItem {
  id: string
  product_id: string
  variant_id: string
  product_name: string
  quantity: number
  unit_price: Money
  total_price: Money
}

export interface Order {
  id: string
  order_number: string
  user_id: string
  status: OrderStatus
  subtotal_amount: Money
  tax_amount: Money
  discount_amount: Money
  total_amount: Money
  points_applied: number
  shipping_address: ShippingAddress
  payment_method: PaymentMethod
  created_at: string
  updated_at: string
  delivery_slot_id?: string
  estimated_delivery_at?: string
  items: OrderItem[]
}

export interface CreateOrderRequest {
  user_id: string
  items: CreateOrderItem[]
  shipping_address: ShippingAddress
  payment_method: PaymentMethod
  points_to_apply?: string
  delivery_slot_id?: string
}

export interface CreateOrderItem {
  product_id: string
  variant_id: string
  quantity: number
}

export interface CreateOrderResponse {
  order_id: string
  order_number: string
  status: OrderStatus
}

export interface ListOrdersParams {
  status?: string
  page?: number
  limit?: number
}

export interface ListOrdersResponse {
  orders: Order[]
  pagination: Pagination
}

export interface CartItem {
  product_id: string
  variant_id: string
  product_name: string
  product_image: string
  unit_price: Money
  quantity: number
  stock_quantity: number
}

export interface CartSummary {
  item_count: number
  subtotal: Money
}
