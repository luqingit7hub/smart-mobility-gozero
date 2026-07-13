import type { GrabOrderItem, GrabOrderResp } from '@/types'
import { driverApi } from '@/api/driver'

const ACTIVE_DRIVER_ORDER_KEY = 'activeDriverOrder'

export function normalizeGrabOrder(raw: Record<string, unknown>): GrabOrderItem {
  return {
    orderNo: String(raw.orderNo ?? raw.order_no ?? ''),
    userId: Number(raw.userId ?? raw.user_id ?? 0),
    startLng: Number(raw.startLng ?? raw.start_lng ?? 0),
    startLat: Number(raw.startLat ?? raw.start_lat ?? 0),
    startAddress: String(raw.startAddress ?? raw.start_address ?? ''),
    endLng: Number(raw.endLng ?? raw.end_lng ?? 0),
    endLat: Number(raw.endLat ?? raw.end_lat ?? 0),
    endAddress: String(raw.endAddress ?? raw.end_address ?? ''),
    distance: Number(raw.distance ?? 0),
    duration: Number(raw.duration ?? 0),
    price: Number(raw.price ?? 0),
    expiresAt: Number(raw.expiresAt ?? raw.expires_at ?? 0),
    distanceToDriver: Number(raw.distanceToDriver ?? raw.distance_to_driver ?? 0),
    status: Number(raw.status ?? 0) || undefined,
  }
}

/** Go protobuf json omitempty 会省略 code=0，需兼容判断 */
export function isGrabSuccess(res: GrabOrderResp & Record<string, unknown>): boolean {
  const rawCode = res.code ?? res.Code
  const code = rawCode === undefined || rawCode === null ? undefined : Number(rawCode)
  const msg = String(res.msg ?? res.Msg ?? '')
  const orderNo = String(res.orderNo ?? res.order_no ?? '')

  if (code === 0) return true
  if (msg.includes('成功')) return true
  if (code === undefined && orderNo) return true
  return false
}

export function getActiveDriverTrip(): GrabOrderItem | null {
  try {
    const raw = sessionStorage.getItem(ACTIVE_DRIVER_ORDER_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw) as Record<string, unknown>
    const order = normalizeGrabOrder(parsed)
    return order.orderNo ? order : null
  } catch {
    return null
  }
}

export function setActiveDriverTrip(order: GrabOrderItem | Record<string, unknown> | null) {
  if (order) {
    const normalized = normalizeGrabOrder(order as Record<string, unknown>)
    sessionStorage.setItem(ACTIVE_DRIVER_ORDER_KEY, JSON.stringify(normalized))
  } else {
    sessionStorage.removeItem(ACTIVE_DRIVER_ORDER_KEY)
  }
  window.dispatchEvent(new CustomEvent('activeDriverOrderChange'))
}

/** 只读服务端进行中订单，不修改本地缓存 */
export async function fetchOngoingTripFromServer(): Promise<GrabOrderItem | null> {
  try {
    const res = await driverApi.ongoingOrder({ silent: true })
    const raw = res?.order as Record<string, unknown> | undefined
    const hasOrder = Boolean(res?.hasOrder ?? raw?.orderNo ?? raw?.order_no)
    if (hasOrder && raw) {
      const order = normalizeGrabOrder(raw)
      return order.orderNo ? order : null
    }
    return null
  } catch {
    return null
  }
}

/**
 * 解析进行中订单：优先服务端，抢单后 Stream 落库有延迟时保留本地缓存。
 * 用于抢单大厅 / 进行中页展示。
 */
export async function resolveActiveDriverTrip(): Promise<GrabOrderItem | null> {
  const cached = getActiveDriverTrip()
  const server = await fetchOngoingTripFromServer()
  if (server) {
    setActiveDriverTrip(server)
    return server
  }
  return cached
}

/** 以服务端为准同步（完单、重登后清本地缓存） */
export async function syncActiveDriverTripFromServer(): Promise<GrabOrderItem | null> {
  const server = await fetchOngoingTripFromServer()
  if (server) {
    setActiveDriverTrip(server)
    return server
  }
  setActiveDriverTrip(null)
  return null
}

export function isOrderAlreadyCompletedError(err: unknown): boolean {
  const msg = err instanceof Error ? err.message : String(err ?? '')
  return msg.includes('已经完成') || msg.includes('已完成')
}

export function isRequestTimeoutError(err: unknown): boolean {
  const msg = err instanceof Error ? err.message : String(err ?? '')
  return (
    msg.includes('timeout') ||
    msg.includes('Timeout') ||
    msg.includes('deadline exceeded') ||
    msg.includes('DeadlineExceeded')
  )
}

/** 完单请求超时后：若服务端已无进行中单，视为完单成功 */
export async function recoverOrderOverAfterTimeout(
  orderNo: string,
): Promise<'completed' | 'still_active' | 'unknown'> {
  const trip = await fetchOngoingTripFromServer()
  if (!trip) return 'completed'
  if (trip.orderNo === orderNo) return 'still_active'
  return 'unknown'
}

export function finishDriverTrip(router: { replace: (path: string) => void }) {
  setActiveDriverTrip(null)
  router.replace('/driver')
}
