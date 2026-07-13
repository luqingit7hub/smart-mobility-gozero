import type { ActiveOrder, OrderNotifyEvent } from '@/types'

export interface CompletedTrip {
  orderNo: string
  start?: string
  end?: string
  price?: number
  driverName?: string
  carNumber?: string
  msg?: string
}

const COMPLETED_TRIP_KEY = 'completedTrip'

export function saveCompletedTrip(trip: CompletedTrip) {
  sessionStorage.setItem(COMPLETED_TRIP_KEY, JSON.stringify(trip))
}

export function getCompletedTrip(): CompletedTrip | null {
  try {
    const raw = sessionStorage.getItem(COMPLETED_TRIP_KEY)
    if (!raw) return null
    return JSON.parse(raw) as CompletedTrip
  } catch {
    return null
  }
}

export function clearCompletedTrip() {
  sessionStorage.removeItem(COMPLETED_TRIP_KEY)
}

/** 将完单 WS 通知与本地行程缓存合并为行程结束页数据 */
export function buildCompletedTripFromNotify(
  evt: OrderNotifyEvent,
  trip?: ActiveOrder | null,
): CompletedTrip {
  return {
    orderNo: trip?.orderNo || evt.order_no || '',
    start: trip?.start,
    end: trip?.end,
    price: trip?.price,
    driverName: trip?.driverName || evt.driver_name,
    carNumber: trip?.carNumber || evt.car_number,
    msg: evt.msg,
  }
}
