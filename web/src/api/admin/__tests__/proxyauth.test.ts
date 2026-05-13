import { describe, it, expect, vi, beforeEach } from 'vitest'
import {
  adminProxyAuthApi,
  type CreateProxyAuthRequest,
  type UpdateProxyAuthRequest,
} from '../proxyauth'
import apiClient from '@/api/client'

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))

describe('adminProxyAuthApi', () => {
  const mockedApiClient = apiClient as unknown as {
    get: ReturnType<typeof vi.fn>
    post: ReturnType<typeof vi.fn>
    put: ReturnType<typeof vi.fn>
    delete: ReturnType<typeof vi.fn>
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('list', () => {
    it('should call get with correct path', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: {
          success: true,
          data: [
            { id: '1', name: 'Proxy 1', proxyUrl: 'https://proxy.example.com', enabled: true },
          ],
          total: 1,
        },
      })

      const result = await adminProxyAuthApi.list()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/admin/proxyauth')
      expect(result.data?.data).toHaveLength(1)
      expect(result.data?.total).toBe(1)
    })

    it('should return empty list when no proxy auth configs exist', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: {
          success: true,
          data: [],
          total: 0,
        },
      })

      const result = await adminProxyAuthApi.list()

      expect(result.data?.data).toHaveLength(0)
      expect(result.data?.total).toBe(0)
    })
  })

  describe('get', () => {
    it('should call get with correct id', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: {
          success: true,
          data: { id: '1', name: 'Proxy 1', proxyUrl: 'https://proxy.example.com', enabled: true },
        },
      })

      const result = await adminProxyAuthApi.get('1')

      expect(mockedApiClient.get).toHaveBeenCalledWith('/admin/proxyauth/1')
      expect(result.data?.data).toEqual({
        id: '1',
        name: 'Proxy 1',
        proxyUrl: 'https://proxy.example.com',
        enabled: true,
      })
    })
  })

  describe('create', () => {
    it('should call post with create request data', async () => {
      const createData: CreateProxyAuthRequest = {
        name: 'New Proxy',
        proxyUrl: 'https://new-proxy.example.com',
        headerName: 'X-User-ID',
        scopes: 'openid profile',
        groupId: 'group1',
        enabled: true,
      }
      mockedApiClient.post.mockResolvedValue({
        data: {
          success: true,
          data: { id: '2', ...createData, createdAt: '2024-01-01T00:00:00Z' },
        },
      })

      const result = await adminProxyAuthApi.create(createData)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/admin/proxyauth', createData)
      expect(result.data?.data).toMatchObject({ id: '2', name: 'New Proxy' })
    })

    it('should send minimal create request with only required fields', async () => {
      const createData: CreateProxyAuthRequest = {
        name: 'Minimal Proxy',
        proxyUrl: 'https://minimal.example.com',
        headerName: 'X-Auth',
      }
      mockedApiClient.post.mockResolvedValue({
        data: {
          success: true,
          data: { id: '3', ...createData, enabled: false, createdAt: '2024-01-01T00:00:00Z' },
        },
      })

      await adminProxyAuthApi.create(createData)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/admin/proxyauth', createData)
    })
  })

  describe('update', () => {
    it('should call put with id and update data', async () => {
      const updateData: UpdateProxyAuthRequest = {
        name: 'Updated Proxy Name',
        proxyUrl: 'https://updated-proxy.example.com',
      }
      mockedApiClient.put.mockResolvedValue({
        data: { success: true },
      })

      await adminProxyAuthApi.update('1', updateData)

      expect(mockedApiClient.put).toHaveBeenCalledWith('/admin/proxyauth/1', updateData)
    })

    it('should allow partial update with only enabled flag changed', async () => {
      const updateData: UpdateProxyAuthRequest = {
        enabled: false,
      }
      mockedApiClient.put.mockResolvedValue({
        data: { success: true },
      })

      await adminProxyAuthApi.update('1', updateData)

      expect(mockedApiClient.put).toHaveBeenCalledWith('/admin/proxyauth/1', updateData)
    })

    it('should allow update of headerName', async () => {
      const updateData: UpdateProxyAuthRequest = {
        headerName: 'X-New-Header',
      }
      mockedApiClient.put.mockResolvedValue({
        data: { success: true },
      })

      await adminProxyAuthApi.update('1', updateData)

      expect(mockedApiClient.put).toHaveBeenCalledWith('/admin/proxyauth/1', updateData)
    })
  })

  describe('delete', () => {
    it('should call delete with correct id', async () => {
      mockedApiClient.delete.mockResolvedValue({
        data: { success: true },
      })

      await adminProxyAuthApi.delete('1')

      expect(mockedApiClient.delete).toHaveBeenCalledWith('/admin/proxyauth/1')
    })
  })
})
