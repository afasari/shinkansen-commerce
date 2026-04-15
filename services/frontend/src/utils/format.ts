import type { Money } from '@/types'

export function formatPrice(money: Money | undefined | null): string {
  if (!money) return '¥0'
  const units = money.units ?? 0
  const nanos = money.nanos ?? 0
  const amount = units + nanos / 1_000_000_000
  return `¥${Math.round(amount).toLocaleString()}`
}

export function formatPriceFromUnits(units: number, currency = 'JPY'): string {
  return `¥${Math.round(units).toLocaleString()}`
}

export function formatDate(dateStr: string | undefined | null): string {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' })
}

export function formatDateTime(dateStr: string | undefined | null): string {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function formatTime(dateStr: string | undefined | null): string {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })
}

export function moneyToNumber(money: Money | undefined | null): number {
  if (!money) return 0
  return (money.units ?? 0) + (money.nanos ?? 0) / 1_000_000_000
}

export function createMoney(units: number, currency = 'JPY'): Money {
  return { currency, units: Math.round(units), nanos: 0 }
}

export function truncate(str: string, length: number): string {
  if (str.length <= length) return str
  return str.substring(0, length) + '...'
}
