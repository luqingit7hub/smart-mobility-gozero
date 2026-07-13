import type { ApiResponse } from '@/types'

export function pickOrderNo(obj: Record<string, unknown>): string {
  const v = obj.orderNo ?? obj.order_no ?? obj.OrderNo
  return typeof v === 'string' ? v : ''
}

export function pickOrderNoDeep(body: ApiResponse<unknown> | Record<string, unknown>): string {
  if (!body || typeof body !== 'object') return ''
  const data = (body as ApiResponse<unknown>).data ?? body
  if (typeof data === 'string') return data
  if (data && typeof data === 'object') {
    const no = pickOrderNo(data as Record<string, unknown>)
    if (no) return no
  }
  return pickOrderNo(body as Record<string, unknown>)
}

export function parseOrderNoFromText(text: string): string {
  const m = text.match(/\d{10,}/)
  return m?.[0] ?? ''
}

export function shortenOrderNo(no: string): string {
  if (no.length <= 12) return no
  return `${no.slice(0, 6)}…${no.slice(-4)}`
}

export function formatPrice(price?: number) {
  const n = Number(price ?? 0)
  const fixed = n.toFixed(2)
  const [integer, decimal] = fixed.split('.')
  return { symbol: '¥', integer, decimal: decimal ?? '00' }
}
