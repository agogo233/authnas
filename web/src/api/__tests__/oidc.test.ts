import { describe, it, expect, vi, beforeEach } from 'vitest'
import { oidcApi } from '../oidc'

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    delete: vi.fn(),
  },
}))

describe('oidcApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getInteraction', () => {
    it('should call GET /oidc/interaction/:uid', async () => {
      const { default: apiClient } = await import('@/api/client')
      const mockInteraction = {
        uid: 'test-uid',
        client: { clientId: 'client-1', name: 'Test App', logoUri: 'https://example.com/logo.png' },
        scopes: ['openid', 'profile'],
        claims: { sub: 'user123' },
        requestUrl: 'https://example.com/auth',
      }
      vi.mocked(apiClient.get).mockResolvedValue({
        data: { success: true, data: mockInteraction },
      })

      const result = await oidcApi.getInteraction('test-uid')

      expect(apiClient.get).toHaveBeenCalledWith('/oidc/interaction/test-uid')
      expect(result.data.data.uid).toBe('test-uid')
      expect(result.data.data.client.name).toBe('Test App')
    })

    it('should return interaction with logoUri', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: {
          success: true,
          data: {
            uid: 'uid-1',
            client: { clientId: 'client-1', name: 'App', logoUri: 'https://example.com/logo.png' },
            scopes: ['openid'],
            claims: {},
            requestUrl: 'https://example.com',
          },
        },
      })

      const result = await oidcApi.getInteraction('uid-1')

      expect(result.data.data.client.logoUri).toBe('https://example.com/logo.png')
    })
  })

  describe('confirmInteraction', () => {
    it('should call POST /oidc/interaction/:uid/confirm', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: { success: true, data: { redirectTo: 'https://example.com/callback' } },
      })

      const result = await oidcApi.confirmInteraction('test-uid')

      expect(apiClient.post).toHaveBeenCalledWith('/oidc/interaction/test-uid/confirm')
      expect(result.data.data.redirectTo).toBe('https://example.com/callback')
    })

    it('should return redirect URL after confirmation', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: { success: true, data: { redirectTo: 'https://app.example.com/dashboard' } },
      })

      const result = await oidcApi.confirmInteraction('uid-123')

      expect(result.data.data.redirectTo).toBe('https://app.example.com/dashboard')
    })
  })

  describe('cancelInteraction', () => {
    it('should call DELETE /oidc/interaction/:uid/cancel', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.delete).mockResolvedValue({
        data: { success: true, data: { redirectTo: 'https://example.com/cancel' } },
      })

      const result = await oidcApi.cancelInteraction('test-uid')

      expect(apiClient.delete).toHaveBeenCalledWith('/oidc/interaction/test-uid/cancel')
      expect(result.data.data.redirectTo).toBe('https://example.com/cancel')
    })
  })

  describe('refreshToken', () => {
    it('should call POST /oidc/token with refresh_token grant', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: {
          success: true,
          data: {
            accessToken: 'new-access-token',
            refreshToken: 'new-refresh-token',
            expiresIn: 3600,
            expiresAt: '2024-12-31T23:59:59Z',
          },
        },
      })

      const result = await oidcApi.refreshToken('refresh-token-123')

      expect(apiClient.post).toHaveBeenCalledWith('/oidc/token', {
        grant_type: 'refresh_token',
        refresh_token: 'refresh-token-123',
      })
      expect(result.data.data.accessToken).toBe('new-access-token')
      expect(result.data.data.refreshToken).toBe('new-refresh-token')
    })

    it('should include expiresIn in response', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: {
          success: true,
          data: { accessToken: 'token', refreshToken: 'refresh', expiresIn: 7200 },
        },
      })

      const result = await oidcApi.refreshToken('refresh')

      expect(result.data.data.expiresIn).toBe(7200)
    })
  })

  describe('logout', () => {
    it('should call GET /oidc/logout without parameters', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: { success: true, data: { redirectTo: '/login' } },
      })

      const result = await oidcApi.logout()

      expect(apiClient.get).toHaveBeenCalledWith('/oidc/logout')
      expect(result.data.data.redirectTo).toBe('/login')
    })

    it('should call GET /oidc/logout with idTokenHint', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: { success: true, data: { redirectTo: '/login' } },
      })

      await oidcApi.logout('id-token-hint')

      expect(apiClient.get).toHaveBeenCalledWith('/oidc/logout?id_token_hint=id-token-hint')
    })

    it('should call GET /oidc/logout with all parameters', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: { success: true, data: { redirectTo: '/login' } },
      })

      await oidcApi.logout('id-token', 'https://example.com/logout', 'state-123')

      expect(apiClient.get).toHaveBeenCalledWith(
        '/oidc/logout?id_token_hint=id-token&post_logout_redirect_uri=https%3A%2F%2Fexample.com%2Flogout&state=state-123'
      )
    })

    it('should handle logout with only postLogoutRedirectURI', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: { success: true, data: { redirectTo: '/login' } },
      })

      await oidcApi.logout(undefined, 'https://example.com/callback')

      expect(apiClient.get).toHaveBeenCalledWith(
        '/oidc/logout?post_logout_redirect_uri=https%3A%2F%2Fexample.com%2Fcallback'
      )
    })

    it('should handle logout with only state', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: { success: true, data: { redirectTo: '/login' } },
      })

      await oidcApi.logout(undefined, undefined, 'state-value')

      expect(apiClient.get).toHaveBeenCalledWith('/oidc/logout?state=state-value')
    })
  })
})
