import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { DeliverySlot, Shipment, ShipmentStatus } from '@/types'
import * as deliveryApi from '@/api/delivery'

export const useDeliveryStore = defineStore('delivery', () => {
  const slots = ref<DeliverySlot[]>([])
  const currentShipment = ref<Shipment | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchSlots(deliveryZoneId: string, date?: string) {
    loading.value = true
    error.value = null
    try {
      slots.value = await deliveryApi.getDeliverySlots({ delivery_zone_id: deliveryZoneId, date })
    } catch (e: unknown) {
      error.value = (e as Error).message
      slots.value = []
    } finally {
      loading.value = false
    }
  }

  async function reserveSlot(slotId: string, orderId: string) {
    return await deliveryApi.reserveDeliverySlot({ slot_id: slotId, order_id: orderId })
  }

  async function fetchShipment(shipmentId: string) {
    loading.value = true
    error.value = null
    try {
      currentShipment.value = await deliveryApi.getShipment(shipmentId)
    } catch (e: unknown) {
      error.value = (e as Error).message
    } finally {
      loading.value = false
    }
  }

  async function updateShipmentStatus(shipmentId: string, status: ShipmentStatus, description: string) {
    await deliveryApi.updateShipmentStatus(shipmentId, { status, description })
    if (currentShipment.value && currentShipment.value.id === shipmentId) {
      currentShipment.value.status = status
    }
  }

  return {
    slots, currentShipment, loading, error,
    fetchSlots, reserveSlot, fetchShipment, updateShipmentStatus,
  }
})
