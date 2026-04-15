<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useProductStore } from '@/stores/product'
import { useOrderStore } from '@/stores/order'
import { formatPrice, formatDate } from '@/utils/format'
import { useRouter } from 'vue-router'
import {
  ShoppingBagIcon,
  CurrencyYenIcon,
  CubeIcon,
  ArchiveBoxIcon,
  TruckIcon,
  BanknotesIcon,
  ClipboardDocumentListIcon,
  ChartBarIcon,
  ExclamationTriangleIcon,
} from '@heroicons/vue/24/outline'

const { t } = useI18n()
const router = useRouter()
const productStore = useProductStore()
const orderStore = useOrderStore()

const apiError = ref(false)

onMounted(async () => {
  try {
    await Promise.all([
      productStore.fetchProducts({ limit: 5 }),
      orderStore.fetchOrders({ limit: 5 }),
    ])
  } catch {
    apiError.value = true
  }
})

const quickLinks = [
  { label: t('nav.products'), desc: t('product.createProduct') + ', ' + t('common.edit') + ', ' + t('common.delete'), icon: CubeIcon, route: 'admin-products', color: 'bg-purple-100 text-purple-600' },
  { label: t('nav.orders'), desc: t('order.updateStatus') + ', ' + t('order.cancelOrder'), icon: ClipboardDocumentListIcon, route: 'admin-orders', color: 'bg-blue-100 text-blue-600' },
  { label: t('nav.inventory'), desc: t('inventory.updateStock') + ', ' + t('inventory.movements'), icon: ArchiveBoxIcon, route: 'admin-inventory', color: 'bg-orange-100 text-orange-600' },
  { label: t('nav.delivery'), desc: t('nav.delivery') + ' slots', icon: TruckIcon, route: 'admin-delivery-slots', color: 'bg-teal-100 text-teal-600' },
  { label: t('nav.shipments'), desc: t('shipment.trackingNumber') + ', ' + t('shipment.updateStatus'), icon: ShoppingBagIcon, route: 'admin-shipments', color: 'bg-indigo-100 text-indigo-600' },
  { label: t('nav.payments'), desc: t('payment.refund') + ', ' + t('payment.title'), icon: BanknotesIcon, route: 'admin-payments', color: 'bg-green-100 text-green-600' },
]
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-gray-900">{{ t('admin.dashboard') }}</h1>
      <router-link to="/" class="btn-secondary text-sm gap-1">
        <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" /></svg>
        {{ t('nav.home') }}
      </router-link>
    </div>

    <div v-if="apiError" class="rounded-lg bg-yellow-50 border border-yellow-200 p-4 mb-6 flex items-start gap-3">
      <ExclamationTriangleIcon class="h-5 w-5 text-yellow-500 flex-shrink-0 mt-0.5" />
      <div>
        <p class="text-sm font-medium text-yellow-800">API Connection Failed</p>
        <p class="text-sm text-yellow-700 mt-0.5">Could not load data from the backend. Make sure the gateway is running on port 8080. The stats below may be empty.</p>
      </div>
    </div>

    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
      <div class="card p-5">
        <div class="flex items-center gap-3">
          <div class="h-10 w-10 rounded-lg bg-blue-100 flex items-center justify-center">
            <ShoppingBagIcon class="h-5 w-5 text-blue-600" />
          </div>
          <div>
            <p class="text-sm text-gray-500">{{ t('admin.totalOrders') }}</p>
            <p class="text-2xl font-bold text-gray-900">{{ orderStore.pagination.total }}</p>
          </div>
        </div>
      </div>
      <div class="card p-5">
        <div class="flex items-center gap-3">
          <div class="h-10 w-10 rounded-lg bg-green-100 flex items-center justify-center">
            <CurrencyYenIcon class="h-5 w-5 text-green-600" />
          </div>
          <div>
            <p class="text-sm text-gray-500">{{ t('admin.totalRevenue') }}</p>
            <p class="text-2xl font-bold text-gray-900">
              ¥{{ orderStore.orders.reduce((s, o) => s + (o.total_amount?.units || 0), 0).toLocaleString() }}
            </p>
          </div>
        </div>
      </div>
      <div class="card p-5">
        <div class="flex items-center gap-3">
          <div class="h-10 w-10 rounded-lg bg-purple-100 flex items-center justify-center">
            <CubeIcon class="h-5 w-5 text-purple-600" />
          </div>
          <div>
            <p class="text-sm text-gray-500">{{ t('admin.totalProducts') }}</p>
            <p class="text-2xl font-bold text-gray-900">{{ productStore.pagination.total }}</p>
          </div>
        </div>
      </div>
      <div class="card p-5">
        <div class="flex items-center gap-3">
          <div class="h-10 w-10 rounded-lg bg-shinkansen-100 flex items-center justify-center">
            <ChartBarIcon class="h-5 w-5 text-shinkansen-600" />
          </div>
          <div>
            <p class="text-sm text-gray-500">{{ t('admin.salesOverview') }}</p>
            <p class="text-lg font-bold text-gray-900">{{ orderStore.orders.length }} recent</p>
          </div>
        </div>
      </div>
    </div>

    <h2 class="text-lg font-semibold text-gray-900 mb-4">Quick Navigation</h2>
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 mb-8">
      <button v-for="link in quickLinks" :key="link.route" @click="router.push({ name: link.route })"
        class="card p-4 text-left hover:shadow-md transition-shadow group">
        <div class="flex items-center gap-3">
          <div :class="[link.color, 'h-10 w-10 rounded-lg flex items-center justify-center']">
            <component :is="link.icon" class="h-5 w-5" />
          </div>
          <div class="flex-1">
            <p class="text-sm font-semibold text-gray-900 group-hover:text-shinkansen-600">{{ link.label }}</p>
            <p class="text-xs text-gray-500 mt-0.5">{{ link.desc }}</p>
          </div>
          <svg class="h-4 w-4 text-gray-400 group-hover:text-shinkansen-600 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
        </div>
      </button>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <div class="card">
        <div class="p-4 border-b border-gray-200 flex items-center justify-between">
          <h2 class="font-semibold text-gray-900">{{ t('admin.recentOrders') }}</h2>
          <button @click="router.push({ name: 'admin-orders' })" class="text-sm text-shinkansen-600 hover:text-shinkansen-500">{{ t('product.viewAll') }}</button>
        </div>
        <div class="divide-y divide-gray-100">
          <div v-for="order in orderStore.orders" :key="order.id" class="px-4 py-3 flex items-center justify-between hover:bg-gray-50 cursor-pointer" @click="router.push({ name: 'admin-order-detail', params: { id: order.id } })">
            <div>
              <p class="text-sm font-medium text-gray-900">{{ order.order_number }}</p>
              <p class="text-xs text-gray-500">{{ formatDate(order.created_at) }}</p>
            </div>
            <div class="text-right">
              <p class="text-sm font-semibold">{{ formatPrice(order.total_amount) }}</p>
              <span :class="[
                order.status === 'ORDER_STATUS_DELIVERED' ? 'bg-green-100 text-green-700' :
                order.status === 'ORDER_STATUS_CANCELLED' ? 'bg-red-100 text-red-700' :
                'bg-yellow-100 text-yellow-700',
                'text-xs px-2 py-0.5 rounded-full font-medium'
              ]">{{ order.status.replace('ORDER_STATUS_', '').replace(/_/g, ' ') }}</span>
            </div>
          </div>
          <div v-if="orderStore.orders.length === 0" class="px-4 py-8 text-center">
            <ShoppingBagIcon class="h-8 w-8 text-gray-300 mx-auto mb-2" />
            <p class="text-sm text-gray-500">No orders yet</p>
            <p class="text-xs text-gray-400 mt-1">Orders will appear here once customers start placing them</p>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="p-4 border-b border-gray-200 flex items-center justify-between">
          <h2 class="font-semibold text-gray-900">{{ t('admin.topProducts') }}</h2>
          <button @click="router.push({ name: 'admin-products' })" class="text-sm text-shinkansen-600 hover:text-shinkansen-500">{{ t('product.viewAll') }}</button>
        </div>
        <div class="divide-y divide-gray-100">
          <div v-for="product in productStore.products" :key="product.id" class="px-4 py-3 flex items-center justify-between hover:bg-gray-50">
            <div class="flex items-center gap-3">
              <div class="h-10 w-10 rounded bg-gray-100 flex items-center justify-center flex-shrink-0">
                <CubeIcon class="h-5 w-5 text-gray-400" />
              </div>
              <div>
                <p class="text-sm font-medium text-gray-900 line-clamp-1">{{ product.name }}</p>
                <p class="text-xs" :class="product.stock_quantity > 0 ? 'text-green-600' : 'text-red-600'">{{ t('common.inStock') }}: {{ product.stock_quantity }}</p>
              </div>
            </div>
            <p class="text-sm font-semibold">{{ formatPrice(product.price) }}</p>
          </div>
          <div v-if="productStore.products.length === 0" class="px-4 py-8 text-center">
            <CubeIcon class="h-8 w-8 text-gray-300 mx-auto mb-2" />
            <p class="text-sm text-gray-500">No products yet</p>
            <router-link :to="{ name: 'admin-product-new' }" class="text-xs text-shinkansen-600 hover:text-shinkansen-500 mt-1 inline-block">{{ t('product.createProduct') }} &rarr;</router-link>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
