import { client } from './client'
import type {
  DeliverySlot,
  Shipment,
  GetDeliverySlotsParams,
  ReserveDeliverySlotRequest,
  ReserveDeliverySlotResponse,
  UpdateShipmentStatusRequest,
} from '@/types'

export async function getDeliverySlots(params: GetDeliverySlotsParams): Promise<DeliverySlot[]> {
  const res = await client.get<{ slots: DeliverySlot[] }>('/v1/delivery/slots', { params })
  return res.data.slots
}

export async function reserveDeliverySlot(data: ReserveDeliverySlotRequest): Promise<ReserveDeliverySlotResponse> {
  const res = await client.post<ReserveDeliverySlotResponse>('/v1/delivery/slots', data)
  return res.data
}

export async function reserveDeliverySlotByPath(slotId: string, orderId: string): Promise<ReserveDeliverySlotResponse> {
  const res = await client.post<ReserveDeliverySlotResponse>(`/v1/delivery/slots/${slotId}/reserve`, { order_id: orderId })
  return res.data
}

export async function getShipment(shipmentId: string): Promise<Shipment> {
  const res = await client.get<{ shipment: Shipment }>(`/v1/shipments/${shipmentId}`)
  return res.data.shipment
}

export async function updateShipmentStatus(shipmentId: string, data: UpdateShipmentStatusRequest): Promise<void> {
  await client.put(`/v1/shipments/${shipmentId}/status`, data)
}
