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
  orderStore.fetchOrders({
    status: statusFilter.value || undefined,
    page: page.value,
    limit: 20,
  })
}

function goToPage(p: number) {
  page.value = p
  loadOrders()
}
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-gray-900">{{ t('order.orderHistory') }}</h1>
      <select v-model="statusFilter" @change="page = 1; loadOrders()" class="input-field w-auto text-sm">
        <option value="">{{ t('common.all') }} {{ t('order.status') }}</option>
        <option v-for="s in ORDER_STATUS_LIST" :key="s" :value="s">{{ ORDER_STATUS_LABELS[s] || s }}</option>
      </select>
    </div>

    <div v-if="orderStore.loading" class="text-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-shinkansen-200 border-t-shinkansen-600 mx-auto"></div>
    </div>

    <div v-else-if="orderStore.orders.length === 0" class="text-center py-12 text-gray-500">
      {{ t('order.noOrders') }}
    </div>

    <template v-else>
      <div class="space-y-4">
        <router-link v-for="order in orderStore.orders" :key="order.id"
          :to="{ name: 'order-detail', params: { id: order.id } }"
          class="card p-4 block hover:shadow-md transition-shadow">
          <div class="flex items-center justify-between">
            <div>
              <p class="font-medium text-gray-900">{{ order.order_number }}</p>
              <p class="text-sm text-gray-500 mt-0.5">{{ formatDate(order.created_at) }}</p>
              <p class="text-sm text-gray-600 mt-0.5">{{ order.items.length }} {{ t('common.items') }}</p>
            </div>
            <div class="text-right">
              <p class="font-semibold text-gray-900">{{ formatPrice(order.total_amount) }}</p>
              <div class="mt-1">
                <StatusBadge :status="order.status" size="sm" />
              </div>
            </div>
          </div>
        </router-link>
      </div>
      <AppPagination :current-page="page" :total-pages="totalPages" @page-change="goToPage" />
    </template>
  </div>
</template>
