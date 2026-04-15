import { client } from './client'
import type {
  Payment,
  CreatePaymentRequest,
  CreatePaymentResponse,
  ProcessPaymentRequest,
  ProcessPaymentResponse,
  RefundPaymentRequest,
} from '@/types'

export async function createPayment(data: CreatePaymentRequest): Promise<CreatePaymentResponse> {
  const res = await client.post<CreatePaymentResponse>('/v1/payments', data)
  return res.data
}

export async function getPayment(paymentId: string): Promise<Payment> {
  const res = await client.get<{ payment: Payment }>(`/v1/payments/${paymentId}`)
  return res.data.payment
}

export async function processPayment(paymentId: string, data: ProcessPaymentRequest): Promise<ProcessPaymentResponse> {
  const res = await client.post<ProcessPaymentResponse>(`/v1/payments/${paymentId}/process`, data)
  return res.data
}

export async function refundPayment(paymentId: string, data: RefundPaymentRequest): Promise<void> {
  await client.post(`/v1/payments/${paymentId}/refund`, data)
}
