import axios, { type AxiosInstance, type AxiosError, type InternalAxiosRequestConfig } from 'axios'
import { useAuthStore } from '@/stores/auth'
import router from '@/router'

const apiClient: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

let isRefreshing = false
let failedQueue: Array<{ resolve: (token: string) => void; reject: (error: any) => void }> = []

const processQueue = (error: any, token: string | null = null) => {
  failedQueue.forEach(({ resolve, reject }) => {
    if (token) {
      resolve(token)
    } else {
      reject(error)
    }
  })
  failedQueue = []
}

function getCsrfTokenFromCookie(): string {
  const match = document.cookie.match(/csrf_token=([^;]+)/)
  return match ? match[1] : ''
}

apiClient.interceptors.request.use(
  (config) => {
    const authStore = useAuthStore()
    if (authStore.accessToken) {
      config.headers.Authorization = `Bearer ${authStore.accessToken}`
    }

    const method = config.method?.toUpperCase()
    if (method === 'POST' || method === 'PUT' || method === 'DELETE' || method === 'PATCH') {
      const csrfToken = getCsrfTokenFromCookie()
      if (csrfToken) {
        config.headers['X-CSRF-Token'] = csrfToken
      }
    }

    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

let skipAuthInterceptor = false

apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    if (skipAuthInterceptor) {
      return Promise.reject(error)
    }

    if (error.response?.status === 403) {
      const errorMessage = (error.response.data as any)?.message || ''
      if (errorMessage.includes('csrf_token')) {
        router.push({ name: 'Login' })
        return Promise.reject(error)
      }
    }

    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean }

    if (error.response?.status === 401 && !originalRequest._retry) {
      if (isRefreshing) {
        return new Promise<string>((resolve, reject) => {
          failedQueue.push({ resolve, reject })
        })
          .then((token) => {
            originalRequest.headers.Authorization = `Bearer ${token}`
            return apiClient(originalRequest)
          })
          .catch((err) => Promise.reject(err))
      }

      originalRequest._retry = true
      isRefreshing = true

      const authStore = useAuthStore()

      try {
        const success = await authStore.refreshAccessToken()

        if (success) {
          const newToken = authStore.accessToken!
          processQueue(null, newToken)
          originalRequest.headers.Authorization = `Bearer ${newToken}`
          return apiClient(originalRequest)
        } else {
          processQueue(error, null)
          authStore.logout()
          router.push({ name: 'Login' })
          return Promise.reject(error)
        }
      } catch {
        processQueue(error, null)
        authStore.logout()
        router.push({ name: 'Login' })
        return Promise.reject(error)
      } finally {
        isRefreshing = false
      }
    }

    return Promise.reject(error)
  }
)

export function setSkipAuthInterceptor(value: boolean) {
  skipAuthInterceptor = value
}

export default apiClient
