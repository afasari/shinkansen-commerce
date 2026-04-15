<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useAddressStore } from '@/stores/address'
import { PREFECTURES } from '@/utils/constants'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const addressStore = useAddressStore()

const isEdit = computed(() => !!route.params.id)
const saving = ref(false)

const name = ref('')
const phone = ref('')
const postalCode = ref('')
const prefecture = ref('')
const city = ref('')
const addressLine1 = ref('')
const addressLine2 = ref('')
const isDefault = ref(false)

onMounted(async () => {
  if (isEdit.value) {
    await addressStore.fetchAddresses()
    const addr = addressStore.addresses.find((a) => a.id === route.params.id)
    if (addr) {
      name.value = addr.name
      phone.value = addr.phone
      postalCode.value = addr.postal_code
      prefecture.value = addr.prefecture
      city.value = addr.city
      addressLine1.value = addr.address_line1
      addressLine2.value = addr.address_line2
      isDefault.value = addr.is_default
    }
  }
})

async function handleSave() {
  saving.value = true
  try {
    const data = {
      name: name.value,
      phone: phone.value,
      postal_code: postalCode.value,
      prefecture: prefecture.value,
      city: city.value,
      address_line1: addressLine1.value,
      address_line2: addressLine2.value,
      is_default: isDefault.value,
    }
    if (isEdit.value) {
      await addressStore.updateAddress(route.params.id as string, data)
    } else {
      await addressStore.addAddress(data)
    }
    router.push('/account/addresses')
  } catch (e: unknown) {
    alert((e as Error).message)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <h1 class="text-2xl font-bold text-gray-900 mb-6">{{ isEdit ? t('address.editAddress') : t('address.addAddress') }}</h1>

    <div class="card p-6">
      <form @submit.prevent="handleSave" class="space-y-4">
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label class="label-field">{{ t('address.name') }} *</label>
            <input v-model="name" required class="input-field mt-1" />
          </div>
          <div>
            <label class="label-field">{{ t('address.phone') }} *</label>
            <input v-model="phone" type="tel" required class="input-field mt-1" />
          </div>
        </div>
        <div>
          <label class="label-field">{{ t('address.postalCode') }} *</label>
          <input v-model="postalCode" required class="input-field mt-1" placeholder="123-4567" />
        </div>
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label class="label-field">{{ t('address.prefecture') }} *</label>
            <select v-model="prefecture" required class="input-field mt-1">
              <option value="" disabled>Select...</option>
              <option v-for="p in PREFECTURES" :key="p" :value="p">{{ p }}</option>
            </select>
          </div>
          <div>
            <label class="label-field">{{ t('address.city') }} *</label>
            <input v-model="city" required class="input-field mt-1" />
          </div>
        </div>
        <div>
          <label class="label-field">{{ t('address.addressLine1') }} *</label>
          <input v-model="addressLine1" required class="input-field mt-1" />
        </div>
        <div>
          <label class="label-field">{{ t('address.addressLine2') }}</label>
          <input v-model="addressLine2" class="input-field mt-1" />
        </div>
        <label class="flex items-center gap-2">
          <input type="checkbox" v-model="isDefault" class="rounded border-gray-300 text-shinkansen-600" />
          <span class="text-sm">{{ t('address.setDefault') }}</span>
        </label>
        <div class="flex items-center gap-3">
          <button type="submit" :disabled="saving" class="btn-primary">
            <span v-if="saving" class="animate-spin mr-2 inline-block h-4 w-4 border-2 border-white border-t-transparent rounded-full"></span>
            {{ t('common.save') }}
          </button>
          <button type="button" @click="router.back()" class="btn-secondary">{{ t('common.cancel') }}</button>
        </div>
      </form>
    </div>
  </div>
</template>
