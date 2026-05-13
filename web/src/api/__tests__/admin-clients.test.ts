import { describe, it, expect, vi, beforeEach } from 'vitest'
import { adminClientsApi } from '@/api/admin/clients'

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))

describe('adminClientsApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('list', () => {
    it('should call GET /admin/clients', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: {
          success: true,
          data: [
            {
              id: 'client-1',
              clientId: 'app-1',
              name: 'My App',
              logoUri: 'https://example.com/logo.png',
              createdAt: '2024-01-01',
            },
            { id: 'client-2', clientId: 'app-2', name: 'Another App', createdAt: '2024-01-02' },
          ],
          total: 2,
        },
      })

      const result = await adminClientsApi.list()

      expect(apiClient.get).toHaveBeenCalledWith('/admin/clients')
      expect(result.data.data!.length).toBe(2)
      expect(result.data.data![0].name).toBe('My App')
    })

    it('should return empty list when no clients', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: { success: true, data: [], total: 0 },
      })

      const result = await adminClientsApi.list()

      expect(result.data.data!).toEqual([])
      expect(result.data.total).toBe(0)
    })
  })

  describe('get', () => {
    it('should call GET /admin/clients/:id', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: {
          success: true,
          data: {
            id: 'client-123',
            clientId: 'my-client-id',
            name: 'My OAuth Client',
            logoUri: 'https://example.com/logo.png',
            createdAt: '2024-01-01',
          },
        },
      })

      const result = await adminClientsApi.get('client-123')

      expect(apiClient.get).toHaveBeenCalledWith('/admin/clients/client-123')
      expect(result.data.data!.clientId).toBe('my-client-id')
    })
  })

  describe('create', () => {
    it('should call POST /admin/clients with all fields', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: {
          success: true,
          data: {
            id: 'new-client-123',
            clientId: 'new-client-id',
            name: 'New Client',
            logoUri: 'https://example.com/new-logo.png',
            createdAt: '2024-01-01',
          },
        },
      })

      const result = await adminClientsApi.create({
        clientId: 'new-client-id',
        name: 'New Client',
        logoUri: 'https://example.com/new-logo.png',
        redirectUris: 'https://example.com/callback',
        postLogoutRedirectUris: 'https://example.com/logout',
        grantTypes: 'authorization_code',
        responseTypes: 'code',
        scopes: 'openid profile',
      })

      expect(apiClient.post).toHaveBeenCalledWith('/admin/clients', {
        clientId: 'new-client-id',
        name: 'New Client',
        logoUri: 'https://example.com/new-logo.png',
        redirectUris: 'https://example.com/callback',
        postLogoutRedirectUris: 'https://example.com/logout',
        grantTypes: 'authorization_code',
        responseTypes: 'code',
        scopes: 'openid profile',
      })
      expect(result.data.data!.id).toBe('new-client-123')
    })

    it('should call POST /admin/clients with only required fields', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: {
          success: true,
          data: {
            id: 'client-456',
            clientId: 'minimal-client',
            name: 'Minimal Client',
            createdAt: '2024-01-01',
          },
        },
      })

      const result = await adminClientsApi.create({
        clientId: 'minimal-client',
        name: 'Minimal Client',
      })

      expect(apiClient.post).toHaveBeenCalledWith('/admin/clients', {
        clientId: 'minimal-client',
        name: 'Minimal Client',
      })
      expect(result.data.data!.name).toBe('Minimal Client')
    })
  })

  describe('update', () => {
    it('should call PUT /admin/clients/:id', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.put).mockResolvedValue({
        data: { success: true },
      })

      const result = await adminClientsApi.update('client-123', {
        name: 'Updated Client Name',
        logoUri: 'https://example.com/updated-logo.png',
      })

      expect(apiClient.put).toHaveBeenCalledWith('/admin/clients/client-123', {
        name: 'Updated Client Name',
        logoUri: 'https://example.com/updated-logo.png',
      })
      expect(result.data.success).toBe(true)
    })

    it('should update redirectUris', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.put).mockResolvedValue({
        data: { success: true },
      })

      await adminClientsApi.update('client-123', {
        redirectUris: 'https://new-callback.com/callback',
      })

      expect(apiClient.put).toHaveBeenCalledWith('/admin/clients/client-123', {
        redirectUris: 'https://new-callback.com/callback',
      })
    })

    it('should update multiple fields at once', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.put).mockResolvedValue({
        data: { success: true },
      })

      await adminClientsApi.update('client-123', {
        name: 'New Name',
        redirectUris: 'https://example.com/callback',
        scopes: 'openid profile email',
      })

      expect(apiClient.put).toHaveBeenCalledWith('/admin/clients/client-123', {
        name: 'New Name',
        redirectUris: 'https://example.com/callback',
        scopes: 'openid profile email',
      })
    })
  })

  describe('delete', () => {
    it('should call DELETE /admin/clients/:id', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.delete).mockResolvedValue({
        data: { success: true },
      })

      const result = await adminClientsApi.delete('client-123')

      expect(apiClient.delete).toHaveBeenCalledWith('/admin/clients/client-123')
      expect(result.data.success).toBe(true)
    })
  })
})
