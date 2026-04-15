import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { StockItem, StockMovement, UpdateStockRequest, ReserveStockRequest, ReserveStockResponse, Pagination } from '@/types'
import * as inventoryApi from '@/api/inventory'

export const useInventoryStore = defineStore('inventory', () => {
  const stockItems = ref<StockItem[]>([])
  const movements = ref<StockMovement[]>([])
  const movementPagination = ref<Pagination>({ page: 1, limit: 20, total: 0 })
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchStock(params: { product_id: string; variant_id?: string; warehouse_id: string }) {
    loading.value = true
    try {
      const stock = await inventoryApi.getStock(params)
      const idx = stockItems.value.findIndex(
        (s) => s.product_id === stock.product_id && s.variant_id === stock.variant_id,
      )
      if (idx >= 0) {
        stockItems.value[idx] = stock
      } else {
        stockItems.value.push(stock)
      }
      return stock
    } catch (e: unknown) {
      error.value = (e as Error).message
      return null
    } finally {
      loading.value = false
    }
  }

  async function updateStock(data: UpdateStockRequest) {
    await inventoryApi.updateStock(data)
  }

  async function reserveStock(data: ReserveStockRequest): Promise<ReserveStockResponse | null> {
    try {
      return await inventoryApi.reserveStock(data)
    } catch (e: unknown) {
      error.value = (e as Error).message
      return null
    }
  }

  async function releaseStock(reservationId: string) {
    await inventoryApi.releaseStock({ reservation_id: reservationId })
  }

  async function fetchMovements(stockItemId: string, page = 1, limit = 20) {
    loading.value = true
    try {
      const res = await inventoryApi.getStockMovements(stockItemId, page, limit)
      movements.value = res.movements
      movementPagination.value = res.pagination
    } catch (e: unknown) {
      error.value = (e as Error).message
    } finally {
      loading.value = false
    }
  }

  return {
    stockItems, movements, movementPagination, loading, error,
    fetchStock, updateStock, reserveStock, releaseStock, fetchMovements,
  }
})
