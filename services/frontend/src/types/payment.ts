import type { Money } from './common'

export enum PaymentStatus {
  UNSPECIFIED = 'PAYMENT_STATUS_UNSPECIFIED',
  PENDING = 'PAYMENT_STATUS_PENDING',
  PROCESSING = 'PAYMENT_STATUS_PROCESSING',
  COMPLETED = 'PAYMENT_STATUS_COMPLETED',
  FAILED = 'PAYMENT_STATUS_FAILED',
  CANCELLED = 'PAYMENT_STATUS_CANCELLED',
  REFUNDED = 'PAYMENT_STATUS_REFUNDED',
}

export interface Payment {
  id: string
  order_id: string
  method: string
  amount: Money
  status: PaymentStatus
  created_at: string
  updated_at: string
  transaction_id: string
}

export interface CreatePaymentRequest {
  order_id: string
  method: string
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
