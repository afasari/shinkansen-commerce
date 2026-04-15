import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/layouts/DefaultLayout.vue'),
    children: [
      { path: '', name: 'home', component: () => import('@/pages/customer/HomePage.vue') },
      { path: 'products', name: 'products', component: () => import('@/pages/customer/ProductListPage.vue') },
      { path: 'products/:id', name: 'product-detail', component: () => import('@/pages/customer/ProductDetailPage.vue') },
      { path: 'search', name: 'search', component: () => import('@/pages/customer/SearchPage.vue') },
      { path: 'cart', name: 'cart', component: () => import('@/pages/customer/CartPage.vue') },
      { path: 'checkout', name: 'checkout', component: () => import('@/pages/customer/CheckoutPage.vue'), meta: { requiresAuth: true } },
      { path: 'checkout/success', name: 'checkout-success', component: () => import('@/pages/customer/CheckoutSuccessPage.vue'), meta: { requiresAuth: true } },
      {
        path: 'account',
        meta: { requiresAuth: true },
        children: [
          { path: '', name: 'account', redirect: '/account/profile' },
          { path: 'profile', name: 'profile', component: () => import('@/pages/account/ProfilePage.vue') },
          { path: 'addresses', name: 'addresses', component: () => import('@/pages/account/AddressListPage.vue') },
          { path: 'addresses/new', name: 'address-new', component: () => import('@/pages/account/AddressFormPage.vue') },
          { path: 'addresses/:id/edit', name: 'address-edit', component: () => import('@/pages/account/AddressFormPage.vue') },
          { path: 'orders', name: 'orders', component: () => import('@/pages/customer/OrderListPage.vue') },
          { path: 'orders/:id', name: 'order-detail', component: () => import('@/pages/customer/OrderDetailPage.vue') },
        ],
      },
      { path: 'orders/:id/payment', name: 'payment', component: () => import('@/pages/customer/PaymentPage.vue'), meta: { requiresAuth: true } },
      { path: 'orders/:id/tracking', name: 'shipment-tracking', component: () => import('@/pages/customer/ShipmentTrackingPage.vue') },
    ],
  },
  {
    path: '/auth',
    component: () => import('@/layouts/AuthLayout.vue'),
    children: [
      { path: '/login', name: 'login', component: () => import('@/pages/auth/LoginPage.vue'), meta: { guestOnly: true } },
      { path: '/register', name: 'register', component: () => import('@/pages/auth/RegisterPage.vue'), meta: { guestOnly: true } },
    ],
  },
  {
    path: '/admin',
    component: () => import('@/layouts/AdminLayout.vue'),
    meta: { requiresAuth: true, requiresAdmin: true },
    children: [
      { path: '', name: 'admin-dashboard', component: () => import('@/pages/admin/DashboardPage.vue') },
      { path: 'products', name: 'admin-products', component: () => import('@/pages/admin/ProductListPage.vue') },
      { path: 'products/new', name: 'admin-product-new', component: () => import('@/pages/admin/ProductFormPage.vue') },
      { path: 'products/:id/edit', name: 'admin-product-edit', component: () => import('@/pages/admin/ProductFormPage.vue') },
      { path: 'orders', name: 'admin-orders', component: () => import('@/pages/admin/AdminOrderListPage.vue') },
      { path: 'orders/:id', name: 'admin-order-detail', component: () => import('@/pages/admin/AdminOrderDetailPage.vue') },
      { path: 'inventory', name: 'admin-inventory', component: () => import('@/pages/admin/InventoryPage.vue') },
      { path: 'inventory/movements/:id?', name: 'admin-stock-movements', component: () => import('@/pages/admin/StockMovementsPage.vue') },
      { path: 'delivery/slots', name: 'admin-delivery-slots', component: () => import('@/pages/admin/DeliverySlotsPage.vue') },
      { path: 'delivery/shipments', name: 'admin-shipments', component: () => import('@/pages/admin/ShipmentsPage.vue') },
      { path: 'payments', name: 'admin-payments', component: () => import('@/pages/admin/PaymentsPage.vue') },
    ],
  },
  { path: '/:pathMatch(.*)*', name: 'not-found', redirect: '/' },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior: () => ({ top: 0 }),
})

router.beforeEach(async (to) => {
  const authStore = useAuthStore()

  if (authStore.isAuthenticated && !authStore.user) {
    await authStore.fetchUser()
  }

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  if (to.meta.requiresAdmin && !authStore.isAdmin) {
    return { name: 'home' }
  }

  if (to.meta.guestOnly && authStore.isAuthenticated) {
    return { name: 'home' }
  }
})

export default router
