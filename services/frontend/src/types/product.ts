import type { Money, Pagination } from './common'

export interface Product {
  id: string
  name: string
  description: string
  category_id: string
  price: Money
  sku: string
  active: boolean
  created_at: string
  updated_at: string
  image_urls: string[]
  stock_quantity: number
}

export interface ProductVariant {
  id: string
  product_id: string
  name: string
  attributes: Record<string, string>
  price: Money
  sku: string
  stock_quantity: number
}

export interface Category {
  id: string
  name: string
  parent_id: string
  level: number
  child_ids: string[]
}

export interface ListProductsParams {
  category_id?: string
  active_only?: boolean
  page?: number
  limit?: number
}

export interface SearchProductsParams {
  q: string
  category_id?: string
  min_price?: number
  max_price?: number
  in_stock_only?: boolean
  page?: number
  limit?: number
}

export interface ListProductsResponse {
  products: Product[]
  pagination: Pagination
}

export interface CreateProductRequest {
  name: string
  description: string
  category_id: string
  price: Money
  sku: string
  image_urls: string[]
  stock_quantity: number
}

export interface UpdateProductRequest {
  name?: string
  description?: string
  category_id?: string
  price?: Money
  active?: boolean
  image_urls?: string[]
}
