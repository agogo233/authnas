import { describe, it, expect, vi, beforeEach } from 'vitest'
import { adminClientsApi, type CreateClientRequest, type UpdateClientRequest } from '../clients'
import apiClient from '@/api/client'

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))

describe('adminClientsApi', () => {
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
          data: [{ id: '1', clientId: 'client1', name: 'Test Client' }],
          total: 1,
        },
      })

      const result = await adminClientsApi.list()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/admin/clients')
      expect(result.data?.data).toHaveLength(1)
      expect(result.data?.total).toBe(1)
    })

    it('should return list response with items and total', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: {
          success: true,
          data: [
            { id: '1', clientId: 'client1', name: 'Client 1' },
            { id: '2', clientId: 'client2', name: 'Client 2' },
          ],
          total: 2,
        },
      })

      const result = await adminClientsApi.list()

      expect(result.data?.data).toHaveLength(2)
      expect(result.data?.total).toBe(2)
    })
  })

  describe('get', () => {
    it('should call get with correct id', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: {
          success: true,
          data: { id: '1', clientId: 'client1', name: 'Test Client' },
        },
      })

      const result = await adminClientsApi.get('1')

      expect(mockedApiClient.get).toHaveBeenCalledWith('/admin/clients/1')
      expect(result.data?.data).toEqual({ id: '1', clientId: 'client1', name: 'Test Client' })
    })
  })

  describe('create', () => {
    it('should call post with create request data', async () => {
      const createData: CreateClientRequest = {
        clientId: 'new-client',
        name: 'New Client',
        logoUri: 'https://example.com/logo.png',
        redirectUris: 'https://example.com/callback',
      }
      mockedApiClient.post.mockResolvedValue({
        data: {
          success: true,
          data: { id: '2', ...createData, createdAt: '2024-01-01T00:00:00Z' },
        },
      })

      const result = await adminClientsApi.create(createData)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/admin/clients', createData)
      expect(result.data?.data).toMatchObject({
        id: '2',
        clientId: 'new-client',
        name: 'New Client',
      })
    })

    it('should send minimal create request with only required fields', async () => {
      const createData: CreateClientRequest = {
        clientId: 'minimal-client',
        name: 'Minimal Client',
      }
      mockedApiClient.post.mockResolvedValue({
        data: {
          success: true,
          data: { id: '3', ...createData, createdAt: '2024-01-01T00:00:00Z' },
        },
      })

      await adminClientsApi.create(createData)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/admin/clients', createData)
    })
  })

  describe('update', () => {
    it('should call put with id and update data', async () => {
      const updateData: UpdateClientRequest = {
        name: 'Updated Client Name',
        logoUri: 'https://new-example.com/logo.png',
      }
      mockedApiClient.put.mockResolvedValue({
        data: { success: true },
      })

      await adminClientsApi.update('1', updateData)

      expect(mockedApiClient.put).toHaveBeenCalledWith('/admin/clients/1', updateData)
    })

    it('should allow partial update with only changed fields', async () => {
      const updateData: UpdateClientRequest = {
        name: 'Only Name Updated',
      }
      mockedApiClient.put.mockResolvedValue({
        data: { success: true },
      })

      await adminClientsApi.update('1', updateData)

      expect(mockedApiClient.put).toHaveBeenCalledWith('/admin/clients/1', updateData)
    })
  })

  describe('delete', () => {
    it('should call delete with correct id', async () => {
      mockedApiClient.delete.mockResolvedValue({
        data: { success: true },
      })

      await adminClientsApi.delete('1')

      expect(mockedApiClient.delete).toHaveBeenCalledWith('/admin/clients/1')
    })

    it('should handle delete response', async () => {
      mockedApiClient.delete.mockResolvedValue({
        data: { success: true },
      })

      const result = await adminClientsApi.delete('1')

      expect(result.data?.success).toBe(true)
    })
  })
})
