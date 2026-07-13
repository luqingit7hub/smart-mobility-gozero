import { postForm } from './request'
import type { GeocodeResult, ReverseGeocodeResult } from '@/types'

export const mapApi = {
  getCoordinates: (address: string, type: 1 | 2) =>
    postForm<GeocodeResult>('/map/auth/get/coordinates', { address, type }),

  reverseGeocode: (lng: number, lat: number) =>
    postForm<ReverseGeocodeResult>('/map/auth/reverse/geocode', { lng, lat }),
}
