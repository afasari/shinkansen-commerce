<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useProductStore } from '@/stores/product'
import type { Money } from '@/types'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const productStore = useProductStore()

const isEdit = computed(() => route.name === 'admin-product-edit')
const saving = ref(false)

const name = ref('')
const description = ref('')
const categoryId = ref('')
const priceUnits = ref(0)
const sku = ref('')
const stockQuantity = ref(100)
const imageUrls = ref('')
const active = ref(true)

onMounted(async () => {
  if (isEdit.value) {
    await productStore.fetchProduct(route.params.id as string)
    const p = productStore.currentProduct
    if (p) {
      name.value = p.name
      description.value = p.description
      categoryId.value = p.category_id
      priceUnits.value = p.price?.units || 0
      sku.value = p.sku
      stockQuantity.value = p.stock_quantity
      imageUrls.value = (p.image_urls || []).join('\n')
      active.value = p.active
    }
  }
})

async function handleSave() {
  saving.value = true
  try {
    const images = imageUrls.value.split('\n').map((u) => u.trim()).filter(Boolean)
    const price: Money = { currency: 'JPY', units: priceUnits.value, nanos: 0 }

    if (isEdit.value) {
      await productStore.updateProduct(route.params.id as string, {
        name: name.value,
        description: description.value,
        category_id: categoryId.value,
        price,
        active: active.value,
        image_urls: images,
      })
    } else {
      await productStore.createProduct({
        name: name.value,
        description: description.value,
        category_id: categoryId.value,
        price,
        sku: sku.value,
        stock_quantity: stockQuantity.value,
        image_urls: images,
      })
    }
    router.push({ name: 'admin-products' })
  } catch (e: unknown) {
    alert((e as Error).message)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-gray-900">{{ isEdit ? t('product.editProduct') : t('product.createProduct') }}</h1>
    </div>

    <div class="card p-6 max-w-2xl">
      <form @submit.prevent="handleSave" class="space-y-4">
        <div>
          <label class="label-field">{{ t('product.productName') }} *</label>
          <input v-model="name" required class="input-field mt-1" />
        </div>
        <div>
          <label class="label-field">{{ t('product.productDescription') }}</label>
          <textarea v-model="description" rows="4" class="input-field mt-1"></textarea>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="label-field">{{ t('product.category') }} ID</label>
            <input v-model="categoryId" class="input-field mt-1" />
          </div>
          <div>
            <label class="label-field">{{ t('common.price') }} (JPY) *</label>
            <input v-model.number="priceUnits" type="number" required class="input-field mt-1" />
          </div>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="label-field">{{ t('product.sku') }}</label>
            <input v-model="sku" class="input-field mt-1" :disabled="isEdit" />
          </div>
          <div>
            <label class="label-field">{{ t('product.stockQuantity') }}</label>
            <input v-model.number="stockQuantity" type="number" class="input-field mt-1" :disabled="isEdit" />
          </div>
        </div>
        <div>
          <label class="label-field">{{ t('product.imageUrls') }} ({{ t('common.optional') }}, {{ t('common.onePerLine') }})</label>
          <textarea v-model="imageUrls" rows="3" class="input-field mt-1" placeholder="https://example.com/image1.jpg"></textarea>
        </div>
        <label class="flex items-center gap-2">
          <input type="checkbox" v-model="active" class="rounded border-gray-300 text-shinkansen-600" />
          <span class="text-sm">{{ t('product.active') }}</span>
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
