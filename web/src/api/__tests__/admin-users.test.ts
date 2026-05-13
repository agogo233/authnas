import { describe, it, expect, vi, beforeEach } from 'vitest'
import { adminUsersApi } from '@/api/admin/users'

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))

describe('adminUsersApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('count', () => {
    it('should call GET /admin/users/count', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: { success: true, data: { total: 100 } },
      })

      const result = await adminUsersApi.count()

      expect(apiClient.get).toHaveBeenCalledWith('/admin/users/count')
      expect(result.data.data!.total).toBe(100)
    })
  })

  describe('list', () => {
    it('should call GET /admin/users without params', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: {
          success: true,
          data: [],
          total: 0,
          page: 1,
          pageSize: 10,
        },
      })

      const result = await adminUsersApi.list()

      expect(apiClient.get).toHaveBeenCalledWith('/admin/users')
      expect(result.data.data!).toEqual([])
    })

    it('should call GET /admin/users with pagination params', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: {
          success: true,
          data: [],
          total: 50,
          page: 2,
          pageSize: 20,
        },
      })

      const result = await adminUsersApi.list({ page: 2, pageSize: 20 })

      expect(apiClient.get).toHaveBeenCalledWith('/admin/users?page=2&pageSize=20')
      expect(result.data.page).toBe(2)
    })

    it('should call GET /admin/users with search param', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: {
          success: true,
          data: [],
          total: 5,
          page: 1,
          pageSize: 10,
        },
      })

      const result = await adminUsersApi.list({ search: 'test' })

      expect(apiClient.get).toHaveBeenCalledWith('/admin/users?search=test')
      expect(result.data.total).toBe(5)
    })

    it('should call GET /admin/users with all params', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: {
          success: true,
          data: [],
          total: 10,
          page: 3,
          pageSize: 25,
        },
      })

      await adminUsersApi.list({ page: 3, pageSize: 25, search: 'admin' })

      expect(apiClient.get).toHaveBeenCalledWith('/admin/users?page=3&pageSize=25&search=admin')
    })
  })

  describe('get', () => {
    it('should call GET /admin/users/:id', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: {
          success: true,
          data: {
            id: 'user-123',
            username: 'testuser',
            email: 'test@example.com',
            emailVerified: true,
            approved: true,
            mfaRequired: false,
            isAdmin: false,
            createdAt: '2024-01-01',
          },
        },
      })

      const result = await adminUsersApi.get('user-123')

      expect(apiClient.get).toHaveBeenCalledWith('/admin/users/user-123')
      expect(result.data.data!.username).toBe('testuser')
    })
  })

  describe('create', () => {
    it('should call POST /admin/users with all fields', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: {
          success: true,
          data: {
            id: 'new-user-123',
            username: 'newuser',
            email: 'new@example.com',
            emailVerified: false,
            approved: true,
            mfaRequired: false,
            isAdmin: false,
            createdAt: '2024-01-01',
          },
        },
      })

      const result = await adminUsersApi.create({
        email: 'new@example.com',
        username: 'newuser',
        password: 'Password123!',
        name: 'New User',
        isAdmin: false,
        approved: true,
        mfaRequired: false,
      })

      expect(apiClient.post).toHaveBeenCalledWith('/admin/users', {
        email: 'new@example.com',
        username: 'newuser',
        password: 'Password123!',
        name: 'New User',
        isAdmin: false,
        approved: true,
        mfaRequired: false,
      })
      expect(result.data.data!.username).toBe('newuser')
    })

    it('should call POST /admin/users without optional fields', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: {
          success: true,
          data: {
            id: 'user-123',
            username: 'user',
            email: 'user@example.com',
            emailVerified: false,
            approved: false,
            mfaRequired: true,
            isAdmin: false,
            createdAt: '2024-01-01',
          },
        },
      })

      await adminUsersApi.create({
        email: 'user@example.com',
        username: 'user',
      })

      expect(apiClient.post).toHaveBeenCalledWith('/admin/users', {
        email: 'user@example.com',
        username: 'user',
      })
    })
  })

  describe('update', () => {
    it('should call PUT /admin/users/:id', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.put).mockResolvedValue({
        data: { success: true },
      })

      const result = await adminUsersApi.update('user-123', {
        name: 'Updated Name',
        isAdmin: true,
      })

      expect(apiClient.put).toHaveBeenCalledWith('/admin/users/user-123', {
        name: 'Updated Name',
        isAdmin: true,
      })
      expect(result.data.success).toBe(true)
    })

    it('should update user with partial data', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.put).mockResolvedValue({
        data: { success: true },
      })

      await adminUsersApi.update('user-123', { approved: false })

      expect(apiClient.put).toHaveBeenCalledWith('/admin/users/user-123', { approved: false })
    })
  })

  describe('delete', () => {
    it('should call DELETE /admin/users/:id', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.delete).mockResolvedValue({
        data: { success: true },
      })

      const result = await adminUsersApi.delete('user-123')

      expect(apiClient.delete).toHaveBeenCalledWith('/admin/users/user-123')
      expect(result.data.success).toBe(true)
    })
  })

  describe('approve', () => {
    it('should call POST /admin/users/:id/approve', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: { success: true },
      })

      const result = await adminUsersApi.approve('user-123', { approved: true })

      expect(apiClient.post).toHaveBeenCalledWith('/admin/users/user-123/approve', {
        approved: true,
      })
      expect(result.data.success).toBe(true)
    })

    it('should disapprove user', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: { success: true },
      })

      await adminUsersApi.approve('user-123', { approved: false })

      expect(apiClient.post).toHaveBeenCalledWith('/admin/users/user-123/approve', {
        approved: false,
      })
    })
  })

  describe('resetPassword', () => {
    it('should call POST /admin/users/:id/reset-password', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: { success: true },
      })

      const result = await adminUsersApi.resetPassword('user-123', {
        newPassword: 'NewPassword123!',
      })

      expect(apiClient.post).toHaveBeenCalledWith('/admin/users/user-123/reset-password', {
        newPassword: 'NewPassword123!',
      })
      expect(result.data.success).toBe(true)
    })
  })
})
