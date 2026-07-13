import type { OrderNotifyEvent } from '@/types'
import { normalizeOrderEvent } from '@/utils/orderEvent'

export type OrderWsHandler = (event: OrderNotifyEvent) => void

export interface OrderWsOptions {
  onOpen?: () => void
  onClose?: () => void
}

function wsBaseUrl(): string {
  const envUrl = import.meta.env.VITE_WS_BASE_URL
  if (envUrl?.trim()) return envUrl.replace(/\/$/, '')
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${proto}//${location.host}/ws`
}

/** 乘客 /ws/user  司机 /ws/driver */
export function connectOrderWs(
  role: 'passenger' | 'driver',
  token: string,
  onMessage: OrderWsHandler,
  options: OrderWsOptions = {},
): () => void {
  const path = role === 'driver' ? '/driver' : '/user'
  const url = `${wsBaseUrl()}${path}?token=${encodeURIComponent(token)}`
  let ws: WebSocket | null = null
  let closed = false
  let retryTimer: ReturnType<typeof setTimeout> | null = null
  let retryCount = 0

  const connect = () => {
    if (closed) return
    ws = new WebSocket(url)
    ws.onopen = () => {
      retryCount = 0
      options.onOpen?.()
    }
    ws.onmessage = (ev) => {
      try {
        const raw = JSON.parse(ev.data as string) as Record<string, unknown>
        onMessage(normalizeOrderEvent(raw))
      } catch {
        /* ignore malformed payload */
      }
    }
    ws.onclose = () => {
      options.onClose?.()
      if (closed) return
      const delay = Math.min(1000 * 2 ** retryCount, 15000)
      retryCount++
      retryTimer = setTimeout(connect, delay)
    }
    ws.onerror = () => ws?.close()
  }

  connect()

  return () => {
    closed = true
    if (retryTimer) clearTimeout(retryTimer)
    ws?.close()
    ws = null
  }
}
