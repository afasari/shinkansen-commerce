import { OrderStatus, PaymentMethod, PaymentStatus, ShipmentStatus, MovementType } from '@/types'
export { OrderStatus, PaymentMethod, PaymentStatus, ShipmentStatus, MovementType } from '@/types'

export const ORDER_STATUS_LIST = Object.values(OrderStatus).filter((s) => s !== OrderStatus.UNSPECIFIED)

export const PAYMENT_METHOD_LIST = Object.values(PaymentMethod).filter((m) => m !== PaymentMethod.UNSPECIFIED)

export const PAYMENT_STATUS_LIST = Object.values(PaymentStatus).filter((s) => s !== PaymentStatus.UNSPECIFIED)

export const SHIPMENT_STATUS_LIST = Object.values(ShipmentStatus).filter((s) => s !== ShipmentStatus.UNSPECIFIED)

export const MOVEMENT_TYPE_LIST = Object.values(MovementType).filter((t) => t !== MovementType.UNSPECIFIED)

export const ORDER_STATUS_COLORS: Record<string, string> = {
  [OrderStatus.PENDING]: 'bg-yellow-100 text-yellow-800',
  [OrderStatus.CONFIRMED]: 'bg-blue-100 text-blue-800',
  [OrderStatus.PROCESSING]: 'bg-indigo-100 text-indigo-800',
  [OrderStatus.SHIPPED]: 'bg-purple-100 text-purple-800',
  [OrderStatus.IN_TRANSIT]: 'bg-violet-100 text-violet-800',
  [OrderStatus.DELIVERED]: 'bg-green-100 text-green-800',
  [OrderStatus.CANCELLED]: 'bg-red-100 text-red-800',
  [OrderStatus.EXPIRED]: 'bg-gray-100 text-gray-800',
  [OrderStatus.READY_FOR_PICKUP]: 'bg-teal-100 text-teal-800',
  [OrderStatus.PICKED_UP]: 'bg-emerald-100 text-emerald-800',
  [OrderStatus.FAILED_DELIVERY]: 'bg-orange-100 text-orange-800',
  [OrderStatus.RETURNED]: 'bg-pink-100 text-pink-800',
}

export const PAYMENT_STATUS_COLORS: Record<string, string> = {
  [PaymentStatus.PENDING]: 'bg-yellow-100 text-yellow-800',
  [PaymentStatus.PROCESSING]: 'bg-blue-100 text-blue-800',
  [PaymentStatus.COMPLETED]: 'bg-green-100 text-green-800',
  [PaymentStatus.FAILED]: 'bg-red-100 text-red-800',
  [PaymentStatus.CANCELLED]: 'bg-gray-100 text-gray-800',
  [PaymentStatus.REFUNDED]: 'bg-purple-100 text-purple-800',
}

export const SHIPMENT_STATUS_COLORS: Record<string, string> = {
  [ShipmentStatus.PREPARING]: 'bg-yellow-100 text-yellow-800',
  [ShipmentStatus.SHIPPED]: 'bg-blue-100 text-blue-800',
  [ShipmentStatus.IN_TRANSIT]: 'bg-purple-100 text-purple-800',
  [ShipmentStatus.DELIVERED]: 'bg-green-100 text-green-800',
  [ShipmentStatus.CANCELLED]: 'bg-red-100 text-red-800',
  [ShipmentStatus.FAILED_DELIVERY]: 'bg-orange-100 text-orange-800',
}

export const ORDER_STATUS_TRANSITIONS: Record<string, OrderStatus[]> = {
  [OrderStatus.PENDING]: [OrderStatus.CONFIRMED, OrderStatus.CANCELLED, OrderStatus.EXPIRED],
  [OrderStatus.CONFIRMED]: [OrderStatus.PROCESSING, OrderStatus.CANCELLED],
  [OrderStatus.PROCESSING]: [OrderStatus.SHIPPED, OrderStatus.READY_FOR_PICKUP],
  [OrderStatus.SHIPPED]: [OrderStatus.IN_TRANSIT],
  [OrderStatus.IN_TRANSIT]: [OrderStatus.DELIVERED, OrderStatus.FAILED_DELIVERY],
  [OrderStatus.READY_FOR_PICKUP]: [OrderStatus.PICKED_UP, OrderStatus.CANCELLED],
  [OrderStatus.PICKED_UP]: [OrderStatus.DELIVERED],
  [OrderStatus.FAILED_DELIVERY]: [OrderStatus.RETURNED, OrderStatus.DELIVERED],
  [OrderStatus.DELIVERED]: [OrderStatus.RETURNED],
}

export const CANCELLABLE_STATUSES = [OrderStatus.PENDING, OrderStatus.CONFIRMED, OrderStatus.READY_FOR_PICKUP]

export const DEFAULT_WAREHOUSE_ID = 'default-warehouse-001'
export const DEFAULT_DELIVERY_ZONE_ID = '00000000-0000-0000-0000-000000000001'

export const PREFECTURES = [
  '北海道', '青森県', '岩手県', '宮城県', '秋田県', '山形県', '福島県',
  '茨城県', '栃木県', '群馬県', '埼玉県', '千葉県', '東京都', '神奈川県',
  '新潟県', '富山県', '石川県', '福井県', '山梨県', '長野県',
  '岐阜県', '静岡県', '愛知県', '三重県',
  '滋賀県', '京都府', '大阪府', '兵庫県', '奈良県', '和歌山県',
  '鳥取県', '島根県', '岡山県', '広島県', '山口県',
  '徳島県', '香川県', '愛媛県', '高知県',
  '福岡県', '佐賀県', '長崎県', '熊本県', '大分県', '宮崎県', '鹿児島県', '沖縄県',
]

export function generateSessionId(): string {
  return 'session-' + crypto.randomUUID()
}
