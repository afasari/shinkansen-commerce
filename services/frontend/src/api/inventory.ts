import { client } from './client'
import type {
  StockItem,
  StockMovementsResponse,
  GetStockParams,
  UpdateStockRequest,
  ReserveStockRequest,
  ReserveStockResponse,
  ReleaseStockRequest,
} from '@/types'

export async function getStock(params: GetStockParams): Promise<StockItem> {
  const res = await client.get<{ stock: StockItem }>('/v1/inventory/stock', { params })
  return res.data.stock
}

export async function updateStock(data: UpdateStockRequest): Promise<void> {
  await client.put('/v1/inventory/stock', data)
}

export async function reserveStock(data: ReserveStockRequest): Promise<ReserveStockResponse> {
  const res = await client.post<ReserveStockResponse>('/v1/inventory/reserve', data)
  return res.data
}

export async function releaseStock(data: ReleaseStockRequest): Promise<void> {
  await client.post('/v1/inventory/release', data)
}

export async function getStockMovements(stockItemId: string, page = 1, limit = 20): Promise<StockMovementsResponse> {
  const res = await client.get<StockMovementsResponse>(`/v1/inventory/stock/${stockItemId}/movements`, {
    params: { page, limit },
  })
  return res.data
}
