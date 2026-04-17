import type { Pagination } from './common'

export interface StockItem {
  id: string
  product_id: string
  variant_id: string
  warehouse_id: string
  quantity: number
  reserved_quantity: number
  available_quantity: number
  updated_at: string
}

export enum MovementType {
  UNSPECIFIED = 0,
  INBOUND = 1,
  OUTBOUND = 2,
  RESERVATION = 3,
  RELEASE = 4,
  ADJUSTMENT = 5,
}

export interface StockMovement {
  id: string
  stock_item_id: string
  type: MovementType
  quantity: number
  reference: string
  created_at: string
}

export interface GetStockParams {
  product_id: string
  variant_id?: string
  warehouse_id: string
}

export interface UpdateStockRequest {
  stock_item_id: string
  product_id: string
  variant_id: string
  warehouse_id: string
  quantity_delta: number
  reason: string
  reference: string
}

export interface ReserveStockRequest {
  order_id: string
  items: StockReservationItem[]
  expires_at: string
}

export interface StockReservationItem {
  product_id: string
  variant_id: string
  warehouse_id: string
  quantity: number
}

export interface ReserveStockResponse {
  reservation_id: string
  success: boolean
  failed_items: string[]
}

export interface ReleaseStockRequest {
  reservation_id: string
}

export interface StockMovementsResponse {
  movements: StockMovement[]
  pagination: Pagination
}
