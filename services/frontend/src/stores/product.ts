import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Product, ProductVariant, ListProductsParams, SearchProductsParams, Pagination } from '@/types'
import * as productsApi from '@/api/products'

export const useProductStore = defineStore('product', () => {
  const products = ref<Product[]>([])
  const currentProduct = ref<Product | null>(null)
  const variants = ref<ProductVariant[]>([])
  const pagination = ref<Pagination>({ page: 1, limit: 20, total: 0 })
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchProducts(params?: ListProductsParams) {
    loading.value = true
    error.value = null
    try {
      const res = await productsApi.listProducts(params)
      products.value = res.products
      pagination.value = res.pagination
    } catch (e: unknown) {
      error.value = (e as Error).message
    } finally {
      loading.value = false
    }
  }

  async function fetchProduct(productId: string) {
    loading.value = true
    error.value = null
    try {
      currentProduct.value = await productsApi.getProduct(productId)
    } catch (e: unknown) {
      error.value = (e as Error).message
    } finally {
      loading.value = false
    }
  }

  async function searchProducts(params: SearchProductsParams) {
    loading.value = true
    error.value = null
    try {
      const res = await productsApi.searchProducts(params)
      products.value = res.products
      pagination.value = res.pagination
    } catch (e: unknown) {
      error.value = (e as Error).message
    } finally {
      loading.value = false
    }
  }

  async function fetchVariants(productId: string) {
    try {
      variants.value = await productsApi.getProductVariants(productId)
    } catch {
      variants.value = []
    }
  }

  async function createProduct(data: Parameters<typeof productsApi.createProduct>[0]) {
    return await productsApi.createProduct(data)
  }

  async function updateProduct(productId: string, data: Parameters<typeof productsApi.updateProduct>[1]) {
    return await productsApi.updateProduct(productId, data)
  }

  async function deleteProduct(productId: string) {
    await productsApi.deleteProduct(productId)
  }

  return {
    products, currentProduct, variants, pagination, loading, error,
    fetchProducts, fetchProduct, searchProducts, fetchVariants,
    createProduct, updateProduct, deleteProduct,
  }
})
