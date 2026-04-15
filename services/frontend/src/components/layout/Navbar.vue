<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { RouterLink, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useCartStore } from '@/stores/cart'
import { ShoppingBagIcon, UserCircleIcon, Bars3Icon, XMarkIcon, MagnifyingGlassIcon, Cog6ToothIcon } from '@heroicons/vue/24/outline'

const { t, locale } = useI18n()
const router = useRouter()
const authStore = useAuthStore()
const cartStore = useCartStore()
const mobileMenuOpen = ref(false)
const searchQuery = ref('')

function toggleLocale() {
  locale.value = locale.value === 'en' ? 'ja' : 'en'
  localStorage.setItem('locale', locale.value)
}

function handleSearch() {
  if (searchQuery.value.trim()) {
    router.push({ name: 'search', query: { q: searchQuery.value.trim() } })
    searchQuery.value = ''
    mobileMenuOpen.value = false
  }
}

function handleLogout() {
  authStore.logout()
  router.push('/')
  mobileMenuOpen.value = false
}
</script>

<template>
  <nav class="bg-white shadow-sm sticky top-0 z-30">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex h-16 items-center justify-between">
        <div class="flex items-center gap-8">
          <router-link to="/" class="text-xl font-bold text-shinkansen-600">{{ t('app.name') }}</router-link>
          <div class="hidden md:flex items-center gap-6">
            <router-link to="/products" class="text-sm font-medium text-gray-700 hover:text-shinkansen-600">{{ t('nav.products') }}</router-link>
            <router-link to="/search" class="text-sm font-medium text-gray-700 hover:text-shinkansen-600">{{ t('nav.search') }}</router-link>
          </div>
        </div>

        <div class="hidden md:flex items-center gap-2 flex-1 max-w-lg mx-8">
          <form @submit.prevent="handleSearch" class="w-full">
            <div class="relative">
              <MagnifyingGlassIcon class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
              <input v-model="searchQuery" type="text" :placeholder="t('product.searchPlaceholder')"
                class="input-field pl-10 pr-4 py-1.5 text-sm" />
            </div>
          </form>
        </div>

        <div class="hidden md:flex items-center gap-4">
          <button @click="toggleLocale" class="text-sm text-gray-500 hover:text-gray-700">
            {{ locale === 'en' ? '日本語' : 'English' }}
          </button>

          <router-link to="/cart" class="relative p-2 text-gray-700 hover:text-shinkansen-600">
            <ShoppingBagIcon class="h-6 w-6" />
            <span v-if="cartStore.itemCount > 0"
              class="absolute -top-1 -right-1 h-5 w-5 rounded-full bg-shinkansen-600 text-white text-xs flex items-center justify-center">
              {{ cartStore.itemCount }}
            </span>
          </router-link>

          <template v-if="authStore.isAuthenticated">
            <router-link to="/account/profile" class="p-2 text-gray-700 hover:text-shinkansen-600">
              <UserCircleIcon class="h-6 w-6" />
            </router-link>
            <router-link to="/admin" class="p-2 text-gray-700 hover:text-shinkansen-600" :title="t('nav.admin')">
              <Cog6ToothIcon class="h-6 w-6" />
            </router-link>
            <button @click="handleLogout" class="text-sm text-gray-500 hover:text-gray-700">{{ t('nav.logout') }}</button>
          </template>
          <template v-else>
            <router-link to="/login" class="text-sm font-medium text-gray-700 hover:text-shinkansen-600">{{ t('nav.login') }}</router-link>
            <router-link to="/register" class="btn-primary text-sm">{{ t('nav.register') }}</router-link>
          </template>
        </div>

        <button @click="mobileMenuOpen = !mobileMenuOpen" class="md:hidden p-2">
          <Bars3Icon v-if="!mobileMenuOpen" class="h-6 w-6" />
          <XMarkIcon v-else class="h-6 w-6" />
        </button>
      </div>
    </div>

    <div v-if="mobileMenuOpen" class="md:hidden border-t border-gray-200 bg-white">
      <div class="px-4 py-3 space-y-3">
        <form @submit.prevent="handleSearch">
          <input v-model="searchQuery" type="text" :placeholder="t('product.searchPlaceholder')" class="input-field text-sm" />
        </form>
        <router-link to="/products" @click="mobileMenuOpen = false" class="block text-sm font-medium text-gray-700 py-2">{{ t('nav.products') }}</router-link>
        <router-link to="/cart" @click="mobileMenuOpen = false" class="block text-sm font-medium text-gray-700 py-2">
          {{ t('nav.cart') }} ({{ cartStore.itemCount }})
        </router-link>
        <template v-if="authStore.isAuthenticated">
          <router-link to="/account/profile" @click="mobileMenuOpen = false" class="block text-sm font-medium text-gray-700 py-2">{{ t('nav.profile') }}</router-link>
          <router-link to="/account/orders" @click="mobileMenuOpen = false" class="block text-sm font-medium text-gray-700 py-2">{{ t('nav.orders') }}</router-link>
          <router-link to="/admin" @click="mobileMenuOpen = false" class="block text-sm font-medium text-shinkansen-600 py-2 font-semibold">{{ t('nav.admin') }}</router-link>
          <button @click="handleLogout" class="block text-sm font-medium text-gray-700 py-2">{{ t('nav.logout') }}</button>
        </template>
        <template v-else>
          <router-link to="/login" @click="mobileMenuOpen = false" class="block text-sm font-medium text-gray-700 py-2">{{ t('nav.login') }}</router-link>
          <router-link to="/register" @click="mobileMenuOpen = false" class="block text-sm font-medium text-gray-700 py-2">{{ t('nav.register') }}</router-link>
        </template>
        <button @click="toggleLocale" class="text-sm text-gray-500">{{ locale === 'en' ? '日本語' : 'English' }}</button>
      </div>
    </div>
  </nav>
</template>
