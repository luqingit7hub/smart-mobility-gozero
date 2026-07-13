import axios, { type AxiosRequestConfig } from 'axios'
import { API_BASE_URL } from '@/config/env'
import type { ApiResponse } from '@/types'
import { showMessage } from '@/utils/toast'

declare module 'axios' {
  interface AxiosRequestConfig {
    silent?: boolean
  }
}

const request = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
})

function extractMessage(data: unknown, fallback: string): string {
  if (data && typeof data === 'object') {
    const obj = data as ApiResponse
    const raw = obj.msg ?? obj.message
    if (typeof raw === 'string' && raw.trim()) return raw.trim()
  }
  return fallback
}

function notifyError(message: string, silent?: boolean) {
  const msg = message.trim() || '请求失败'
  if (!silent) showMessage(msg, 'danger')
  return msg
}

request.interceptors.request.use((config) => {
  const role = localStorage.getItem('role') as 'passenger' | 'driver' | null
  const tokenKey = role === 'driver' ? 'driver_token' : 'passenger_token'
  const token = localStorage.getItem(tokenKey)
  if (token) config.headers.token = token
  return config
})

request.interceptors.response.use(
  (response) => {
    const res = response.data as ApiResponse
    if (res?.code === 200) return response
    const msg = notifyError(extractMessage(res, '请求失败'), response.config.silent)
    return Promise.reject(new Error(msg))
  },
  (error) => {
    const msg = notifyError(
      extractMessage(error.response?.data, error.message || '网络错误'),
      error.config?.silent,
    )
    return Promise.reject(new Error(msg))
  },
)

export async function postForm<T>(
  url: string,
  data: Record<string, string | number | boolean | undefined> = {},
  config?: AxiosRequestConfig,
) {
  const params = new URLSearchParams()
  for (const [key, value] of Object.entries(data)) {
    if (value !== undefined && value !== null && value !== '') {
      params.append(key, String(value))
    }
  }
  const res = await request.post<ApiResponse<T>>(url, params, {
    ...config,
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
      ...config?.headers,
    },
  })
  return res.data.data
}

export async function postRaw<T>(
  url: string,
  data: Record<string, string | number | boolean | undefined> = {},
  config?: AxiosRequestConfig,
) {
  const params = new URLSearchParams()
  for (const [key, value] of Object.entries(data)) {
    if (value !== undefined && value !== null && value !== '') {
      params.append(key, String(value))
    }
  }
  const res = await request.post<ApiResponse<T>>(url, params, {
    ...config,
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
      ...config?.headers,
    },
  })
  return res.data
}

/** multipart/form-data 上传（实名认证等） */
export async function uploadForm<T>(url: string, formData: FormData, config?: AxiosRequestConfig) {
  const res = await request.post<ApiResponse<T>>(url, formData, {
    ...config,
    headers: {
      'Content-Type': 'multipart/form-data',
      ...config?.headers,
    },
  })
  return res.data.data
}

export default request
