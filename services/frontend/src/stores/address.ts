import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Address, AddAddressRequest, UpdateAddressRequest } from '@/types'
import * as usersApi from '@/api/users'

export const useAddressStore = defineStore('address', () => {
  const addresses = ref<Address[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchAddresses() {
    loading.value = true
    error.value = null
    try {
      addresses.value = await usersApi.listMyAddresses()
    } catch (e: unknown) {
      error.value = (e as Error).message
    } finally {
      loading.value = false
    }
  }

  async function addAddress(data: AddAddressRequest) {
    const res = await usersApi.addAddress(data)
    await fetchAddresses()
    return res
  }

  async function updateAddress(addressId: string, data: UpdateAddressRequest) {
    const addr = await usersApi.updateAddress(addressId, data)
    await fetchAddresses()
    return addr
  }

  async function deleteAddress(addressId: string) {
    await usersApi.deleteAddress(addressId)
    await fetchAddresses()
  }

  return { addresses, loading, error, fetchAddresses, addAddress, updateAddress, deleteAddress }
})
