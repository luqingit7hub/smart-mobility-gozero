import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { UserRole } from '@/types'
import { COMPANY_USER_ID } from '@/types'
import { clearPassengerLocation, clearDriverLocation } from '@/utils/geo'

function parseUserIdFromToken(token: string): number {
  try {
    const part = token.split('.')[1]
    if (!part) return 0
    const json = JSON.parse(
      decodeURIComponent(
        atob(part.replace(/-/g, '+').replace(/_/g, '/'))
          .split('')
          .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
          .join(''),
      ),
    ) as { userId?: string }
    return Number(json.userId ?? 0)
  } catch {
    return 0
  }
}

export const useAuthStore = defineStore('auth', () => {
  const role = ref<UserRole | null>((localStorage.getItem('role') as UserRole) || null)
  const passengerToken = ref(localStorage.getItem('passenger_token') || '')
  const driverToken = ref(localStorage.getItem('driver_token') || '')
  const phone = ref(localStorage.getItem('phone') || '')

  const token = computed(() =>
    role.value === 'driver' ? driverToken.value : passengerToken.value,
  )

  const userId = computed(() => (token.value ? parseUserIdFromToken(token.value) : 0))
  const isCompanyUser = computed(() => userId.value === COMPANY_USER_ID)

  function setPassengerAuth(t: string, p: string) {
    role.value = 'passenger'
    passengerToken.value = t
    driverToken.value = ''
    phone.value = p
    localStorage.setItem('role', 'passenger')
    localStorage.setItem('passenger_token', t)
    localStorage.removeItem('driver_token')
    localStorage.setItem('phone', p)
  }

  function setDriverAuth(t: string, p: string) {
    role.value = 'driver'
    driverToken.value = t
    passengerToken.value = ''
    phone.value = p
    localStorage.setItem('role', 'driver')
    localStorage.setItem('driver_token', t)
    localStorage.removeItem('passenger_token')
    localStorage.setItem('phone', p)
  }

  function logout() {
    if (role.value === 'driver') {
      driverToken.value = ''
      localStorage.removeItem('driver_token')
      clearDriverLocation()
    } else {
      passengerToken.value = ''
      localStorage.removeItem('passenger_token')
      clearPassengerLocation()
    }
    role.value = null
    phone.value = ''
    localStorage.removeItem('role')
    localStorage.removeItem('phone')
  }

  return {
    role,
    token,
    phone,
    userId,
    isCompanyUser,
    setPassengerAuth,
    setDriverAuth,
    logout,
  }
})
