<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const authStore = useAuthStore()

const name = ref('')
const phone = ref('')
const saving = ref(false)
const saved = ref(false)

onMounted(async () => {
  if (!authStore.user) await authStore.fetchUser()
  name.value = authStore.user?.name || ''
  phone.value = authStore.user?.phone || ''
})

async function handleSave() {
  saving.value = true
  saved.value = false
  try {
    await authStore.updateProfile({ name: name.value, phone: phone.value })
    saved.value = true
    setTimeout(() => { saved.value = false }, 2000)
  } catch (e: unknown) {
    alert((e as Error).message)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <h1 class="text-2xl font-bold text-gray-900 mb-6">{{ t('nav.profile') }}</h1>

    <div class="card p-6">
      <form @submit.prevent="handleSave" class="space-y-4">
        <div>
          <label class="label-field">{{ t('auth.email') }}</label>
          <input :value="authStore.user?.email" disabled class="input-field mt-1 bg-gray-50" />
        </div>
        <div>
          <label class="label-field">{{ t('auth.name') }}</label>
          <input v-model="name" type="text" required class="input-field mt-1" />
        </div>
        <div>
          <label class="label-field">{{ t('auth.phone') }}</label>
          <input v-model="phone" type="tel" class="input-field mt-1" />
        </div>
        <div class="flex items-center gap-3">
          <button type="submit" :disabled="saving" class="btn-primary">
            <span v-if="saving" class="animate-spin mr-2 inline-block h-4 w-4 border-2 border-white border-t-transparent rounded-full"></span>
            {{ t('common.save') }}
          </button>
          <span v-if="saved" class="text-sm text-green-600">{{ t('common.success') }}!</span>
        </div>
      </form>
    </div>

    <div class="mt-6 flex gap-4">
      <router-link to="/account/addresses" class="btn-secondary">{{ t('nav.addresses') }}</router-link>
      <router-link to="/account/orders" class="btn-secondary">{{ t('nav.orders') }}</router-link>
    </div>
  </div>
</template>
