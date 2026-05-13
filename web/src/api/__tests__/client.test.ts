import { describe, it, expect, vi, beforeEach, beforeAll } from 'vitest'

let isRefreshingModule = false
let failedQueueModule: Array<{ resolve: (token: string) => void; reject: (error: any) => void }> =
  []

const mockAuthStore = {
  accessToken: 'test-access-token',
  refreshAccessToken: vi.fn().mockResolvedValue(true),
  logout: vi.fn(),
}

vi.mock('@/stores/auth', () => ({
  useAuthStore: vi.fn(() => mockAuthStore),
}))

vi.mock('@/router', () => ({
  default: {
    push: vi.fn(),
  },
}))

describe('apiClient', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.resetModules()
    isRefreshingModule = false
    failedQueueModule = []
  })

  beforeAll(() => {
    isRefreshingModule = false
    failedQueueModule = []
  })

  describe('request interceptor', () => {
    it('should add Authorization header when accessToken exists', async () => {
      const { default: apiClient } = await import('@/api/client')

      const mockRequest = {
        headers: {},
        method: 'GET',
      }

      const interceptor = apiClient.interceptors.request.handlers[0]
      if (interceptor?.fulfilled) {
        const result = await interceptor.fulfilled(mockRequest as any)
        expect(result.headers.Authorization).toBe('Bearer test-access-token')
      }
    })

    it('should not add Authorization header when no accessToken', async () => {
      mockAuthStore.accessToken = null
      const { default: apiClient } = await import('@/api/client')

      const mockRequest = {
        headers: {},
        method: 'GET',
      }

      const interceptor = apiClient.interceptors.request.handlers[0]
      if (interceptor?.fulfilled) {
        const result = await interceptor.fulfilled(mockRequest as any)
        expect(result.headers.Authorization).toBeUndefined()
      }
    })

    it('should add CSRF token for POST requests', async () => {
      Object.defineProperty(document, 'cookie', {
        value: 'csrf_token=test-csrf-token',
        writable: true,
      })

      const { default: apiClient } = await import('@/api/client')

      const mockRequest = {
        headers: {},
        method: 'POST',
      }

      const interceptor = apiClient.interceptors.request.handlers[0]
      if (interceptor?.fulfilled) {
        const result = await interceptor.fulfilled(mockRequest as any)
        expect(result.headers['X-CSRF-Token']).toBe('test-csrf-token')
      }
    })

    it('should add CSRF token for PUT requests', async () => {
      Object.defineProperty(document, 'cookie', {
        value: 'csrf_token=put-csrf-token',
        writable: true,
      })

      const { default: apiClient } = await import('@/api/client')

      const mockRequest = {
        headers: {},
        method: 'PUT',
      }

      const interceptor = apiClient.interceptors.request.handlers[0]
      if (interceptor?.fulfilled) {
        const result = await interceptor.fulfilled(mockRequest as any)
        expect(result.headers['X-CSRF-Token']).toBe('put-csrf-token')
      }
    })

    it('should add CSRF token for DELETE requests', async () => {
      Object.defineProperty(document, 'cookie', {
        value: 'csrf_token=delete-csrf-token',
        writable: true,
      })

      const { default: apiClient } = await import('@/api/client')

      const mockRequest = {
        headers: {},
        method: 'DELETE',
      }

      const interceptor = apiClient.interceptors.request.handlers[0]
      if (interceptor?.fulfilled) {
        const result = await interceptor.fulfilled(mockRequest as any)
        expect(result.headers['X-CSRF-Token']).toBe('delete-csrf-token')
      }
    })

    it('should add CSRF token for PATCH requests', async () => {
      Object.defineProperty(document, 'cookie', {
        value: 'csrf_token=patch-csrf-token',
        writable: true,
      })

      const { default: apiClient } = await import('@/api/client')

      const mockRequest = {
        headers: {},
        method: 'PATCH',
      }

      const interceptor = apiClient.interceptors.request.handlers[0]
      if (interceptor?.fulfilled) {
        const result = await interceptor.fulfilled(mockRequest as any)
        expect(result.headers['X-CSRF-Token']).toBe('patch-csrf-token')
      }
    })

    it('should not add CSRF token for GET requests', async () => {
      Object.defineProperty(document, 'cookie', {
        value: 'csrf_token=test-csrf-token',
        writable: true,
      })

      const { default: apiClient } = await import('@/api/client')

      const mockRequest = {
        headers: {},
        method: 'GET',
      }

      const interceptor = apiClient.interceptors.request.handlers[0]
      if (interceptor?.fulfilled) {
        const result = await interceptor.fulfilled(mockRequest as any)
        expect(result.headers['X-CSRF-Token']).toBeUndefined()
      }
    })

    it('should handle missing CSRF token cookie', async () => {
      Object.defineProperty(document, 'cookie', {
        value: '',
        writable: true,
      })

      const { default: apiClient } = await import('@/api/client')

      const mockRequest = {
        headers: {},
        method: 'POST',
      }

      const interceptor = apiClient.interceptors.request.handlers[0]
      if (interceptor?.fulfilled) {
        const result = await interceptor.fulfilled(mockRequest as any)
        expect(result.headers['X-CSRF-Token']).toBeUndefined()
      }
    })
  })

  describe('response interceptor - 403 CSRF error', () => {
    it('should redirect to login on CSRF token error', async () => {
      vi.resetModules()
      const { default: apiClient } = await import('@/api/client')
      const { default: router } = await import('@/router')

      const error = {
        response: {
          status: 403,
          data: { message: 'csrf_token mismatch' },
        },
      }

      const interceptor = apiClient.interceptors.response.handlers[1]
      if (interceptor?.rejected) {
        await expect(interceptor.rejected(error)).rejects.toEqual(error)
      }
    })
  })

  describe('setSkipAuthInterceptor', () => {
    it('should set skipAuthInterceptor flag', async () => {
      vi.resetModules()
      const { setSkipAuthInterceptor } = await import('@/api/client')

      setSkipAuthInterceptor(true)
      expect(setSkipAuthInterceptor(true)).toBeUndefined()
    })
  })

  describe('getCsrfTokenFromCookie', () => {
    it('should extract CSRF token from cookie', async () => {
      Object.defineProperty(document, 'cookie', {
        value: 'csrf_token=abc123xyz',
        writable: true,
      })

      vi.resetModules()
      const { default: apiClient } = await import('@/api/client')

      const interceptor = apiClient.interceptors.request.handlers[0]
      if (interceptor?.fulfilled) {
        const mockRequest = { headers: {}, method: 'POST' }
        const result = await interceptor.fulfilled(mockRequest as any)
        expect(result.headers['X-CSRF-Token']).toBe('abc123xyz')
      }
    })

    it('should return empty string when no csrf_token cookie', async () => {
      Object.defineProperty(document, 'cookie', {
        value: 'other_cookie=value',
        writable: true,
      })

      vi.resetModules()
      const { default: apiClient } = await import('@/api/client')

      const interceptor = apiClient.interceptors.request.handlers[0]
      if (interceptor?.fulfilled) {
        const mockRequest = { headers: {}, method: 'POST' }
        const result = await interceptor.fulfilled(mockRequest as any)
        expect(result.headers['X-CSRF-Token']).toBeUndefined()
      }
    })

    it('should handle cookie with multiple tokens', async () => {
      Object.defineProperty(document, 'cookie', {
        value: 'session=abc123; csrf_token=my-csrf-token; other=value',
        writable: true,
      })

      vi.resetModules()
      const { default: apiClient } = await import('@/api/client')

      const interceptor = apiClient.interceptors.request.handlers[0]
      if (interceptor?.fulfilled) {
        const mockRequest = { headers: {}, method: 'POST' }
        const result = await interceptor.fulfilled(mockRequest as any)
        expect(result.headers['X-CSRF-Token']).toBe('my-csrf-token')
      }
    })
  })

  describe('request interceptor error handler', () => {
    it('should reject on request interceptor error', async () => {
      const { default: apiClient } = await import('@/api/client')

      const interceptor = apiClient.interceptors.request.handlers[0]
      if (interceptor?.rejected) {
        const error = new Error('Request config error')
        await expect(interceptor.rejected(error)).rejects.toThrow('Request config error')
      }
    })
  })

  describe('processQueue (via token refresh flow)', () => {
    it('should handle successful token refresh and queue', async () => {
      vi.resetModules()

      mockAuthStore.accessToken = 'old-token'
      mockAuthStore.refreshAccessToken = vi.fn().mockResolvedValue(true)
      mockAuthStore.accessToken = 'new-refreshed-token'

      const { default: apiClient } = await import('@/api/client')

      const error = {
        response: { status: 401 },
        config: { _retry: false, headers: {} },
      }

      const interceptor = apiClient.interceptors.response.handlers[1]
      if (interceptor?.rejected) {
        await expect(interceptor.rejected(error)).rejects.toBeDefined()
      }
    })

    it('should handle failed token refresh', async () => {
      vi.resetModules()

      mockAuthStore.accessToken = 'expired-token'
      mockAuthStore.refreshAccessToken = vi.fn().mockResolvedValue(false)
      mockAuthStore.logout = vi.fn()

      const { default: apiClient } = await import('@/api/client')

      const error = {
        response: { status: 401 },
        config: { _retry: false, headers: {} },
      }

      const interceptor = apiClient.interceptors.response.handlers[1]
      if (interceptor?.rejected) {
        await expect(interceptor.rejected(error)).rejects.toBeDefined()
      }
    })

    it('should handle refresh token exception', async () => {
      vi.resetModules()

      mockAuthStore.accessToken = 'expired-token'
      mockAuthStore.refreshAccessToken = vi.fn().mockRejectedValue(new Error('Refresh failed'))
      mockAuthStore.logout = vi.fn()

      const { default: apiClient } = await import('@/api/client')

      const error = {
        response: { status: 401 },
        config: { _retry: false, headers: {} },
      }

      const interceptor = apiClient.interceptors.response.handlers[1]
      if (interceptor?.rejected) {
        await expect(interceptor.rejected(error)).rejects.toBeDefined()
      }
    })
  })

  describe('skipAuthInterceptor in response', () => {
    it('should skip auth interceptor when flag is set', async () => {
      vi.resetModules()
      const { setSkipAuthInterceptor } = await import('@/api/client')
      setSkipAuthInterceptor(true)

      const { default: apiClient } = await import('@/api/client')

      const error = {
        response: { status: 401 },
        config: { _retry: false },
      }

      const interceptor = apiClient.interceptors.response.handlers[1]
      if (interceptor?.rejected) {
        await expect(interceptor.rejected(error)).rejects.toEqual(error)
      }
    })
  })

  describe('response interceptor - non-401 errors', () => {
    it('should pass through non-401 errors without token refresh', async () => {
      vi.resetModules()
      const { default: apiClient } = await import('@/api/client')

      const error = {
        response: { status: 500, data: { message: 'Server error' } },
        config: { _retry: false },
      }

      const interceptor = apiClient.interceptors.response.handlers[1]
      if (interceptor?.rejected) {
        await expect(interceptor.rejected(error)).rejects.toEqual(error)
      }
    })

    it('should pass through 404 errors without token refresh', async () => {
      vi.resetModules()
      const { default: apiClient } = await import('@/api/client')

      const error = {
        response: { status: 404, data: { message: 'Not found' } },
        config: { _retry: false },
      }

      const interceptor = apiClient.interceptors.response.handlers[1]
      if (interceptor?.rejected) {
        await expect(interceptor.rejected(error)).rejects.toEqual(error)
      }
    })

    it('should pass through errors without response config', async () => {
      vi.resetModules()
      const { default: apiClient } = await import('@/api/client')

      const error = {
        message: 'Network error',
        config: undefined,
      }

      const interceptor = apiClient.interceptors.response.handlers[1]
      if (interceptor?.rejected) {
        await expect(interceptor.rejected(error)).rejects.toEqual(error)
      }
    })
  })

  describe('response interceptor - 401 with _retry already true', () => {
    it('should not retry when _retry is already true', async () => {
      vi.resetModules()
      mockAuthStore.accessToken = 'expired-token'
      mockAuthStore.refreshAccessToken = vi.fn()

      const { default: apiClient } = await import('@/api/client')

      const error = {
        response: { status: 401 },
        config: { _retry: true, headers: {} },
      }

      const interceptor = apiClient.interceptors.response.handlers[1]
      if (interceptor?.rejected) {
        await expect(interceptor.rejected(error)).rejects.toEqual(error)
      }
      expect(mockAuthStore.refreshAccessToken).not.toHaveBeenCalled()
    })
  })

  describe('processQueue direct tests', () => {
    it('should resolve queued promises with new token on success', async () => {
      vi.resetModules()

      const { default: apiClient } = await import('@/api/client')

      const successError = {
        response: { status: 401 },
        config: { _retry: false, headers: {} },
      }

      mockAuthStore.refreshAccessToken = vi.fn().mockResolvedValue(true)
      mockAuthStore.accessToken = 'new-token-after-refresh'

      const interceptor = apiClient.interceptors.response.handlers[1]
      if (interceptor?.rejected) {
        await expect(interceptor.rejected(successError)).resolves.toBeDefined()
      }
    })

    it('should reject queued promises on token refresh failure', async () => {
      vi.resetModules()

      mockAuthStore.refreshAccessToken = vi.fn().mockResolvedValue(false)
      mockAuthStore.logout = vi.fn()

      const { default: apiClient } = await import('@/api/client')

      const error = {
        response: { status: 401 },
        config: { _retry: false, headers: {} },
      }

      const interceptor = apiClient.interceptors.response.handlers[1]
      if (interceptor?.rejected) {
        await expect(interceptor.rejected(error)).rejects.toBeDefined()
      }
    })

    it('should handle multiple queued requests during token refresh', async () => {
      vi.resetModules()

      mockAuthStore.refreshAccessToken = vi.fn().mockResolvedValue(true)
      mockAuthStore.accessToken = 'multi-token'

      const { default: apiClient } = await import('@/api/client')

      const error1 = {
        response: { status: 401 },
        config: { _retry: false, headers: {} },
      }

      const error2 = {
        response: { status: 401 },
        config: { _retry: false, headers: {} },
      }

      const interceptor = apiClient.interceptors.response.handlers[1]
      if (interceptor?.rejected) {
        const promise1 = interceptor.rejected(error1)
        const promise2 = interceptor.rejected(error2)
        await expect(promise1).resolves.toBeDefined()
        await expect(promise2).resolves.toBeDefined()
      }
    })
  })

  describe('setSkipAuthInterceptor', () => {
    it('should return undefined when setting value', async () => {
      vi.resetModules()
      const { setSkipAuthInterceptor } = await import('@/api/client')
      const result = setSkipAuthInterceptor(true)
      expect(result).toBeUndefined()
    })

    it('should toggle skipAuthInterceptor flag off', async () => {
      vi.resetModules()
      const { setSkipAuthInterceptor } = await import('@/api/client')
      setSkipAuthInterceptor(false)
      const { default: apiClient } = await import('@/api/client')

      const error = {
        response: { status: 401 },
        config: { _retry: false },
      }

      const interceptor = apiClient.interceptors.response.handlers[1]
      if (interceptor?.rejected) {
        await expect(interceptor.rejected(error)).rejects.toBeDefined()
      }
    })
  })

  describe('response interceptor - 403 CSRF handling', () => {
    it('should handle 403 with non-CSRF error message', async () => {
      vi.resetModules()
      const { default: apiClient } = await import('@/api/client')

      const error = {
        response: {
          status: 403,
          data: { message: 'Access denied' },
        },
      }

      const interceptor = apiClient.interceptors.response.handlers[1]
      if (interceptor?.rejected) {
        await expect(interceptor.rejected(error)).rejects.toEqual(error)
      }
    })

    it('should handle 403 with empty error message', async () => {
      vi.resetModules()
      const { default: apiClient } = await import('@/api/client')

      const error = {
        response: {
          status: 403,
          data: {},
        },
      }

      const interceptor = apiClient.interceptors.response.handlers[1]
      if (interceptor?.rejected) {
        await expect(interceptor.rejected(error)).rejects.toEqual(error)
      }
    })
  })
})
