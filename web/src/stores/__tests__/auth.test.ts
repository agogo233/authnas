import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useAuthStore } from '../auth'

vi.mock('@/router', () => ({
  default: {
    push: vi.fn(),
  },
}))

vi.mock('@/api/oidc', () => ({
  oidcApi: {
    logout: vi.fn().mockResolvedValue({ data: { data: { redirectTo: '/login' } } }),
  },
}))

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    interceptors: {
      request: { use: vi.fn(), eject: vi.fn() },
      response: { use: vi.fn(), eject: vi.fn() },
    },
  },
}))

describe('auth store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  describe('initial state', () => {
    it('should have null user and accessToken initially', () => {
      const authStore = useAuthStore()
      expect(authStore.user).toBeNull()
      expect(authStore.accessToken).toBeNull()
    })

    it('should not be authenticated initially', () => {
      const authStore = useAuthStore()
      expect(authStore.isAuthenticated).toBe(false)
    })

    it('should not be admin initially', () => {
      const authStore = useAuthStore()
      expect(authStore.isAdmin).toBe(false)
    })
  })

  describe('setTokens', () => {
    it('should set access token', () => {
      const authStore = useAuthStore()
      authStore.setTokens('test-access-token')
      expect(authStore.accessToken).toBe('test-access-token')
    })

    it('should set access token with expiresAt', () => {
      const authStore = useAuthStore()
      const expiresAt = new Date(Date.now() + 3600000).toISOString()
      authStore.setTokens('test-access-token', expiresAt)
      expect(authStore.accessToken).toBe('test-access-token')
      expect(authStore.tokenExpiresAt).toBe(expiresAt)
    })

    it('should set isAuthenticated to true after setting tokens and user', () => {
      const authStore = useAuthStore()
      authStore.setTokens('test-access-token')
      authStore.setUser({
        id: '123',
        username: 'test',
        email: 'test@test.com',
        emailVerified: true,
        approved: true,
        isAdmin: false,
        createdAt: '2024-01-01T00:00:00Z',
      })
      expect(authStore.isAuthenticated).toBe(true)
    })
  })

  describe('setUser', () => {
    it('should set user data', () => {
      const authStore = useAuthStore()
      const mockUser = {
        id: '123',
        username: 'testuser',
        email: 'test@example.com',
        emailVerified: true,
        approved: true,
        isAdmin: false,
        createdAt: '2024-01-01T00:00:00Z',
      }
      authStore.setUser(mockUser)
      expect(authStore.user).toEqual(mockUser)
    })

    it('should set isAdmin based on user data', () => {
      const authStore = useAuthStore()
      const adminUser = {
        id: '123',
        username: 'admin',
        email: 'admin@example.com',
        emailVerified: true,
        approved: true,
        isAdmin: true,
        createdAt: '2024-01-01T00:00:00Z',
      }
      authStore.setUser(adminUser)
      expect(authStore.isAdmin).toBe(true)
    })

    it('should handle isAdmin undefined', () => {
      const authStore = useAuthStore()
      authStore.setUser({
        id: '123',
        username: 'test',
        email: 'test@test.com',
        emailVerified: true,
        approved: true,
        createdAt: '2024-01-01T00:00:00Z',
      })
      expect(authStore.isAdmin).toBe(false)
    })
  })

  describe('logout', () => {
    it('should clear user and tokens', async () => {
      const authStore = useAuthStore()
      authStore.setTokens('test-access-token')
      authStore.setUser({
        id: '123',
        username: 'test',
        email: 'test@test.com',
        emailVerified: true,
        approved: true,
        isAdmin: false,
        createdAt: '2024-01-01T00:00:00Z',
      })
      await authStore.logout()
      expect(authStore.user).toBeNull()
      expect(authStore.accessToken).toBeNull()
    })

    it('should set isAuthenticated to false after logout', async () => {
      const authStore = useAuthStore()
      authStore.setTokens('test-access-token')
      await authStore.logout()
      expect(authStore.isAuthenticated).toBe(false)
    })

    it('should handle logout error gracefully', async () => {
      const { oidcApi } = await import('@/api/oidc')
      vi.mocked(oidcApi.logout).mockRejectedValueOnce(new Error('Logout failed'))
      const authStore = useAuthStore()
      authStore.setTokens('test-access-token')
      await authStore.logout()
      expect(authStore.user).toBeNull()
      expect(authStore.accessToken).toBeNull()
    })
  })

  describe('initFromStorage', () => {
    it('should call /api/auth/me and set user data', async () => {
      const apiClient = (await import('@/api/client')).default
      vi.mocked(apiClient.get).mockResolvedValueOnce({
        data: {
          data: {
            id: 'user-123',
            email: 'test@example.com',
            username: 'testuser',
            name: 'Test User',
            emailVerified: true,
            approved: true,
            isAdmin: false,
            mfaRequired: false,
            hasTotp: true,
            hasPasskeys: false,
            hasPassword: true,
            tokenVersion: 1,
            createdAt: '2024-01-01T00:00:00Z',
            updatedAt: null,
            expiresAt: null,
          },
        },
      })

      const authStore = useAuthStore()
      await authStore.initFromStorage()

      expect(apiClient.get).toHaveBeenCalledWith('/api/auth/me')
      expect(authStore.user?.id).toBe('user-123')
      expect(authStore.user?.username).toBe('testuser')
      expect(authStore.initialized).toBe(true)
    })

    it('should clear state when /api/auth/me fails', async () => {
      const apiClient = (await import('@/api/client')).default
      vi.mocked(apiClient.get).mockRejectedValueOnce(new Error('Unauthorized'))

      const authStore = useAuthStore()
      authStore.accessToken = 'old-token'
      await authStore.initFromStorage()

      expect(authStore.user).toBeNull()
      expect(authStore.accessToken).toBeNull()
      expect(authStore.initialized).toBe(true)
    })

    it('should not reinitialize if already initialized', async () => {
      const apiClient = (await import('@/api/client')).default
      vi.mocked(apiClient.get).mockResolvedValueOnce({
        data: {
          data: {
            id: 'user-123',
            email: null,
            username: 'testuser',
            name: null,
            emailVerified: false,
            approved: false,
            isAdmin: false,
            mfaRequired: false,
            hasTotp: false,
            hasPasskeys: false,
            hasPassword: true,
            tokenVersion: 0,
            createdAt: '2024-01-01T00:00:00Z',
            updatedAt: null,
            expiresAt: null,
          },
        },
      })

      const authStore = useAuthStore()
      await authStore.initFromStorage()

      vi.mocked(apiClient.get).mockClear()
      await authStore.initFromStorage()

      expect(apiClient.get).not.toHaveBeenCalled()
    })
  })

  describe('refreshAccessToken', () => {
    it('should call /api/auth/refresh and update tokens on success', async () => {
      const apiClient = (await import('@/api/client')).default
      vi.mocked(apiClient.post).mockResolvedValueOnce({
        data: {
          data: {
            accessToken: 'new-access-token',
            expiresAt: new Date(Date.now() + 3600000).toISOString(),
          },
        },
      })

      const authStore = useAuthStore()
      const result = await authStore.refreshAccessToken()

      expect(result).toBe(true)
      expect(authStore.accessToken).toBe('new-access-token')
      expect(apiClient.post).toHaveBeenCalledWith('/api/auth/refresh')
    })

    it('should call logout on refresh failure', async () => {
      const apiClient = (await import('@/api/client')).default
      vi.mocked(apiClient.post).mockRejectedValueOnce(new Error('Unauthorized'))

      const authStore = useAuthStore()
      authStore.accessToken = 'old-token'
      const result = await authStore.refreshAccessToken()

      expect(result).toBe(false)
      expect(authStore.accessToken).toBeNull()
    })

    it('should return false when response has no access token', async () => {
      const apiClient = (await import('@/api/client')).default
      vi.mocked(apiClient.post).mockResolvedValueOnce({
        data: { data: {} },
      })

      const authStore = useAuthStore()
      const result = await authStore.refreshAccessToken()

      expect(result).toBe(false)
    })
  })

  describe('cleanup', () => {
    it('should handle cleanup when no timeout exists', () => {
      const authStore = useAuthStore()
      authStore.cleanup()
      expect(authStore.initialized).toBe(false)
    })
  })

  describe('scheduleRefreshIfNeeded', () => {
    it('should clear existing timeout before scheduling new one', () => {
      const authStore = useAuthStore()
      authStore.setTokens('test-access', new Date(Date.now() + 60000).toISOString())
      authStore.setTokens('test-access-2', new Date(Date.now() + 120000).toISOString())
      expect(authStore.accessToken).toBe('test-access-2')
    })

    it('should not schedule refresh when expiresAtTimestamp is 0', () => {
      const authStore = useAuthStore()
      authStore.setTokens('test-access')
      expect(authStore.accessToken).toBe('test-access')
    })
  })

  describe('isExpired computed', () => {
    it('should return true when user is null', () => {
      const authStore = useAuthStore()
      expect(authStore.isExpired).toBe(true)
    })

    it('should use user.exp for expiration check', () => {
      const authStore = useAuthStore()
      authStore.setTokens('access')
      authStore.setUser({
        id: '123',
        username: 'test',
        email: 'test@test.com',
        emailVerified: true,
        approved: true,
        isAdmin: false,
        createdAt: '2024-01-01T00:00:00Z',
        exp: Math.floor(Date.now() / 1000) + 3600,
      } as any)
      expect(authStore.isExpired).toBe(false)
    })

    it('should be expired when user.exp is in the past', () => {
      const authStore = useAuthStore()
      authStore.setTokens('access')
      authStore.setUser({
        id: '123',
        username: 'test',
        email: 'test@test.com',
        emailVerified: true,
        approved: true,
        isAdmin: false,
        createdAt: '2024-01-01T00:00:00Z',
        exp: Math.floor(Date.now() / 1000) - 3600,
      } as any)
      expect(authStore.isExpired).toBe(true)
    })

    it('should return true when token is expired', () => {
      const authStore = useAuthStore()
      authStore.setTokens('access', '2020-01-01T00:00:00Z')
      authStore.setUser({
        id: '123',
        username: 'test',
        email: 'test@test.com',
        emailVerified: true,
        approved: true,
        isAdmin: false,
        createdAt: '2024-01-01T00:00:00Z',
      })
      expect(authStore.isExpired).toBe(true)
    })

    it('should return false when token is not expired with user set', () => {
      const authStore = useAuthStore()
      authStore.setTokens('access', new Date(Date.now() + 3600000).toISOString())
      authStore.setUser({
        id: '123',
        username: 'test',
        email: 'test@test.com',
        emailVerified: true,
        approved: true,
        isAdmin: false,
        createdAt: '2024-01-01T00:00:00Z',
      })
      expect(authStore.isExpired).toBe(false)
    })
  })
})
