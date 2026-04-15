<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { RouterLink, useRoute } from 'vue-router'
import { useHealthCheck } from '@/composables/useHealthCheck'
import ApiDisconnectedBanner from '@/components/common/ApiDisconnectedBanner.vue'
import {
  Bars3Icon,
  XMarkIcon,
  HomeIcon,
  CubeIcon,
  ShoppingCartIcon,
  TruckIcon,
  BanknotesIcon,
  ClipboardDocumentListIcon,
  ChartBarIcon,
  ArchiveBoxIcon,
} from '@heroicons/vue/24/outline'

const { locale } = useI18n()
const route = useRoute()
const sidebarOpen = ref(false)
const { isHealthy } = useHealthCheck()

function toggleLocale() {
  locale.value = locale.value === 'en' ? 'ja' : 'en'
  localStorage.setItem('locale', locale.value)
}

const navItems = [
  { name: 'admin.dashboard', icon: ChartBarIcon, route: 'admin-dashboard' },
  { name: 'nav.orders', icon: ClipboardDocumentListIcon, route: 'admin-orders' },
  { name: 'nav.products', icon: CubeIcon, route: 'admin-products' },
  { name: 'nav.inventory', icon: ArchiveBoxIcon, route: 'admin-inventory' },
  { name: 'nav.delivery', icon: TruckIcon, route: 'admin-delivery-slots' },
  { name: 'nav.shipments', icon: ShoppingCartIcon, route: 'admin-shipments' },
  { name: 'nav.payments', icon: BanknotesIcon, route: 'admin-payments' },
]

function isActive(routeName: string): boolean {
  return route.name === routeName
}
</script>

<template>
  <div class="min-h-screen bg-gray-100">
    <ApiDisconnectedBanner :show="!isHealthy" />
    <div class="lg:hidden fixed inset-0 z-40 flex" v-if="sidebarOpen">
      <div class="fixed inset-0 bg-gray-600 bg-opacity-75" @click="sidebarOpen = false" />
      <div class="relative flex-1 flex flex-col max-w-xs w-full bg-white">
        <div class="absolute top-0 right-0 -mr-12 pt-2">
          <button @click="sidebarOpen = false" class="ml-1 flex items-center justify-center h-10 w-10 rounded-full focus:outline-none focus:ring-2 focus:ring-inset focus:ring-white">
            <XMarkIcon class="h-6 w-6 text-white" />
          </button>
        </div>
        <div class="flex-1 h-0 pt-5 pb-4 overflow-y-auto">
          <div class="flex-shrink-0 flex items-center px-4">
            <span class="text-xl font-bold text-shinkansen-600">{{ $t('app.name') }}</span>
          </div>
          <nav class="mt-5 px-2 space-y-1">
            <router-link v-for="item in navItems" :key="item.route" :to="{ name: item.route }" @click="sidebarOpen = false"
              :class="[isActive(item.route) ? 'bg-shinkansen-50 text-shinkansen-700' : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900', 'group flex items-center px-2 py-2 text-sm font-medium rounded-md']">
              <component :is="item.icon" class="mr-3 h-5 w-5 flex-shrink-0" />
              {{ $t(item.name) }}
            </router-link>
          </nav>
        </div>
      </div>
    </div>

    <div class="hidden lg:flex lg:w-64 lg:flex-col lg:fixed lg:inset-y-0">
      <div class="flex-1 flex flex-col min-h-0 bg-white border-r border-gray-200">
        <div class="flex-1 flex flex-col pt-5 pb-4 overflow-y-auto">
          <div class="flex items-center flex-shrink-0 px-4">
            <router-link to="/" class="text-xl font-bold text-shinkansen-600">{{ $t('app.name') }}</router-link>
          </div>
          <nav class="mt-5 flex-1 px-2 space-y-1">
            <router-link v-for="item in navItems" :key="item.route" :to="{ name: item.route }"
              :class="[isActive(item.route) ? 'bg-shinkansen-50 text-shinkansen-700' : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900', 'group flex items-center px-2 py-2 text-sm font-medium rounded-md']">
              <component :is="item.icon" class="mr-3 h-5 w-5 flex-shrink-0" />
              {{ $t(item.name) }}
            </router-link>
          </nav>
        </div>
      </div>
    </div>

    <div class="lg:pl-64 flex flex-col flex-1">
      <div class="sticky top-0 z-10 bg-white border-b border-gray-200 px-4 py-3 sm:px-6 lg:px-8">
        <div class="flex items-center justify-between">
          <button @click="sidebarOpen = true" class="lg:hidden -ml-0.5 -mt-0.5 h-12 w-12 inline-flex items-center justify-center rounded-md text-gray-500 hover:text-gray-900">
            <Bars3Icon class="h-6 w-6" />
          </button>
          <div class="flex items-center gap-4">
            <button @click="toggleLocale" class="text-sm text-gray-500 hover:text-gray-700">
              {{ locale === 'en' ? '日本語' : 'English' }}
            </button>
            <router-link to="/" class="text-sm text-gray-500 hover:text-gray-700 flex items-center gap-1">
              <HomeIcon class="h-4 w-4" />
              {{ $t('nav.home') }}
            </router-link>
          </div>
        </div>
      </div>
      <main class="flex-1 p-4 sm:p-6 lg:p-8">
        <router-view />
      </main>
    </div>
  </div>
</template>
