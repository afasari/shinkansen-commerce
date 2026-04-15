import { client } from './client'
import type { Address, AddAddressRequest, UpdateAddressRequest } from '@/types'

export async function listMyAddresses(): Promise<Address[]> {
  const res = await client.get<{ addresses: Address[] }>('/v1/users/me/addresses')
  return res.data.addresses
}

export async function addAddress(data: AddAddressRequest): Promise<{ address_id: string }> {
  const res = await client.post<{ address_id: string }>('/v1/users/me/addresses', data)
  return res.data
}

export async function updateAddress(addressId: string, data: UpdateAddressRequest): Promise<Address> {
  const res = await client.put<{ address: Address }>(`/v1/addresses/${addressId}`, data)
  return res.data.address
}

export async function deleteAddress(addressId: string): Promise<void> {
  await client.delete(`/v1/addresses/${addressId}`)
}
