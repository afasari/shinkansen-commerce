<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function handleLogin() {
  error.value = ''
  loading.value = true
  try {
    await authStore.login(email.value, password.value)
    const redirect = (route.query.redirect as string) || '/'
    router.push(redirect)
  } catch (e: unknown) {
    error.value = t('auth.invalidCredentials')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="card p-8">
    <h2 class="text-xl font-semibold text-gray-900 text-center mb-6">{{ t('auth.loginTitle') }}</h2>
    <form @submit.prevent="handleLogin" class="space-y-4">
      <div v-if="error" class="rounded-md bg-red-50 p-3">
        <p class="text-sm text-red-700">{{ error }}</p>
      </div>
      <div>
        <label class="label-field">{{ t('auth.email') }}</label>
        <input v-model="email" type="email" required class="input-field mt-1" />
      </div>
      <div>
        <label class="label-field">{{ t('auth.password') }}</label>
        <input v-model="password" type="password" required class="input-field mt-1" />
      </div>
      <button type="submit" :disabled="loading" class="btn-primary w-full">
        <span v-if="loading" class="animate-spin mr-2 inline-block h-4 w-4 border-2 border-white border-t-transparent rounded-full"></span>
        {{ t('auth.loginButton') }}
      </button>
    </form>
    <p class="mt-4 text-center text-sm text-gray-600">
      {{ t('auth.noAccount') }}
      <router-link to="/register" class="font-medium text-shinkansen-600 hover:text-shinkansen-500">{{ t('auth.signUpHere') }}</router-link>
    </p>
  </div>
</template>
