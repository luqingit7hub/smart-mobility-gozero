import type { AxiosRequestConfig } from 'axios'
import { postForm, uploadForm } from './request'
import { mapApi as sharedMapApi } from './map'
import type { GrabOrderItem, GrabOrderResp, OrderListResult, OrderRatingItem } from '@/types'

export const GRAB_LIST_RADIUS_M = 5000

export const driverApi = {
  sendSms: (phone: string, config?: AxiosRequestConfig) =>
    postForm<{ status: string }>('/driver/sms', { phone }, config),

  register: (data: { phone: string; password: string; code: string }) =>
    postForm<{ id: number }>('/driver/register', data),

  login: (data: {
    phone: string
    type: number
    password?: string
    code?: string
    lng?: number
    lat?: number
  }) => postForm<string>('/driver/login', data),

  realName: (formData: FormData) =>
    uploadForm<{ msg: string }>('/driver/auth/real/name', formData),

  realNameStatus: (config?: AxiosRequestConfig) =>
    postForm<import('@/types').RealNameStatus>('/driver/auth/real/name/status', {}, config),

  offline: (config?: AxiosRequestConfig) =>
    postForm<{ msg: string }>('/driver/auth/offline', {}, config),

  ongoingOrder: (config?: AxiosRequestConfig) =>
    postForm<{ hasOrder?: boolean; order?: GrabOrderItem }>(
      '/order/auth/driver/ongoing',
      {},
      config,
    ),

  grabList: (radius = GRAB_LIST_RADIUS_M, limit = 20) =>
    postForm<{ orders: GrabOrderItem[] }>('/order/auth/grab/list', { radius, limit }),

  grabOrder: (order_no: string) =>
    postForm<GrabOrderResp>('/order/auth/grab/order', { order_no }),

  orderOver: (order_no: string, config?: AxiosRequestConfig) =>
    postForm<{ status: string }>('/order/auth/order/over', { order_no }, {
      timeout: 60000,
      ...config,
    }),

  startOrder: (order_no: string, phone_tail: string, config?: AxiosRequestConfig) =>
    postForm<{ status: string; msg?: string }>(
      '/order/auth/start/order',
      { order_no, phone_tail },
      config,
    ),

  walletBalance: (config?: AxiosRequestConfig) =>
    postForm<{ balance: number }>('/driver/auth/wallet/balance', {}, config),

  walletLogs: (data?: { page?: number; page_size?: number; order_no?: string }, config?: AxiosRequestConfig) =>
    postForm<{ list: import('@/types').WalletLogItem[]; total: number; page: number; page_size: number }>(
      '/driver/auth/wallet/logs',
      data || {},
      config,
    ),

  orderList: (
    data?: { page?: number; page_size?: number; order_no?: string; status?: number },
    config?: AxiosRequestConfig,
  ) => postForm<OrderListResult>('/driver/auth/order/list', data || {}, config),

  orderRating: (order_no: string, config?: AxiosRequestConfig) =>
    postForm<OrderRatingItem>('/driver/auth/order/rating', { order_no }, config),
}

const AI_CHAT_TIMEOUT = 90000

export const mapApi = {
  chat: (question: string, type: 1 | 2, role: 'passenger' | 'driver', config?: AxiosRequestConfig) => {
    const url = role === 'passenger' ? '/user/auth/map/chat' : '/driver/auth/map/chat'
    return postForm<{ answer: string }>(url, { question, type }, { timeout: AI_CHAT_TIMEOUT, ...config })
  },

  getCoordinates: sharedMapApi.getCoordinates,

  reverseGeocode: sharedMapApi.reverseGeocode,

  issueCoupons: (data: {
    address: string
    type: number
    out_time: string
    money_quan?: number
    discount?: number
  }) => postForm<{ issued_count?: number; issuedCount?: number }>('/map/auth/issue/coupons', data),
}
