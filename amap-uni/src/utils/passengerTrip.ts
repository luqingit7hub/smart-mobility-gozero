import type { OrderNotifyEvent, ActiveOrder } from '@/types'
import { orderApi } from '@/api/user'
import { connectOrderWs } from '@/utils/orderWs'

const ACTIVE_PASSENGER_ORDER_KEY = 'activeOrder'
const PASSENGER_ORDER_CHANGE = 'activePassengerOrderChange'

type PassengerOrderListener = (evt: OrderNotifyEvent) => void

let wsClose: (() => void) | null = null
const wsListeners = new Set<PassengerOrderListener>()

export function normalizePassengerOrder(raw: Record<string, unknown>): ActiveOrder {
  return {
    orderNo: String(raw.orderNo ?? raw.order_no ?? ''),
    start: String(raw.start ?? raw.startAddress ?? raw.start_address ?? ''),
    end: String(raw.end ?? raw.endAddress ?? raw.end_address ?? ''),
    price: Number(raw.price ?? 0),
    distance: Number(raw.distance ?? 0) || undefined,
    duration: Number(raw.duration ?? 0) || undefined,
    text: raw.text ? String(raw.text) : undefined,
    waitStartedAt: Number(raw.waitStartedAt ?? raw.wait_started_at ?? 0) || undefined,
    status: Number(raw.status ?? 0) || undefined,
    driverId: Number(raw.driverId ?? raw.driver_id ?? 0) || undefined,
    driverName: String(raw.driverName ?? raw.driver_name ?? '') || undefined,
    carNumber: String(raw.carNumber ?? raw.car_number ?? '') || undefined,
    carType: String(raw.carType ?? raw.car_type ?? '') || undefined,
    driverRating: Number(raw.driverRating ?? raw.driver_rating ?? raw.rating ?? 0) || undefined,
    startLng: Number(raw.startLng ?? raw.start_lng ?? 0) || undefined,
    startLat: Number(raw.startLat ?? raw.start_lat ?? 0) || undefined,
    endLng: Number(raw.endLng ?? raw.end_lng ?? 0) || undefined,
    endLat: Number(raw.endLat ?? raw.end_lat ?? 0) || undefined,
  }
}

export function getActivePassengerOrder(): ActiveOrder | null {
  try {
    const raw = sessionStorage.getItem(ACTIVE_PASSENGER_ORDER_KEY)
    if (!raw) return null
    const order = normalizePassengerOrder(JSON.parse(raw) as Record<string, unknown>)
    return order.orderNo ? order : null
  } catch {
    return null
  }
}

export function setActivePassengerOrder(order: ActiveOrder | Record<string, unknown> | null) {
  if (order) {
    const normalized = normalizePassengerOrder(order as Record<string, unknown>)
    sessionStorage.setItem(ACTIVE_PASSENGER_ORDER_KEY, JSON.stringify(normalized))
  } else {
    sessionStorage.removeItem(ACTIVE_PASSENGER_ORDER_KEY)
    sessionStorage.removeItem('takeCarRaw')
  }
  window.dispatchEvent(new CustomEvent(PASSENGER_ORDER_CHANGE))
}

export function orderStatusLabel(status?: number): string {
  if (status === 5) return '行程进行中'
  if (status === 2) return '司机已接单'
  if (status === 1) return '等待接单'
  return '进行中'
}

async function fetchOngoingFromServer(): Promise<ActiveOrder | null> {
  try {
    const res = await orderApi.ongoingOrder({ silent: true })
    const raw = res?.order as Record<string, unknown> | undefined
    const hasOrder = Boolean(
      res?.hasOrder ?? (res as { has_order?: boolean }).has_order ?? raw?.orderNo ?? raw?.order_no,
    )
    if (!hasOrder || !raw) return null
    const order = normalizePassengerOrder(raw)
    if (!order.waitStartedAt) order.waitStartedAt = Date.now()
    return order.orderNo ? order : null
  } catch {
    return null
  }
}

function isFreshWaitingOrder(order: ActiveOrder): boolean {
  const status = order.status
  const isWaiting = status === 1 || status === undefined || status === 0
  if (!isWaiting) return false
  if (!order.waitStartedAt) return true
  return Date.now() - order.waitStartedAt < 120_000
}

export function isFreshPassengerWaitingOrder(order: ActiveOrder | null | undefined): boolean {
  return !!order?.orderNo && isFreshWaitingOrder(order)
}

/** 优先服务端；无进行中单时清除过期缓存（完单/取消后不再误用本地） */
export async function resolveActivePassengerOrder(): Promise<ActiveOrder | null> {
  const cached = getActivePassengerOrder()
  const server = await fetchOngoingFromServer()
  if (server) {
    const merged: ActiveOrder = {
      ...server,
      waitStartedAt: cached?.orderNo === server.orderNo ? cached.waitStartedAt : server.waitStartedAt,
    }
    setActivePassengerOrder(merged)
    return merged
  }
  if (cached && isFreshWaitingOrder(cached)) {
    return cached
  }
  if (cached) {
    setActivePassengerOrder(null)
  }
  return null
}

/** 以服务端为准；刚下单 60 秒内若服务端尚未查到，保留本地待接单缓存 */
export async function syncActivePassengerOrderFromServer(): Promise<ActiveOrder | null> {
  const cached = getActivePassengerOrder()
  const server = await fetchOngoingFromServer()
  if (server) {
    const merged: ActiveOrder = {
      ...server,
      waitStartedAt: cached?.orderNo === server.orderNo ? cached.waitStartedAt : server.waitStartedAt,
    }
    setActivePassengerOrder(merged)
    return merged
  }
  if (cached && isFreshWaitingOrder(cached)) {
    return cached
  }
  setActivePassengerOrder(null)
  return null
}

export function subscribePassengerOrderChange(handler: () => void) {
  const onCustom = () => handler()
  const onStorage = (e: StorageEvent) => {
    if (e.key === ACTIVE_PASSENGER_ORDER_KEY) handler()
  }
  window.addEventListener(PASSENGER_ORDER_CHANGE, onCustom)
  window.addEventListener('storage', onStorage)
  return () => {
    window.removeEventListener(PASSENGER_ORDER_CHANGE, onCustom)
    window.removeEventListener('storage', onStorage)
  }
}

export function subscribePassengerOrderWs(listener: PassengerOrderListener) {
  ensurePassengerOrderWs()
  wsListeners.add(listener)
  return () => wsListeners.delete(listener)
}

export function ensurePassengerOrderWs() {
  if (wsClose) return
  const token = localStorage.getItem('passenger_token')
  if (!token || localStorage.getItem('role') !== 'passenger') return
  wsClose = connectOrderWs('passenger', token, (evt) => {
    wsListeners.forEach((fn) => fn(evt))
  })
}

export function stopPassengerOrderWs() {
  wsClose?.()
  wsClose = null
}

export function applyDriverAcceptedToOrder(evt: OrderNotifyEvent, order: ActiveOrder): ActiveOrder {
  return {
    ...order,
    status: 2,
    driverId: evt.driver_id ?? order.driverId,
    driverName: evt.driver_name || order.driverName,
    carNumber: evt.car_number || order.carNumber,
    carType: evt.car_type || order.carType,
    driverRating: evt.rating ?? order.driverRating,
  }
}

export function applyTripStartedToOrder(order: ActiveOrder): ActiveOrder {
  return { ...order, status: 5 }
}

export function passengerPhaseFromStatus(status?: number): 'waiting' | 'accepted' | 'inTrip' {
  if (status === 5) return 'inTrip'
  if (status === 2) return 'accepted'
  return 'waiting'
}

export function buildActiveOrderFromTakeCar(
  orderNo: string,
  start: string,
  end: string,
  res: { price?: number; distance?: number; duration?: number; text?: string },
  coords?: { startLng?: number; startLat?: number; endLng?: number; endLat?: number },
): ActiveOrder {
  return {
    orderNo,
    start,
    end,
    price: res.price ?? 0,
    distance: res.distance,
    duration: res.duration,
    text: res.text,
    status: 1,
    waitStartedAt: Date.now(),
    startLng: coords?.startLng,
    startLat: coords?.startLat,
    endLng: coords?.endLng,
    endLat: coords?.endLat,
  }
}
