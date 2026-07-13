export type UserRole = 'passenger' | 'driver'

export type AiChatType = 1 | 2

export interface MapChatResult {
  answer: string
}

export interface AiChatMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  loading?: boolean
}

export interface ApiResponse<T = unknown> {
  code: number
  msg?: string
  message?: string
  data: T
}

export interface JourneyResult {
  price: number
  distance: number
  duration: number
  status: number
  waittine?: number
  startAdcode?: string
  start_adcode?: string
  startCityName?: string
  start_city_name?: string
  startLng?: number
  start_lat?: number
  start_lng?: number
  startLat?: number
  endLng?: number
  end_lat?: number
  end_lng?: number
  endLat?: number
  routePoints?: RoutePoint[]
  route_points?: RoutePoint[]
}

export interface RoutePoint {
  lng: number
  lat: number
}

export interface TakeCarResult {
  status: number | string
  price: number
  distance: number
  duration: number
  text: string
  orderNo?: string
  order_no?: string
}

export interface ActiveOrder {
  orderNo: string
  start: string
  end: string
  price: number
  distance?: number
  duration?: number
  text?: string
  waitStartedAt?: number
  status?: number
  driverId?: number
  driverName?: string
  carNumber?: string
  carType?: string
  driverRating?: number
  startLng?: number
  startLat?: number
  endLng?: number
  endLat?: number
}

export interface PassengerOngoingResp {
  hasOrder?: boolean
  order?: Record<string, unknown>
}

export interface CouponItem {
  id: number
  type: number
  moneyQuan?: number
  discount?: number
  cityCode?: string
  outTime?: string
}

export interface OrderNotifyEvent {
  event: string
  order_no?: string
  user_id?: number
  driver_id?: number
  driver_name?: string
  car_number?: string
  car_type?: string
  rating?: number
  accept_at?: number
  msg?: string
  pushed_driver_count?: number
  push_radius_km?: number
  start_address?: string
  price?: number
  distance?: number
}

export interface GrabOrderItem {
  orderNo: string
  userId: number
  startLng: number
  startLat: number
  startAddress: string
  endLng: number
  endLat: number
  endAddress: string
  distance: number
  duration: number
  price: number
  expiresAt: number
  distanceToDriver: number
  /** 2已接单 5用户已上车 */
  status?: number
}

export interface GrabOrderResp {
  code: number
  msg: string
  orderNo?: string
}

export interface GeocodeResult {
  address: string
  lng: number
  lat: number
}

export interface ReverseGeocodeResult {
  address: string
  lng: number
  lat: number
}

export interface WalletBalanceResult {
  balance: number
}

export interface WalletLogItem {
  id: number
  orderNo?: string
  order_no?: string
  amount: number
  signedAmount?: number
  signed_amount?: number
  direction?: 'in' | 'out'
  balanceBefore?: number
  balance_before?: number
  balanceAfter?: number
  balance_after?: number
  type: number
  typeName?: string
  type_name?: string
  status: number
  statusName?: string
  status_name?: string
  remark?: string
  createdAt?: string
  created_at?: string
}

export interface WalletLogsResult {
  list: WalletLogItem[]
  total: number
  page: number
  page_size?: number
  pageSize?: number
}

export const COMPANY_USER_ID = 999

/** 1待接单 2已接单 3已完成 4已取消 5用户已上车 */
export type OrderStatus = 1 | 2 | 3 | 4 | 5

export const ORDER_STATUS_ON_BOARD = 5
export const ORDER_STATUS_ACCEPTED = 2

export interface OrderListItem {
  id: number
  orderNo?: string
  order_no?: string
  userId?: number
  user_id?: number
  driverId?: number
  driver_id?: number
  startAddress?: string
  start_address?: string
  endAddress?: string
  end_address?: string
  distance?: number
  duration?: number
  price?: number
  payType?: number
  pay_type?: number
  status?: number
  statusName?: string
  status_name?: string
  cancelReason?: string
  cancel_reason?: string
  acceptTime?: string
  accept_time?: string
  createdAt?: string
  created_at?: string
}

export interface OrderListResult {
  list: OrderListItem[]
  total: number
  page?: number
  page_size?: number
  pageSize?: number
}

export interface OrderRatingItem {
  id?: number
  orderNo?: string
  order_no?: string
  userId?: number
  user_id?: number
  driverId?: number
  driver_id?: number
  rating?: number
  comment?: string
  tags?: string
  createdAt?: string
  created_at?: string
}

/** 实名/资质认证状态：status 1正常 2未实名 3禁用 */
export interface RealNameStatus {
  verified?: boolean
  status?: number
  statusName?: string
  status_name?: string
  realName?: string
  real_name?: string
  nickname?: string
  avatar?: string
  carNumber?: string
  car_number?: string
  carType?: string
  car_type?: string
  carColor?: string
  car_color?: string
}
