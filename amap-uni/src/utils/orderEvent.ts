import type { OrderNotifyEvent } from '@/types'

/** 统一 MQ/WebSocket 事件字段（兼容 snake_case 与 camelCase） */
export function normalizeOrderEvent(raw: Record<string, unknown>): OrderNotifyEvent {
  const orderNo = String(raw.order_no ?? raw.orderNo ?? '')
  return {
    event: String(raw.event ?? ''),
    order_no: orderNo || undefined,
    user_id: Number(raw.user_id ?? raw.userId ?? 0) || undefined,
    driver_id: Number(raw.driver_id ?? raw.driverId ?? 0) || undefined,
    driver_name: String(raw.driver_name ?? raw.driverName ?? '') || undefined,
    car_number: String(raw.car_number ?? raw.carNumber ?? '') || undefined,
    car_type: String(raw.car_type ?? raw.carType ?? '') || undefined,
    rating: Number(raw.rating ?? 0) || undefined,
    accept_at: Number(raw.accept_at ?? raw.acceptAt ?? 0) || undefined,
    msg: String(raw.msg ?? '') || undefined,
    pushed_driver_count: Number(raw.pushed_driver_count ?? raw.pushedDriverCount ?? 0) || undefined,
    push_radius_km: Number(raw.push_radius_km ?? raw.pushRadiusKm ?? 0) || undefined,
    start_address: String(raw.start_address ?? raw.startAddress ?? '') || undefined,
    price: Number(raw.price ?? 0) || undefined,
    distance: Number(raw.distance ?? 0) || undefined,
  }
}

export function eventOrderNo(evt: OrderNotifyEvent): string {
  return evt.order_no?.trim() || ''
}

export function isSameOrder(evt: OrderNotifyEvent, localOrderNo: string): boolean {
  const local = localOrderNo.trim()
  const remote = eventOrderNo(evt)
  if (!local || !remote) return true
  return local === remote
}
