import { describe, it, expect, vi, beforeEach } from 'vitest'
import { adminGroupsApi } from '@/api/admin/groups'

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))

describe('adminGroupsApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('list', () => {
    it('should call GET /admin/groups', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: {
          success: true,
          data: [
            {
              id: 'group-1',
              name: 'Admins',
              description: 'Administrator group',
              createdAt: '2024-01-01',
            },
            { id: 'group-2', name: 'Users', description: 'Regular users', createdAt: '2024-01-02' },
          ],
          total: 2,
        },
      })

      const result = await adminGroupsApi.list()

      expect(apiClient.get).toHaveBeenCalledWith('/admin/groups')
      expect(result.data.data!.length).toBe(2)
      expect(result.data.data![0].name).toBe('Admins')
    })

    it('should return empty list when no groups', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: { success: true, data: [], total: 0 },
      })

      const result = await adminGroupsApi.list()

      expect(result.data.data!).toEqual([])
      expect(result.data.total).toBe(0)
    })
  })

  describe('get', () => {
    it('should call GET /admin/groups/:id', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: {
          success: true,
          data: {
            id: 'group-123',
            name: 'Test Group',
            description: 'A test group',
            createdAt: '2024-01-01',
          },
        },
      })

      const result = await adminGroupsApi.get('group-123')

      expect(apiClient.get).toHaveBeenCalledWith('/admin/groups/group-123')
      expect(result.data.data!.name).toBe('Test Group')
    })

    it('should return group without description', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: {
          success: true,
          data: {
            id: 'group-456',
            name: 'Simple Group',
            createdAt: '2024-01-01',
          },
        },
      })

      const result = await adminGroupsApi.get('group-456')

      expect(result.data.data!.description).toBeUndefined()
    })
  })

  describe('create', () => {
    it('should call POST /admin/groups with name and description', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: {
          success: true,
          data: {
            id: 'new-group-123',
            name: 'New Group',
            description: 'A new group',
            createdAt: '2024-01-01',
          },
        },
      })

      const result = await adminGroupsApi.create({
        name: 'New Group',
        description: 'A new group',
      })

      expect(apiClient.post).toHaveBeenCalledWith('/admin/groups', {
        name: 'New Group',
        description: 'A new group',
      })
      expect(result.data.data!.id).toBe('new-group-123')
    })

    it('should call POST /admin/groups with only name', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: {
          success: true,
          data: {
            id: 'group-789',
            name: 'Name Only Group',
            createdAt: '2024-01-01',
          },
        },
      })

      const result = await adminGroupsApi.create({ name: 'Name Only Group' })

      expect(apiClient.post).toHaveBeenCalledWith('/admin/groups', { name: 'Name Only Group' })
      expect(result.data.data!.name).toBe('Name Only Group')
    })
  })

  describe('update', () => {
    it('should call PUT /admin/groups/:id with name and description', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.put).mockResolvedValue({
        data: {
          success: true,
          data: {
            id: 'group-123',
            name: 'Updated Group',
            description: 'Updated description',
            createdAt: '2024-01-01',
          },
        },
      })

      const result = await adminGroupsApi.update('group-123', {
        name: 'Updated Group',
        description: 'Updated description',
      })

      expect(apiClient.put).toHaveBeenCalledWith('/admin/groups/group-123', {
        name: 'Updated Group',
        description: 'Updated description',
      })
      expect(result.data.data!.name).toBe('Updated Group')
    })

    it('should update only name', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.put).mockResolvedValue({
        data: {
          success: true,
          data: {
            id: 'group-123',
            name: 'Name Updated',
            createdAt: '2024-01-01',
          },
        },
      })

      await adminGroupsApi.update('group-123', { name: 'Name Updated' })

      expect(apiClient.put).toHaveBeenCalledWith('/admin/groups/group-123', {
        name: 'Name Updated',
      })
    })

    it('should update only description', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.put).mockResolvedValue({
        data: {
          success: true,
          data: {
            id: 'group-123',
            name: 'Original Name',
            description: 'Description Updated',
            createdAt: '2024-01-01',
          },
        },
      })

      await adminGroupsApi.update('group-123', { description: 'Description Updated' })

      expect(apiClient.put).toHaveBeenCalledWith('/admin/groups/group-123', {
        description: 'Description Updated',
      })
    })
  })

  describe('delete', () => {
    it('should call DELETE /admin/groups/:id', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.delete).mockResolvedValue({
        data: { success: true },
      })

      const result = await adminGroupsApi.delete('group-123')

      expect(apiClient.delete).toHaveBeenCalledWith('/admin/groups/group-123')
      expect(result.data.success).toBe(true)
    })
  })
})
