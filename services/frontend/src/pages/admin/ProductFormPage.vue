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
const error = ref('')

const name = ref('')
const description = ref('')
const categoryId = ref('')
const priceUnits = ref<string>('')
const sku = ref('')
const stockQuantity = ref<string>('100')
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
      priceUnits.value = String(p.price?.units ?? '')
      sku.value = p.sku
      stockQuantity.value = String(p.stock_quantity)
      imageUrls.value = (p.image_urls || []).join('\n')
      active.value = p.active
    }
  }
})

async function handleSave() {
  error.value = ''
  if (!name.value.trim()) {
    error.value = 'Product name is required'
    return
  }
  const price = Number(priceUnits.value)
  if (!priceUnits.value || isNaN(price) || price <= 0) {
    error.value = 'Price must be a positive number'
    return
  }

  saving.value = true
  try {
    const images = imageUrls.value.split('\n').map((u) => u.trim()).filter(Boolean)
    const money: Money = { currency: 'JPY', units: price }

    if (isEdit.value) {
      await productStore.updateProduct(route.params.id as string, {
        name: name.value,
        description: description.value,
        category_id: categoryId.value || undefined,
        price: money,
        active: active.value,
        image_urls: images.length > 0 ? images : undefined,
      })
    } else {
      await productStore.createProduct({
        name: name.value,
        description: description.value,
        category_id: categoryId.value || '',
        price: money,
        sku: sku.value || '',
        stock_quantity: Number(stockQuantity.value) || 0,
        image_urls: images,
      })
    }
    router.push({ name: 'admin-products' })
  } catch (e: unknown) {
    const err = e as any
    error.value = err.response?.data?.message || err.message || 'Failed to save product'
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
      <div v-if="error" class="rounded-md bg-red-50 p-3 mb-4">
        <p class="text-sm text-red-700">{{ error }}</p>
      </div>

      <form @submit.prevent="handleSave" class="space-y-4">
        <div>
          <label class="label-field">{{ t('product.productName') }} *</label>
          <input v-model="name" required class="input-field mt-1" placeholder="e.g. Shinkansen Bento Box" />
        </div>
        <div>
          <label class="label-field">{{ t('product.productDescription') }}</label>
          <textarea v-model="description" rows="4" class="input-field mt-1" placeholder="Product description..."></textarea>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="label-field">{{ t('product.category') }} ID</label>
            <input v-model="categoryId" class="input-field mt-1" placeholder="e.g. cat-food" />
          </div>
          <div>
            <label class="label-field">{{ t('common.price') }} (JPY) *</label>
            <input v-model="priceUnits" type="number" required min="1" class="input-field mt-1" placeholder="e.g. 1500" />
          </div>
        </div>
        <div v-if="!isEdit" class="grid grid-cols-2 gap-4">
          <div>
            <label class="label-field">{{ t('product.sku') }}</label>
            <input v-model="sku" class="input-field mt-1" placeholder="e.g. BENTO-001" />
          </div>
          <div>
            <label class="label-field">{{ t('product.stockQuantity') }}</label>
            <input v-model="stockQuantity" type="number" min="0" class="input-field mt-1" />
          </div>
        </div>
        <div>
          <label class="label-field">{{ t('product.imageUrls') }} (optional, one per line)</label>
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
