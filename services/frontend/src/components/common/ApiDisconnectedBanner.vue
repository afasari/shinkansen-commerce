<script setup lang="ts">
import { useI18n } from 'vue-i18n'

defineProps<{ show: boolean }>()

const { t } = useI18n()
</script>

<template>
  <Transition enter-active-class="transition ease-out duration-200" enter-from-class="opacity-0 -translate-y-2" enter-to-class="opacity-100 translate-y-0" leave-active-class="transition ease-in duration-150" leave-from-class="opacity-100 translate-y-0" leave-to-class="opacity-0 -translate-y-2">
    <div v-if="show" class="bg-red-50 border-b border-red-200 px-4 py-2.5">
      <div class="max-w-7xl mx-auto flex items-center justify-center gap-2">
        <svg class="h-4 w-4 text-red-500 flex-shrink-0" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
        </svg>
        <p class="text-sm text-red-700 font-medium">
          {{ t('common.apiDisconnected', 'API server is not reachable. Make sure the gateway is running on port 8080.') }}
        </p>
      </div>
    </div>
  </Transition>
</template>
