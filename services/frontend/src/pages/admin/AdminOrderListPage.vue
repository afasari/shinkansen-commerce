<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useOrderStore } from '@/stores/order'
import { formatPrice, formatDate } from '@/utils/format'
import { OrderStatus, ORDER_STATUS_LIST, ORDER_STATUS_LABELS } from '@/utils/constants'
import StatusBadge from '@/components/common/StatusBadge.vue'
import AppPagination from '@/components/common/AppPagination.vue'

const { t } = useI18n()
const router = useRouter()
const orderStore = useOrderStore()

const page = ref(1)
const statusFilter = ref('')
const totalPages = ref(1)

onMounted(() => { loadOrders() })

watch(() => orderStore.pagination, (pag) => {
  totalPages.value = Math.ceil(pag.total / pag.limit) || 1
}, { deep: true })

function loadOrders() {
  orderStore.fetchOrders({ status: statusFilter.value || undefined, page: page.value, limit: 20 })
}

function goToPage(p: number) {
  page.value = p
  loadOrders()
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-gray-900">{{ t('order.title') }}</h1>
      <select v-model="statusFilter" @change="page = 1; loadOrders()" class="input-field w-auto text-sm">
        <option value="">{{ t('common.all') }}</option>
        <option v-for="s in ORDER_STATUS_LIST" :key="s" :value="s">{{ ORDER_STATUS_LABELS[s] || s }}</option>
      </select>
    </div>

    <div v-if="orderStore.loading" class="text-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-shinkansen-200 border-t-shinkansen-600 mx-auto"></div>
    </div>

    <div v-else class="card overflow-hidden">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('order.orderNumber') }}</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('common.total') }}</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('order.status') }}</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('common.date') }}</th>
            <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">{{ t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100">
          <tr v-for="order in orderStore.orders" :key="order.id" class="hover:bg-gray-50">
            <td class="px-4 py-3 text-sm font-medium text-gray-900">{{ order.order_number }}</td>
            <td class="px-4 py-3 text-sm text-gray-600">{{ formatPrice(order.total_amount) }}</td>
            <td class="px-4 py-3"><StatusBadge :status="order.status" size="sm" /></td>
            <td class="px-4 py-3 text-sm text-gray-500">{{ formatDate(order.created_at) }}</td>
            <td class="px-4 py-3 text-right">
              <router-link :to="{ name: 'admin-order-detail', params: { id: order.id } }" class="text-sm text-shinkansen-600 hover:text-shinkansen-500">
                {{ t('common.view') }}
              </router-link>
            </td>
          </tr>
          <tr v-if="orderStore.orders.length === 0">
            <td colspan="5" class="px-4 py-8 text-center text-sm text-gray-500">{{ t('order.noOrders') }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <AppPagination :current-page="page" :total-pages="totalPages" @page-change="goToPage" />
  </div>
</template>
