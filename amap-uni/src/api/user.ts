import { postForm, postRaw, uploadForm } from './request'
import type { AxiosRequestConfig } from 'axios'
import type { CouponItem, JourneyResult, OrderListResult, OrderRatingItem, PassengerOngoingResp, TakeCarResult } from '@/types'
import { pickOrderNoDeep } from '@/utils/order'

export const userApi = {
  sendSms: (phone: string, config?: AxiosRequestConfig) =>
    postForm<{ status: string }>('/user/sms', { phone }, config),

  register: (data: { phone: string; password: string; code: string }) =>
    postForm<{ id: number }>('/user/register', data),

  login: (data: {
    phone: string
    type: number
    password?: string
    code?: string
    lng?: number
    lat?: number
  }) => postForm<string>('/user/login', data),

  realName: (formData: FormData) =>
    uploadForm<{ msg: string }>('/user/auth/real/name', formData),

  realNameStatus: (config?: AxiosRequestConfig) =>
    postForm<import('@/types').RealNameStatus>('/user/auth/real/name/status', {}, config),

  listCoupons: (config?: AxiosRequestConfig) =>
    postForm<{ list: CouponItem[]; count: number }>('/user/auth/list/coupons', {}, config),

  recharge: (money: number) =>
    postForm<{ alipayUrl?: string; alipay_url?: string }>('/user/auth/alipay/balance', { money }),

  walletBalance: (config?: AxiosRequestConfig) =>
    postForm<{ balance: number }>('/user/auth/wallet/balance', {}, config),

  walletLogs: (data?: { page?: number; page_size?: number; order_no?: string }, config?: AxiosRequestConfig) =>
    postForm<{ list: import('@/types').WalletLogItem[]; total: number; page: number; page_size: number }>(
      '/user/auth/wallet/logs',
      data || {},
      config,
    ),

  rateOrder: (data: {
    order_no: string
    rating: number
    comment?: string
    tags?: string
  }) =>
    postForm<{ msg?: string; ratingId?: number; rating_id?: number; driverRating?: number; driver_rating?: number }>(
      '/user/auth/rate/order',
      {
        order_no: data.order_no,
        rating: data.rating,
        comment: data.comment,
        tags: data.tags,
      },
    ),

  orderList: (
    data?: { page?: number; page_size?: number; order_no?: string; status?: number },
    config?: AxiosRequestConfig,
  ) => postForm<OrderListResult>('/user/auth/order/list', data || {}, config),

  orderRating: (order_no: string, config?: AxiosRequestConfig) =>
    postForm<OrderRatingItem>('/user/auth/order/rating', { order_no }, config),

  deleteOrderRating: (order_no: string) =>
    postForm<{ msg?: string; driverRating?: number; driver_rating?: number }>(
      '/user/auth/order/rating/delete',
      { order_no },
    ),
}

export const orderApi = {
  journey: (starting_point: string, destination: string) =>
    postForm<JourneyResult>('/order/auth/journey', { starting_point, destination }),

  takeCar: async (data: { starting_point: string; destination: string; tid?: number }) => {
    const body = await postRaw<TakeCarResult>('/order/auth/take/car', data)
    const orderNo = pickOrderNoDeep(body)
    const result = (body.data || {}) as TakeCarResult
    return { ...result, orderNo: orderNo || result.orderNo || result.order_no || '' }
  },

  cancelOrder: (order_no: string, reason?: string) =>
    postForm<{ msg: string }>('/order/auth/cancel/order', { order_no, reason }),

  ongoingOrder: (config?: AxiosRequestConfig) =>
    postForm<PassengerOngoingResp>('/order/auth/user/ongoing', {}, config),
}
