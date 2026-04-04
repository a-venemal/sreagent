import axios from 'axios'
import type { ApiResponse } from '@/types'

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' },
})

// Map backend error codes to user-friendly messages
// Key is the error code, values are [zh, en]
const errorMessages: Record<number, [string, string]> = {
  10102: ['用户名或密码错误', 'Invalid username or password'],
  10101: ['登录已过期，请重新登录', 'Session expired, please log in again'],
  10100: ['未授权访问', 'Unauthorized'],
  10200: ['权限不足', 'Insufficient permissions'],
  10300: ['资源不存在', 'Resource not found'],
  10400: ['名称已被占用', 'Name already taken'],
}

function getLocale(): string {
  try { return localStorage.getItem('locale') || 'zh-CN' } catch { return 'zh-CN' }
}

function localizeError(code: number, fallback: string): string {
  const msgs = errorMessages[code]
  if (!msgs) return fallback
  return getLocale() === 'zh-CN' ? msgs[0] : msgs[1]
}

// Request interceptor - attach JWT token
request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) config.headers.Authorization = `Bearer ${token}`
    return config
  },
  (error) => Promise.reject(error)
)

// Prevent multiple simultaneous 401 redirects
let isRedirecting = false

// Response interceptor
request.interceptors.response.use(
  (response) => {
    const data = response.data as ApiResponse
    if (data.code !== 0) {
      const msg = localizeError(data.code, data.message || 'Unknown error')
      return Promise.reject(new Error(msg))
    }
    return response
  },
  (error) => {
    if (error.response?.status === 401 && !isRedirecting) {
      isRedirecting = true
      localStorage.removeItem('token')
      localStorage.removeItem('user_role')
      // Use Vue Router-style navigation to avoid full page reload
      // Dynamically import to avoid circular dependency
      import('@/router').then(({ default: router }) => {
        router.push({ name: 'Login', query: { redirect: router.currentRoute.value.fullPath } })
      }).finally(() => {
        // Reset flag after a short delay so subsequent 401s can still trigger
        setTimeout(() => { isRedirecting = false }, 2000)
      })
      return Promise.reject(error)
    }
    const data = error.response?.data as ApiResponse | undefined
    const code = data?.code || 0
    const fallback = data?.message || error.message || 'Network error'
    return Promise.reject(new Error(localizeError(code, fallback)))
  }
)

export default request
