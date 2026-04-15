<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useProductStore } from '@/stores/product'
import { formatPrice, formatDate } from '@/utils/format'
import AppPagination from '@/components/common/AppPagination.vue'
import { PlusIcon, PencilSquareIcon, TrashIcon } from '@heroicons/vue/24/outline'
import AppModal from '@/components/common/AppModal.vue'

const { t } = useI18n()
const router = useRouter()
const productStore = useProductStore()

const page = ref(1)
const totalPages = ref(1)
const deleteTarget = ref<string | null>(null)
const showDeleteModal = ref(false)

onMounted(() => { loadProducts() })

watch(() => productStore.pagination, (pag) => {
  totalPages.value = Math.ceil(pag.total / pag.limit) || 1
}, { deep: true })

function loadProducts() {
  productStore.fetchProducts({ page: page.value, limit: 20 })
}

function goToPage(p: number) {
  page.value = p
  loadProducts()
}

async function handleDelete() {
  if (deleteTarget.value) {
    await productStore.deleteProduct(deleteTarget.value)
    deleteTarget.value = null
    loadProducts()
  }
  showDeleteModal.value = false
}

function confirmDelete(id: string) {
  deleteTarget.value = id
  showDeleteModal.value = true
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-gray-900">{{ t('product.title') }}</h1>
      <router-link :to="{ name: 'admin-product-new' }" class="btn-primary gap-1">
        <PlusIcon class="h-4 w-4" /> {{ t('product.createProduct') }}
      </router-link>
    </div>

    <div v-if="productStore.loading" class="text-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-shinkansen-200 border-t-shinkansen-600 mx-auto"></div>
    </div>

    <div v-else class="card overflow-hidden">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('product.productName') }}</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('common.price') }}</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('product.stockQuantity') }}</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('common.status') }}</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('common.date') }}</th>
            <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">{{ t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100">
          <tr v-for="product in productStore.products" :key="product.id" class="hover:bg-gray-50">
            <td class="px-4 py-3 text-sm font-medium text-gray-900">{{ product.name }}</td>
            <td class="px-4 py-3 text-sm text-gray-600">{{ formatPrice(product.price) }}</td>
            <td class="px-4 py-3 text-sm" :class="product.stock_quantity > 0 ? 'text-green-600' : 'text-red-600'">{{ product.stock_quantity }}</td>
            <td class="px-4 py-3 text-sm">
              <span :class="[product.active ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-600', 'px-2 py-0.5 rounded-full text-xs font-medium']">
                {{ product.active ? t('product.active') : t('product.inactive') }}
              </span>
            </td>
            <td class="px-4 py-3 text-sm text-gray-500">{{ formatDate(product.created_at) }}</td>
            <td class="px-4 py-3 text-right">
              <div class="flex items-center justify-end gap-2">
                <router-link :to="{ name: 'admin-product-edit', params: { id: product.id } }" class="text-shinkansen-600 hover:text-shinkansen-500">
                  <PencilSquareIcon class="h-4 w-4" />
                </router-link>
                <button @click="confirmDelete(product.id)" class="text-red-500 hover:text-red-600">
                  <TrashIcon class="h-4 w-4" />
                </button>
              </div>
            </td>
          </tr>
          <tr v-if="productStore.products.length === 0">
            <td colspan="6" class="px-4 py-8 text-center text-sm text-gray-500">{{ t('product.noProducts') }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <AppPagination :current-page="page" :total-pages="totalPages" @page-change="goToPage" />

    <AppModal :open="showDeleteModal" :message="t('product.deleteConfirm')" :danger="true"
      :confirm-label="t('common.delete')" @close="showDeleteModal = false" @confirm="handleDelete" />
  </div>
</template>
