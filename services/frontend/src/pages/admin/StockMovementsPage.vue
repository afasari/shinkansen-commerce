<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { useInventoryStore } from '@/stores/inventory'
import { MovementType, MOVEMENT_TYPE_LIST } from '@/utils/constants'
import { formatDateTime } from '@/utils/format'
import AppPagination from '@/components/common/AppPagination.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'

const { t } = useI18n()
const route = useRoute()
const inventoryStore = useInventoryStore()

const stockItemId = ref((route.params.id as string) || '')
const page = ref(1)
const totalPages = ref(1)

watch(() => inventoryStore.movementPagination, (pag) => {
  totalPages.value = Math.ceil(pag.total / pag.limit) || 1
}, { deep: true })

function loadMovements() {
  if (!stockItemId.value) return
  inventoryStore.fetchMovements(stockItemId.value, page.value)
}

function goToPage(p: number) {
  page.value = p
  loadMovements()
}

const movementLabels: Record<string, string> = {
  [MovementType.INBOUND]: t('inventory.inbound'),
  [MovementType.OUTBOUND]: t('inventory.outbound'),
  [MovementType.RESERVATION]: t('inventory.reservation'),
  [MovementType.RELEASE]: t('inventory.release'),
  [MovementType.ADJUSTMENT]: t('inventory.adjustment'),
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-6">{{ t('inventory.movements') }}</h1>

    <div class="card p-4 mb-6">
      <form @submit.prevent="loadMovements" class="flex gap-3 items-end">
        <div class="flex-1">
          <label class="label-field">Stock Item ID</label>
          <input v-model="stockItemId" required class="input-field mt-1" />
        </div>
        <button type="submit" class="btn-primary text-sm">{{ t('common.search') }}</button>
      </form>
    </div>

    <div v-if="inventoryStore.loading" class="text-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-shinkansen-200 border-t-shinkansen-600 mx-auto"></div>
    </div>

    <div v-else-if="inventoryStore.movements.length > 0" class="card overflow-hidden">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">ID</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('inventory.movementType') }}</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('common.quantity') }}</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('inventory.reference') }}</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('common.date') }}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100">
          <tr v-for="m in inventoryStore.movements" :key="m.id" class="hover:bg-gray-50">
            <td class="px-4 py-3 text-sm text-gray-500 font-mono">{{ m.id.substring(0, 8) }}...</td>
            <td class="px-4 py-3 text-sm">
              <span :class="[
                m.type === MovementType.INBOUND ? 'bg-green-100 text-green-800' :
                m.type === MovementType.OUTBOUND ? 'bg-red-100 text-red-800' :
                'bg-blue-100 text-blue-800',
                'px-2 py-0.5 rounded-full text-xs font-medium'
              ]">{{ movementLabels[m.type] || m.type }}</span>
            </td>
            <td class="px-4 py-3 text-sm font-medium" :class="m.quantity > 0 ? 'text-green-600' : 'text-red-600'">{{ m.quantity > 0 ? '+' : '' }}{{ m.quantity }}</td>
            <td class="px-4 py-3 text-sm text-gray-500">{{ m.reference || '-' }}</td>
            <td class="px-4 py-3 text-sm text-gray-500">{{ formatDateTime(m.created_at) }}</td>
          </tr>
        </tbody>
      </table>
      <AppPagination :current-page="page" :total-pages="totalPages" @page-change="goToPage" />
    </div>
  </div>
</template>
