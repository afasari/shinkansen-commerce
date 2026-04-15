<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useInventoryStore } from '@/stores/inventory'
import { DEFAULT_WAREHOUSE_ID } from '@/utils/constants'

const { t } = useI18n()
const inventoryStore = useInventoryStore()

const productId = ref('')
const variantId = ref('')
const warehouseId = ref(DEFAULT_WAREHOUSE_ID)
const stockItemId = ref('')
const quantityDelta = ref(0)
const reason = ref('')
const reference = ref('')
const updating = ref(false)
const currentStock = ref<string | null>(null)

async function lookupStock() {
  if (!productId.value) return
  const stock = await inventoryStore.fetchStock({
    product_id: productId.value,
    variant_id: variantId.value || undefined,
    warehouse_id: warehouseId.value,
  })
  if (stock) {
    currentStock.value = `Qty: ${stock.quantity}, Reserved: ${stock.reserved_quantity}, Available: ${stock.available_quantity}`
    stockItemId.value = stock.id
  } else {
    currentStock.value = null
  }
}

async function handleUpdate() {
  if (!productId.value) return
  updating.value = true
  try {
    await inventoryStore.updateStock({
      stock_item_id: stockItemId.value,
      product_id: productId.value,
      variant_id: variantId.value,
      warehouse_id: warehouseId.value,
      quantity_delta: quantityDelta.value,
      reason: reason.value,
      reference: reference.value,
    })
    await lookupStock()
    quantityDelta.value = 0
    reason.value = ''
    reference.value = ''
  } catch (e: unknown) {
    alert((e as Error).message)
  } finally {
    updating.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-6">{{ t('inventory.title') }}</h1>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <div class="card p-6">
        <h2 class="font-semibold text-gray-900 mb-4">{{ t('inventory.stockLevel') }}</h2>
        <form @submit.prevent="lookupStock" class="space-y-3">
          <div>
            <label class="label-field">{{ t('inventory.productId') }} *</label>
            <input v-model="productId" required class="input-field mt-1" />
          </div>
          <div>
            <label class="label-field">{{ t('inventory.variantId') }}</label>
            <input v-model="variantId" class="input-field mt-1" />
          </div>
          <div>
            <label class="label-field">{{ t('inventory.warehouse') }}</label>
            <input v-model="warehouseId" class="input-field mt-1" />
          </div>
          <button type="submit" class="btn-primary text-sm">{{ t('common.search') }}</button>
        </form>
        <div v-if="currentStock" class="mt-4 p-3 bg-gray-50 rounded-lg text-sm text-gray-700">
          {{ currentStock }}
        </div>
      </div>

      <div class="card p-6">
        <h2 class="font-semibold text-gray-900 mb-4">{{ t('inventory.updateStock') }}</h2>
        <form @submit.prevent="handleUpdate" class="space-y-3">
          <div>
            <label class="label-field">{{ t('inventory.quantityDelta') }}</label>
            <input v-model.number="quantityDelta" type="number" class="input-field mt-1" placeholder="+10 or -5" />
          </div>
          <div>
            <label class="label-field">{{ t('inventory.reason') }}</label>
            <input v-model="reason" class="input-field mt-1" />
          </div>
          <div>
            <label class="label-field">{{ t('inventory.reference') }}</label>
            <input v-model="reference" class="input-field mt-1" />
          </div>
          <button type="submit" :disabled="updating || !stockItemId" class="btn-primary text-sm">
            <span v-if="updating" class="animate-spin mr-2 inline-block h-4 w-4 border-2 border-white border-t-transparent rounded-full"></span>
            {{ t('inventory.updateStock') }}
          </button>
        </form>
      </div>
    </div>
  </div>
</template>
