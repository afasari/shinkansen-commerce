import { client } from './client'
import type {
  Product,
  ProductVariant,
  ListProductsParams,
  ListProductsResponse,
  SearchProductsParams,
  CreateProductRequest,
  UpdateProductRequest,
} from '@/types'

export async function listProducts(params?: ListProductsParams): Promise<ListProductsResponse> {
  const res = await client.get<ListProductsResponse>('/v1/products', { params })
  return res.data
}

export async function getProduct(productId: string): Promise<Product> {
  const res = await client.get<{ product: Product }>(`/v1/products/${productId}`)
  return res.data.product
}

export async function createProduct(data: CreateProductRequest): Promise<{ product_id: string }> {
  const res = await client.post<{ product_id: string }>('/v1/products', data)
  return res.data
}

export async function updateProduct(productId: string, data: UpdateProductRequest): Promise<Product> {
  const res = await client.put<{ product: Product }>(`/v1/products/${productId}`, data)
  return res.data.product
}

export async function deleteProduct(productId: string): Promise<void> {
  await client.delete(`/v1/products/${productId}`)
}

export async function searchProducts(params: SearchProductsParams): Promise<ListProductsResponse> {
  const res = await client.get<ListProductsResponse>('/v1/products/search', { params })
  return res.data
}

export async function getProductVariants(productId: string): Promise<ProductVariant[]> {
  const res = await client.get<{ variants: ProductVariant[] }>('/v1/products/variants', {
    params: { product_id: productId },
  })
  return res.data.variants
}
