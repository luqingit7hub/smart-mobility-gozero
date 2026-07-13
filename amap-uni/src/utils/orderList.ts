import type { OrderListItem, OrderRatingItem } from '@/types'

export const ORDER_STATUS_TABS = [
  { label: '全部', value: 0 },
  { label: '待接单', value: 1 },
  { label: '已接单', value: 2 },
  { label: '用户已上车', value: 5 },
  { label: '已完成', value: 3 },
  { label: '已取消', value: 4 },
] as const

export function pickOrderListNo(item: OrderListItem): string {
  return item.orderNo || item.order_no || ''
}

export function pickOrderListStart(item: OrderListItem): string {
  return item.startAddress || item.start_address || ''
}

export function pickOrderListEnd(item: OrderListItem): string {
  return item.endAddress || item.end_address || ''
}

export function pickOrderListStatusName(item: OrderListItem): string {
  return item.statusName || item.status_name || '未知'
}

export function pickOrderListCreatedAt(item: OrderListItem): string {
  return item.createdAt || item.created_at || ''
}

export function normalizeOrderList(list: OrderListItem[] = []): OrderListItem[] {
  return list.map((item) => ({
    ...item,
    orderNo: pickOrderListNo(item),
    startAddress: pickOrderListStart(item),
    endAddress: pickOrderListEnd(item),
    statusName: pickOrderListStatusName(item),
    createdAt: pickOrderListCreatedAt(item),
  }))
}

export function orderStatusTagType(status?: number): 'primary' | 'success' | 'warning' | 'danger' | 'default' {
  switch (status) {
    case 1:
      return 'warning'
    case 2:
      return 'primary'
    case 5:
      return 'success'
    case 3:
      return 'success'
    case 4:
      return 'default'
    default:
      return 'default'
  }
}

export function isOrderCompleted(status?: number): boolean {
  return status === 3
}

export function isNoRatingError(err: unknown): boolean {
  const msg = err instanceof Error ? err.message : String(err ?? '')
  return msg.includes('暂无评价')
}

export function normalizeOrderRating(data: OrderRatingItem): OrderRatingItem {
  return {
    ...data,
    orderNo: data.orderNo || data.order_no || '',
    createdAt: data.createdAt || data.created_at || '',
    rating: Number(data.rating ?? 0),
  }
}

export function parseRatingTags(tags?: string): string[] {
  if (!tags?.trim()) return []
  try {
    const arr = JSON.parse(tags) as unknown
    if (!Array.isArray(arr)) return []
    return arr.filter((t): t is string => typeof t === 'string' && t.trim().length > 0)
  } catch {
    return []
  }
}

export function buildOrderRatingQuery(item: OrderListItem): Record<string, string> {
  const query: Record<string, string> = {
    orderNo: pickOrderListNo(item),
  }
  if (item.status != null) query.status = String(item.status)
  const start = pickOrderListStart(item)
  const end = pickOrderListEnd(item)
  if (start) query.start = start
  if (end) query.end = end
  if (item.price != null) query.price = String(item.price)
  return query
}
