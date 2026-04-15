<script setup lang="ts">
import { onMounted } from 'vue'
import { useCartStore } from '@/stores/cart'
import { useHealthCheck } from '@/composables/useHealthCheck'
import ApiDisconnectedBanner from '@/components/common/ApiDisconnectedBanner.vue'

const cartStore = useCartStore()
const { isHealthy } = useHealthCheck()

onMounted(() => {
  cartStore.initSession()
})
</script>

<template>
  <div class="min-h-screen flex flex-col">
    <ApiDisconnectedBanner :show="!isHealthy" />
    <Navbar />
    <main class="flex-1">
      <router-view />
    </main>
    <footer class="bg-white border-t border-gray-200 py-8 mt-auto">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
          <div>
            <h3 class="text-sm font-semibold text-gray-900">{{ $t('app.name') }}</h3>
            <p class="mt-2 text-sm text-gray-500">{{ $t('app.tagline') }}</p>
          </div>
          <div>
            <h3 class="text-sm font-semibold text-gray-900">{{ $t('nav.products') }}</h3>
            <ul class="mt-2 space-y-1">
              <li><router-link to="/products" class="text-sm text-gray-500 hover:text-shinkansen-600">{{ $t('nav.products') }}</router-link></li>
              <li><router-link to="/search" class="text-sm text-gray-500 hover:text-shinkansen-600">{{ $t('nav.search') }}</router-link></li>
            </ul>
          </div>
          <div>
            <h3 class="text-sm font-semibold text-gray-900">{{ $t('nav.account') }}</h3>
            <ul class="mt-2 space-y-1">
              <li><router-link to="/account/profile" class="text-sm text-gray-500 hover:text-shinkansen-600">{{ $t('nav.profile') }}</router-link></li>
              <li><router-link to="/account/orders" class="text-sm text-gray-500 hover:text-shinkansen-600">{{ $t('nav.orders') }}</router-link></li>
            </ul>
          </div>
        </div>
        <div class="mt-8 border-t border-gray-200 pt-4">
          <p class="text-sm text-gray-400 text-center">&copy; {{ new Date().getFullYear() }} Shinkansen Commerce</p>
        </div>
      </div>
    </footer>
  </div>
</template>
