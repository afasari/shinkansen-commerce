import type { Pagination } from './common'

export enum ShipmentStatus {
  UNSPECIFIED = 0,
  PREPARING = 1,
  SHIPPED = 2,
  IN_TRANSIT = 3,
  DELIVERED = 4,
  CANCELLED = 5,
  FAILED_DELIVERY = 6,
}

export interface DeliverySlot {
  id: string
  delivery_zone_id: string
  start_time: string
  end_time: string
  capacity: number
  reserved: number
  available: number
  date: string
}

export interface DeliveryZone {
  id: string
  name: string
  postal_codes: string[]
  prefectures: string[]
  delivery_days: number
}

export interface TrackingEvent {
  id: string
  status: string
  location: string
  timestamp: string
  description: string
}

export interface Shipment {
  id: string
  order_id: string
  tracking_number: string
  status: ShipmentStatus
  estimated_delivery_at: string
  actual_delivery_at: string
  carrier: string
  tracking_events: TrackingEvent[]
}

export interface GetDeliverySlotsParams {
  delivery_zone_id: string
  date?: string
}

export interface ReserveDeliverySlotRequest {
  slot_id: string
  order_id: string
}

export interface ReserveDeliverySlotResponse {
  reservation_id: string
  reserved_at: string
}

export interface UpdateShipmentStatusRequest {
  status: ShipmentStatus
  description: string
}
