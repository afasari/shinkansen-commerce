import type { Money, Pagination } from './common'

export enum OrderStatus {
  UNSPECIFIED = 'ORDER_STATUS_UNSPECIFIED',
  PENDING = 'ORDER_STATUS_PENDING',
  CONFIRMED = 'ORDER_STATUS_CONFIRMED',
  PROCESSING = 'ORDER_STATUS_PROCESSING',
  SHIPPED = 'ORDER_STATUS_SHIPPED',
  IN_TRANSIT = 'ORDER_STATUS_IN_TRANSIT',
  DELIVERED = 'ORDER_STATUS_DELIVERED',
  CANCELLED = 'ORDER_STATUS_CANCELLED',
  EXPIRED = 'ORDER_STATUS_EXPIRED',
  READY_FOR_PICKUP = 'ORDER_STATUS_READY_FOR_PICKUP',
  PICKED_UP = 'ORDER_STATUS_PICKED_UP',
  FAILED_DELIVERY = 'ORDER_STATUS_FAILED_DELIVERY',
  RETURNED = 'ORDER_STATUS_RETURNED',
}

export enum PaymentMethod {
  UNSPECIFIED = 'PAYMENT_METHOD_UNSPECIFIED',
  CREDIT_CARD = 'PAYMENT_METHOD_CREDIT_CARD',
  KONBINI_SEVENELEVEN = 'PAYMENT_METHOD_KONBINI_SEVENELEVEN',
  KONBINI_LAWSON = 'PAYMENT_METHOD_KONBINI_LAWSON',
  KONBINI_FAMILYMART = 'PAYMENT_METHOD_KONBINI_FAMILYMART',
  PAYPAY = 'PAYMENT_METHOD_PAYPAY',
  RAKUTEN_PAY = 'PAYMENT_METHOD_RAKUTEN_PAY',
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
  points_applied: string
  shipping_address: ShippingAddress
  payment_method: PaymentMethod
  created_at: string
  updated_at: string
  delivery_slot_id: string
  estimated_delivery_at: string
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
