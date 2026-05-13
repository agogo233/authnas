import { describe, it, expect, vi, beforeEach } from 'vitest'
import { adminGroupsApi, type CreateGroupRequest, type UpdateGroupRequest } from '../groups'
import apiClient from '@/api/client'

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))

describe('adminGroupsApi', () => {
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
          data: [{ id: '1', name: 'Admin Group', description: 'Administrators' }],
          total: 1,
        },
      })

      const result = await adminGroupsApi.list()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/admin/groups')
      expect(result.data?.data).toHaveLength(1)
      expect(result.data?.total).toBe(1)
    })

    it('should return empty list when no groups exist', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: {
          success: true,
          data: [],
          total: 0,
        },
      })

      const result = await adminGroupsApi.list()

      expect(result.data?.data).toHaveLength(0)
      expect(result.data?.total).toBe(0)
    })
  })

  describe('get', () => {
    it('should call get with correct id', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: {
          success: true,
          data: { id: '1', name: 'Admin Group', description: 'Administrators' },
        },
      })

      const result = await adminGroupsApi.get('1')

      expect(mockedApiClient.get).toHaveBeenCalledWith('/admin/groups/1')
      expect(result.data?.data).toEqual({
        id: '1',
        name: 'Admin Group',
        description: 'Administrators',
      })
    })
  })

  describe('create', () => {
    it('should call post with create request data', async () => {
      const createData: CreateGroupRequest = {
        name: 'New Group',
        description: 'A new group description',
      }
      mockedApiClient.post.mockResolvedValue({
        data: {
          success: true,
          data: { id: '2', ...createData, createdAt: '2024-01-01T00:00:00Z' },
        },
      })

      const result = await adminGroupsApi.create(createData)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/admin/groups', createData)
      expect(result.data?.data).toMatchObject({ id: '2', name: 'New Group' })
    })

    it('should send minimal create request with only required fields', async () => {
      const createData: CreateGroupRequest = {
        name: 'Minimal Group',
      }
      mockedApiClient.post.mockResolvedValue({
        data: {
          success: true,
          data: { id: '3', ...createData, createdAt: '2024-01-01T00:00:00Z' },
        },
      })

      await adminGroupsApi.create(createData)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/admin/groups', createData)
    })
  })

  describe('update', () => {
    it('should call put with id and update data', async () => {
      const updateData: UpdateGroupRequest = {
        name: 'Updated Group Name',
        description: 'Updated description',
      }
      mockedApiClient.put.mockResolvedValue({
        data: { success: true, data: { id: '1', ...updateData } },
      })

      await adminGroupsApi.update('1', updateData)

      expect(mockedApiClient.put).toHaveBeenCalledWith('/admin/groups/1', updateData)
    })

    it('should allow partial update with only name changed', async () => {
      const updateData: UpdateGroupRequest = {
        name: 'Only Name Updated',
      }
      mockedApiClient.put.mockResolvedValue({
        data: { success: true, data: { id: '1', ...updateData } },
      })

      await adminGroupsApi.update('1', updateData)

      expect(mockedApiClient.put).toHaveBeenCalledWith('/admin/groups/1', updateData)
    })

    it('should allow partial update with only description changed', async () => {
      const updateData: UpdateGroupRequest = {
        description: 'Only description updated',
      }
      mockedApiClient.put.mockResolvedValue({
        data: { success: true, data: { id: '1', ...updateData } },
      })

      await adminGroupsApi.update('1', updateData)

      expect(mockedApiClient.put).toHaveBeenCalledWith('/admin/groups/1', updateData)
    })
  })

  describe('delete', () => {
    it('should call delete with correct id', async () => {
      mockedApiClient.delete.mockResolvedValue({
        data: { success: true },
      })

      await adminGroupsApi.delete('1')

      expect(mockedApiClient.delete).toHaveBeenCalledWith('/admin/groups/1')
    })

    it('should handle delete response', async () => {
      mockedApiClient.delete.mockResolvedValue({
        data: { success: true },
      })

      const result = await adminGroupsApi.delete('1')

      expect(result.data?.success).toBe(true)
    })
  })
})
