import type { Pagination } from './common'

export enum ShipmentStatus {
  UNSPECIFIED = 'SHIPMENT_STATUS_UNSPECIFIED',
  PREPARING = 'SHIPMENT_STATUS_PREPARING',
  SHIPPED = 'SHIPMENT_STATUS_SHIPPED',
  IN_TRANSIT = 'SHIPMENT_STATUS_IN_TRANSIT',
  DELIVERED = 'SHIPMENT_STATUS_DELIVERED',
  CANCELLED = 'SHIPMENT_STATUS_CANCELLED',
  FAILED_DELIVERY = 'SHIPMENT_STATUS_FAILED_DELIVERY',
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
