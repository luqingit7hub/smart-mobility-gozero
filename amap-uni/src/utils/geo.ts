export {
  couponTypeLabels,
  formatCouponValue,
  formatDiscountLabel,
  formatExpireTime,
  calcDiscountedPrice,
  normalizeCoupon,
  normalizeCoupons,
  isCouponExpired,
  cityLevelAdcode,
  matchCouponCityCode,
  annotateCouponsForOrder,
} from './coupon'
export type { CouponWithStatus } from './coupon'

export const trafficLabels: Record<number, string> = {
  0: '无路况',
  1: '畅通',
  2: '缓行',
  3: '拥堵',
  4: '严重拥堵',
}

const PI = Math.PI
const X_PI = (PI * 3000) / 180
const A = 6378245
const EE = 0.006693421622965943

function outOfChina(lng: number, lat: number) {
  return lng < 72.004 || lng > 137.8347 || lat < 0.8293 || lat > 55.8271
}

function transformLat(lng: number, lat: number) {
  let ret =
    -100 + 2 * lng + 3 * lat + 0.2 * lat * lat + 0.1 * lng * lat + 0.2 * Math.sqrt(Math.abs(lng))
  ret += ((20 * Math.sin(6 * lng * PI) + 20 * Math.sin(2 * lng * PI)) * 2) / 3
  ret += ((20 * Math.sin(lat * PI) + 40 * Math.sin((lat / 3) * PI)) * 2) / 3
  ret += ((160 * Math.sin((lat / 12) * PI) + 320 * Math.sin((lat * PI) / 30)) * 2) / 3
  return ret
}

function transformLng(lng: number, lat: number) {
  let ret = 300 + lng + 2 * lat + 0.1 * lng * lng + 0.1 * lng * lat + 0.1 * Math.sqrt(Math.abs(lng))
  ret += ((20 * Math.sin(6 * lng * PI) + 20 * Math.sin(2 * lng * PI)) * 2) / 3
  ret += ((20 * Math.sin(lng * PI) + 40 * Math.sin((lng / 3) * PI)) * 2) / 3
  ret += ((150 * Math.sin((lng / 12) * PI) + 300 * Math.sin((lng / 30) * PI)) * 2) / 3
  return ret
}

/** WGS84（浏览器 GPS）→ GCJ02 */
export function wgs84ToGcj02(lng: number, lat: number) {
  if (outOfChina(lng, lat)) return { lng, lat }
  let dLat = transformLat(lng - 105, lat - 35)
  let dLng = transformLng(lng - 105, lat - 35)
  const radLat = (lat / 180) * PI
  let magic = Math.sin(radLat)
  magic = 1 - EE * magic * magic
  const sqrtMagic = Math.sqrt(magic)
  dLat = (dLat * 180) / (((A * (1 - EE)) / (magic * sqrtMagic)) * PI)
  dLng = (dLng * 180) / ((A / sqrtMagic) * Math.cos(radLat) * PI)
  return { lng: lng + dLng, lat: lat + dLat }
}

/** GCJ02 → BD09（百度地图坐标系） */
export function gcj02ToBd09(lng: number, lat: number) {
  const z = Math.sqrt(lng * lng + lat * lat) + 0.00002 * Math.sin(lat * X_PI)
  const theta = Math.atan2(lat, lng) + 0.000003 * Math.cos(lng * X_PI)
  return {
    lng: z * Math.cos(theta) + 0.0065,
    lat: z * Math.sin(theta) + 0.006,
  }
}

/** WGS84 → BD09，供百度地图与后端逆地理编码使用 */
export function wgs84ToBd09(lng: number, lat: number) {
  const gcj = wgs84ToGcj02(lng, lat)
  return gcj02ToBd09(gcj.lng, gcj.lat)
}

export function getCurrentPosition(): Promise<GeolocationPosition> {
  return new Promise((resolve, reject) => {
    if (!navigator.geolocation) {
      reject(new Error('浏览器不支持定位'))
      return
    }
    navigator.geolocation.getCurrentPosition(resolve, reject, {
      enableHighAccuracy: true,
      timeout: 15000,
      maximumAge: 0,
    })
  })
}

export async function getLngLat(): Promise<{ lng: number; lat: number } | null> {
  try {
    const pos = await getCurrentPosition()
    const { longitude: lng, latitude: lat } = pos.coords
    if (!Number.isFinite(lng) || !Number.isFinite(lat)) return null
    return { lng, lat }
  } catch {
    return null
  }
}

/** 浏览器定位并转为百度 BD09 坐标 */
export async function getBd09LngLat(): Promise<{ lng: number; lat: number } | null> {
  const wgs = await getLngLat()
  if (!wgs) return null
  return wgs84ToBd09(wgs.lng, wgs.lat)
}

const PASSENGER_LOCATION_KEY = 'passenger_location_cache'

export interface PassengerLocationCache {
  lng: number
  lat: number
  address: string
}

export function savePassengerLocation(data: PassengerLocationCache) {
  if (!Number.isFinite(data.lng) || !Number.isFinite(data.lat)) return
  if (data.lng === 0 && data.lat === 0) return
  localStorage.setItem(
    PASSENGER_LOCATION_KEY,
    JSON.stringify({
      lng: data.lng,
      lat: data.lat,
      address: data.address || '',
    }),
  )
}

export function loadPassengerLocation(): PassengerLocationCache | null {
  try {
    const raw = localStorage.getItem(PASSENGER_LOCATION_KEY)
    if (!raw) return null
    const data = JSON.parse(raw) as PassengerLocationCache
    if (!Number.isFinite(data.lng) || !Number.isFinite(data.lat)) return null
    if (data.lng === 0 && data.lat === 0) return null
    return { lng: data.lng, lat: data.lat, address: data.address || '' }
  } catch {
    return null
  }
}

export function clearPassengerLocation() {
  localStorage.removeItem(PASSENGER_LOCATION_KEY)
}

const DRIVER_LOCATION_KEY = 'driver_location_cache'

export interface DriverLocationCache {
  lng: number
  lat: number
  address: string
}

export function saveDriverLocation(data: DriverLocationCache) {
  if (!Number.isFinite(data.lng) || !Number.isFinite(data.lat)) return
  if (data.lng === 0 && data.lat === 0) return
  localStorage.setItem(
    DRIVER_LOCATION_KEY,
    JSON.stringify({
      lng: data.lng,
      lat: data.lat,
      address: data.address || '',
    }),
  )
}

export function loadDriverLocation(): DriverLocationCache | null {
  try {
    const raw = localStorage.getItem(DRIVER_LOCATION_KEY)
    if (!raw) return null
    const data = JSON.parse(raw) as DriverLocationCache
    if (!Number.isFinite(data.lng) || !Number.isFinite(data.lat)) return null
    if (data.lng === 0 && data.lat === 0) return null
    return { lng: data.lng, lat: data.lat, address: data.address || '' }
  } catch {
    return null
  }
}

export function clearDriverLocation() {
  localStorage.removeItem(DRIVER_LOCATION_KEY)
  localStorage.removeItem('driver_location_address')
}
