import { mapApi } from '@/api/map'
import { saveDriverLocation, savePassengerLocation } from '@/utils/geo'

export const DRIVER_LOCATION_ADDRESS_KEY = 'driver_location_address'
export const DRIVER_LOCATION_UPDATED = 'driverLocationUpdated'
export const PASSENGER_LOCATION_UPDATED = 'passengerLocationUpdated'

export interface OrderEndLocation {
  endLng?: number
  endLat?: number
  endAddress?: string
}

function hasValidCoords(lng?: number, lat?: number) {
  return Number.isFinite(lng) && Number.isFinite(lat) && !(lng === 0 && lat === 0)
}

async function resolveEndCoords(
  loc: OrderEndLocation,
  geocodeType: 1 | 2,
): Promise<{ lng: number; lat: number; address: string } | null> {
  let { endLng, endLat, endAddress } = loc
  if (!hasValidCoords(endLng, endLat) && endAddress?.trim()) {
    try {
      const res = await mapApi.getCoordinates(endAddress.trim(), geocodeType)
      if (hasValidCoords(res.lng, res.lat)) {
        return {
          lng: res.lng,
          lat: res.lat,
          address: res.address || endAddress.trim(),
        }
      }
    } catch {
      return null
    }
    return null
  }
  if (!hasValidCoords(endLng, endLat)) return null

  let address = endAddress?.trim() || ''
  if (!address) {
    try {
      const res = await mapApi.reverseGeocode(endLng!, endLat!)
      address = res.address || '目的地'
    } catch {
      address = '目的地'
    }
  }
  return { lng: endLng!, lat: endLat!, address }
}

/** 完单后将司机接单位置同步为订单终点 */
export async function syncDriverLocationToOrderEnd(loc: OrderEndLocation) {
  const resolved = await resolveEndCoords(loc, 2)
  if (!resolved) return
  saveDriverLocation({ lng: resolved.lng, lat: resolved.lat, address: resolved.address })
  localStorage.setItem(DRIVER_LOCATION_ADDRESS_KEY, resolved.address)
  window.dispatchEvent(new CustomEvent(DRIVER_LOCATION_UPDATED, { detail: resolved }))
}

/** 完单后将乘客叫车起点缓存同步为订单终点 */
export async function syncPassengerLocationToOrderEnd(loc: OrderEndLocation) {
  const resolved = await resolveEndCoords(loc, 1)
  if (!resolved) return
  savePassengerLocation({ lng: resolved.lng, lat: resolved.lat, address: resolved.address })
  window.dispatchEvent(new CustomEvent(PASSENGER_LOCATION_UPDATED, { detail: resolved }))
}
