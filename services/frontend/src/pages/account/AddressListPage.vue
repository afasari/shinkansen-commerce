<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useAddressStore } from '@/stores/address'
import AppModal from '@/components/common/AppModal.vue'

const { t } = useI18n()
const router = useRouter()
const addressStore = useAddressStore()

const deleteTarget = ref<string | null>(null)
const showDeleteModal = ref(false)

onMounted(() => {
  addressStore.fetchAddresses()
})

function confirmDelete(id: string) {
  deleteTarget.value = id
  showDeleteModal.value = true
}

async function handleDelete() {
  if (deleteTarget.value) {
    await addressStore.deleteAddress(deleteTarget.value)
    deleteTarget.value = null
  }
  showDeleteModal.value = false
}
</script>

<template>
  <div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-gray-900">{{ t('address.title') }}</h1>
      <router-link to="/account/addresses/new" class="btn-primary">{{ t('address.addAddress') }}</router-link>
    </div>

    <div v-if="addressStore.loading" class="text-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-shinkansen-200 border-t-shinkansen-600 mx-auto"></div>
    </div>

    <div v-else-if="addressStore.addresses.length === 0" class="text-center py-12 text-gray-500">
      {{ t('address.noAddresses') }}
    </div>

    <div v-else class="space-y-4">
      <div v-for="addr in addressStore.addresses" :key="addr.id" class="card p-4">
        <div class="flex items-start justify-between">
          <div>
            <div class="flex items-center gap-2">
              <span class="font-medium text-gray-900">{{ addr.name }}</span>
              <span v-if="addr.is_default" class="px-2 py-0.5 text-xs bg-shinkansen-100 text-shinkansen-700 rounded-full">{{ t('address.default') }}</span>
            </div>
            <p class="text-sm text-gray-600 mt-1">{{ addr.phone }}</p>
            <p class="text-sm text-gray-600">{{ addr.postal_code }}</p>
            <p class="text-sm text-gray-600">{{ addr.prefecture }} {{ addr.city }} {{ addr.address_line1 }}</p>
            <p v-if="addr.address_line2" class="text-sm text-gray-600">{{ addr.address_line2 }}</p>
          </div>
          <div class="flex items-center gap-2">
            <router-link :to="{ name: 'address-edit', params: { id: addr.id } }" class="text-sm text-shinkansen-600 hover:text-shinkansen-500">{{ t('common.edit') }}</router-link>
            <button @click="confirmDelete(addr.id)" class="text-sm text-red-600 hover:text-red-500">{{ t('common.delete') }}</button>
          </div>
        </div>
      </div>
    </div>

    <AppModal :open="showDeleteModal" :message="t('address.deleteConfirm')" :danger="true"
      :confirm-label="t('common.delete')" @close="showDeleteModal = false" @confirm="handleDelete" />
  </div>
</template>
