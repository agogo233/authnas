import { describe, it, expect, vi, beforeEach } from 'vitest'
import { adminProxyAuthApi } from '@/api/admin/proxyauth'

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))

describe('adminProxyAuthApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('list', () => {
    it('should call GET /admin/proxyauth', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: {
          success: true,
          data: [
            {
              id: 'proxy-1',
              name: 'Auth Proxy 1',
              proxyUrl: 'https://proxy1.example.com',
              headerName: 'X-User-ID',
              enabled: true,
              createdAt: '2024-01-01',
            },
            {
              id: 'proxy-2',
              name: 'Auth Proxy 2',
              proxyUrl: 'https://proxy2.example.com',
              headerName: 'X-Auth-User',
              enabled: false,
              createdAt: '2024-01-02',
            },
          ],
          total: 2,
        },
      })

      const result = await adminProxyAuthApi.list()

      expect(apiClient.get).toHaveBeenCalledWith('/admin/proxyauth')
      expect(result.data.data!.length).toBe(2)
      expect(result.data.data![0].name).toBe('Auth Proxy 1')
      expect(result.data.data![0].enabled).toBe(true)
    })

    it('should return empty list when no proxy auth configs', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: { success: true, data: [], total: 0 },
      })

      const result = await adminProxyAuthApi.list()

      expect(result.data.data!).toEqual([])
      expect(result.data.total).toBe(0)
    })
  })

  describe('get', () => {
    it('should call GET /admin/proxyauth/:id', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: {
          success: true,
          data: {
            id: 'proxy-123',
            name: 'Test Proxy',
            proxyUrl: 'https://test-proxy.example.com',
            headerName: 'X-User-Id',
            enabled: true,
            createdAt: '2024-01-01',
          },
        },
      })

      const result = await adminProxyAuthApi.get('proxy-123')

      expect(apiClient.get).toHaveBeenCalledWith('/admin/proxyauth/proxy-123')
      expect(result.data.data!.name).toBe('Test Proxy')
      expect(result.data.data!.proxyUrl).toBe('https://test-proxy.example.com')
    })
  })

  describe('create', () => {
    it('should call POST /admin/proxyauth with all fields', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: {
          success: true,
          data: {
            id: 'new-proxy-123',
            name: 'New Proxy',
            proxyUrl: 'https://new-proxy.example.com',
            headerName: 'X-New-User',
            scopes: 'openid profile',
            groupId: 'group-123',
            enabled: true,
            createdAt: '2024-01-01',
          },
        },
      })

      const result = await adminProxyAuthApi.create({
        name: 'New Proxy',
        proxyUrl: 'https://new-proxy.example.com',
        headerName: 'X-New-User',
        scopes: 'openid profile',
        groupId: 'group-123',
        enabled: true,
      })

      expect(apiClient.post).toHaveBeenCalledWith('/admin/proxyauth', {
        name: 'New Proxy',
        proxyUrl: 'https://new-proxy.example.com',
        headerName: 'X-New-User',
        scopes: 'openid profile',
        groupId: 'group-123',
        enabled: true,
      })
      expect(result.data.data!.id).toBe('new-proxy-123')
    })

    it('should call POST /admin/proxyauth with only required fields', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: {
          success: true,
          data: {
            id: 'proxy-min',
            name: 'Minimal Proxy',
            proxyUrl: 'https://min-proxy.example.com',
            headerName: 'X-User',
            enabled: false,
            createdAt: '2024-01-01',
          },
        },
      })

      const result = await adminProxyAuthApi.create({
        name: 'Minimal Proxy',
        proxyUrl: 'https://min-proxy.example.com',
        headerName: 'X-User',
      })

      expect(apiClient.post).toHaveBeenCalledWith('/admin/proxyauth', {
        name: 'Minimal Proxy',
        proxyUrl: 'https://min-proxy.example.com',
        headerName: 'X-User',
      })
      expect(result.data.data!.enabled).toBe(false)
    })

    it('should create proxy auth disabled by default', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: {
          success: true,
          data: {
            id: 'proxy-def',
            name: 'Default Disabled Proxy',
            proxyUrl: 'https://def.example.com',
            headerName: 'X-Def',
            enabled: false,
            createdAt: '2024-01-01',
          },
        },
      })

      await adminProxyAuthApi.create({
        name: 'Default Disabled Proxy',
        proxyUrl: 'https://def.example.com',
        headerName: 'X-Def',
      })

      expect(apiClient.post).toHaveBeenCalledWith('/admin/proxyauth', {
        name: 'Default Disabled Proxy',
        proxyUrl: 'https://def.example.com',
        headerName: 'X-Def',
      })
    })
  })

  describe('update', () => {
    it('should call PUT /admin/proxyauth/:id', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.put).mockResolvedValue({
        data: { success: true },
      })

      const result = await adminProxyAuthApi.update('proxy-123', {
        name: 'Updated Proxy Name',
        enabled: true,
      })

      expect(apiClient.put).toHaveBeenCalledWith('/admin/proxyauth/proxy-123', {
        name: 'Updated Proxy Name',
        enabled: true,
      })
      expect(result.data.success).toBe(true)
    })

    it('should update proxyUrl', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.put).mockResolvedValue({
        data: { success: true },
      })

      await adminProxyAuthApi.update('proxy-123', {
        proxyUrl: 'https://updated-proxy.example.com',
      })

      expect(apiClient.put).toHaveBeenCalledWith('/admin/proxyauth/proxy-123', {
        proxyUrl: 'https://updated-proxy.example.com',
      })
    })

    it('should update headerName', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.put).mockResolvedValue({
        data: { success: true },
      })

      await adminProxyAuthApi.update('proxy-123', {
        headerName: 'X-Updated-Header',
      })

      expect(apiClient.put).toHaveBeenCalledWith('/admin/proxyauth/proxy-123', {
        headerName: 'X-Updated-Header',
      })
    })

    it('should update multiple fields at once', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.put).mockResolvedValue({
        data: { success: true },
      })

      await adminProxyAuthApi.update('proxy-123', {
        name: 'New Name',
        proxyUrl: 'https://new.example.com',
        headerName: 'X-New',
        scopes: 'openid email',
        enabled: true,
      })

      expect(apiClient.put).toHaveBeenCalledWith('/admin/proxyauth/proxy-123', {
        name: 'New Name',
        proxyUrl: 'https://new.example.com',
        headerName: 'X-New',
        scopes: 'openid email',
        enabled: true,
      })
    })
  })

  describe('delete', () => {
    it('should call DELETE /admin/proxyauth/:id', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.delete).mockResolvedValue({
        data: { success: true },
      })

      const result = await adminProxyAuthApi.delete('proxy-123')

      expect(apiClient.delete).toHaveBeenCalledWith('/admin/proxyauth/proxy-123')
      expect(result.data.success).toBe(true)
    })
  })
})
