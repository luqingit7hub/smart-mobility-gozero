import type { WalletLogItem } from '@/types'

export function pickWalletLogOrderNo(item: WalletLogItem): string {
  return item.orderNo || item.order_no || ''
}

export function pickWalletLogSignedAmount(item: WalletLogItem): number {
  if (item.signedAmount != null) return item.signedAmount
  if (item.signed_amount != null) return item.signed_amount
  return item.amount
}

export function pickWalletLogTypeName(item: WalletLogItem): string {
  return item.typeName || item.type_name || '其他'
}

export function pickWalletLogStatusName(item: WalletLogItem): string {
  return item.statusName || item.status_name || ''
}

export function pickWalletLogCreatedAt(item: WalletLogItem): string {
  return item.createdAt || item.created_at || ''
}

export function pickWalletLogBalanceAfter(item: WalletLogItem): number | undefined {
  const v = item.balanceAfter ?? item.balance_after
  return v == null ? undefined : v
}

export function normalizeWalletLogs(list: WalletLogItem[] = []): WalletLogItem[] {
  return list.map((item) => ({
    ...item,
    orderNo: pickWalletLogOrderNo(item),
    signedAmount: pickWalletLogSignedAmount(item),
    typeName: pickWalletLogTypeName(item),
    statusName: pickWalletLogStatusName(item),
    createdAt: pickWalletLogCreatedAt(item),
    balanceAfter: pickWalletLogBalanceAfter(item),
    direction: item.direction || (pickWalletLogSignedAmount(item) >= 0 ? 'in' : 'out'),
  }))
}

export function formatWalletAmount(amount: number): string {
  const sign = amount >= 0 ? '+' : ''
  return `${sign}¥${Math.abs(amount).toFixed(2)}`
}
