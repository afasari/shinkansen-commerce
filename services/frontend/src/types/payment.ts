import type { Money } from './common'
import type { PaymentMethod } from './order'

export enum PaymentStatus {
  UNSPECIFIED = 0,
  PENDING = 1,
  PROCESSING = 2,
  COMPLETED = 3,
  FAILED = 4,
  CANCELLED = 5,
  REFUNDED = 6,
}

export interface Payment {
  id: string
  order_id: string
  method: string | number
  amount: Money
  status: PaymentStatus | string
  created_at: string
  updated_at: string
  transaction_id: string
}

export interface CreatePaymentRequest {
  order_id: string
  method: PaymentMethod | string | number
  amount: Money
}

export interface CreatePaymentResponse {
  payment_id: string
  status: PaymentStatus
}

export interface ProcessPaymentRequest {
  payment_data: Record<string, string>
}

export interface ProcessPaymentResponse {
  status: string
  transaction_id: string
}

export interface RefundPaymentRequest {
  amount: Money
}

export interface PointBalance {
  available_points: number
  pending_points: number
  last_updated: string
}

export interface PointTransaction {
  id: string
  user_id: string
  amount: number
  type: string
  reason: string
  created_at: string
}
